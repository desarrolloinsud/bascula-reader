package app

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"bascula-connector/internal/config"
	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"
	"bascula-connector/internal/scale"
	"bascula-connector/internal/server"
	"bascula-connector/internal/status"
)

// Status es un snapshot del estado del conector en un momento dado.
type Status struct {
	HTTPRunning     bool
	COMConnected    bool
	COMReconnecting bool
	COMError        string
	ActivePort      string
	SerialPort      string
	LastWeight      string
	LastReadAt      string
}

// App encapsula el ciclo de vida de todos los servicios del conector.
// La GUI (y el CLI) la usan para arrancar/detener sin conocer los internos.
type App struct {
	Cfg         config.Config
	scale       domain.Scale
	httpSrv     *server.HTTPServer
	statusSndr  *status.StatusSender
	log         *logger.Logger
	httpRunning bool
	comRunning  bool
	mu          sync.RWMutex
	subs        []chan Status
	subsMu      sync.Mutex
}

// New crea un App con la configuración dada. No inicia ningún servicio.
func New(cfg config.Config, log *logger.Logger) *App {
	return &App{
		Cfg: cfg,
		log: log,
	}
}

// StartAll inicia COM y HTTP. Idempotente: no hace nada si ya corren.
func (a *App) StartAll() error {
	a.StartCOM()
	return a.StartHTTP()
}

// StopAll detiene HTTP y COM en orden.
func (a *App) StopAll() {
	a.StopHTTP()
	a.StopCOM()
}

// StartHTTP inicia el servidor HTTP en una goroutine.
func (a *App) StartHTTP() error {
	a.mu.Lock()
	if a.httpRunning {
		a.mu.Unlock()
		return nil
	}
	if a.scale == nil {
		a.mu.Unlock()
		return fmt.Errorf("COM no iniciado, arrancá primero StartCOM")
	}

	a.statusSndr = status.NewStatusSender(a.Cfg.StatusEndpointURL, a.scale, a.Cfg)
	a.httpSrv = server.NewHTTPServer(
		a.scale,
		a.Cfg.ScaleID,
		a.Cfg.ServerPort,
		a.Cfg.AllowedOrigin,
		a.Cfg.SerialPort,
		a.Cfg.BaudRate,
		a.Cfg.UseMock,
		a.statusSndr,
	)
	a.httpRunning = true
	a.mu.Unlock()

	go a.statusSndr.StartPeriodicSender()

	go func() {
		err := a.httpSrv.Start()
		a.mu.Lock()
		a.httpRunning = false
		a.mu.Unlock()
		if err != nil && err != http.ErrServerClosed {
			if a.log != nil {
				a.log.Error("HTTP server error: %v", err)
			}
		}
		a.notify()
	}()

	a.notify()
	return nil
}

// StopHTTP detiene el servidor HTTP.
func (a *App) StopHTTP() {
	a.mu.Lock()
	srv := a.httpSrv
	a.mu.Unlock()

	if srv == nil {
		return
	}
	if err := srv.Stop(); err != nil && a.log != nil {
		a.log.Error("StopHTTP: %v", err)
	}

	a.mu.Lock()
	a.httpRunning = false
	a.httpSrv = nil
	a.mu.Unlock()
	a.notify()
}

// StartCOM inicia la lectura serial en una goroutine.
// Siempre crea una instancia nueva (stopCh no es reutilizable).
func (a *App) StartCOM() {
	a.mu.Lock()
	if a.comRunning {
		a.mu.Unlock()
		return
	}

	if a.Cfg.UseMock {
		a.scale = scale.NewMockScale(a.Cfg.ScaleID)
	} else {
		a.scale = scale.NewSerialScale(a.Cfg.SerialPort, a.Cfg.BaudRate, a.Cfg.ScaleID)
	}
	a.comRunning = true
	a.mu.Unlock()

	go a.scale.StartReading()
	go a.pollCOMStatus()
	a.notify()
}

// StopCOM detiene la lectura serial.
func (a *App) StopCOM() {
	a.mu.Lock()
	s := a.scale
	a.mu.Unlock()

	if s == nil {
		return
	}
	s.StopReading()

	a.mu.Lock()
	a.comRunning = false
	a.scale = nil
	a.mu.Unlock()
	a.notify()
}

// GetStatus retorna un snapshot inmutable del estado actual.
func (a *App) GetStatus() Status {
	a.mu.RLock()
	defer a.mu.RUnlock()

	st := Status{
		HTTPRunning: a.httpRunning,
		ActivePort:  a.Cfg.ServerPort,
		SerialPort:  fmt.Sprintf("%s @ %d baud", a.Cfg.SerialPort, a.Cfg.BaudRate),
	}

	if a.Cfg.UseMock {
		st.SerialPort = "MOCK"
	}

	if a.scale != nil {
		r := a.scale.LastReading()
		if r.Weight != "" {
			st.LastWeight = r.Weight
			st.LastReadAt = r.Time.Format("15:04:05")
		}
		if sp, ok := a.scale.(domain.ScaleStatusProvider); ok {
			info := sp.GetStatus()
			st.COMConnected = info.Connected
			st.COMError = info.Error
			st.COMReconnecting = !info.Connected && a.comRunning
		}
	}

	return st
}

// Subscribe devuelve un canal que recibe Status en cada cambio de estado.
func (a *App) Subscribe() <-chan Status {
	ch := make(chan Status, 4)
	a.subsMu.Lock()
	a.subs = append(a.subs, ch)
	a.subsMu.Unlock()
	return ch
}

// Unsubscribe elimina el canal de la lista de suscriptores.
func (a *App) Unsubscribe(ch <-chan Status) {
	a.subsMu.Lock()
	defer a.subsMu.Unlock()
	for i, s := range a.subs {
		if s == ch {
			a.subs = append(a.subs[:i], a.subs[i+1:]...)
			close(s)
			return
		}
	}
}

// notify emite el estado actual a todos los suscriptores (non-blocking).
func (a *App) notify() {
	st := a.GetStatus()
	a.subsMu.Lock()
	defer a.subsMu.Unlock()
	for _, ch := range a.subs {
		select {
		case ch <- st:
		default:
		}
	}
}

// pollCOMStatus emite notificaciones periódicas mientras el COM está activo
// para que la GUI actualice el peso en vivo.
func (a *App) pollCOMStatus() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		a.mu.RLock()
		running := a.comRunning
		a.mu.RUnlock()
		if !running {
			return
		}
		a.notify()
	}
}
