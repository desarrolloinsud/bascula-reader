package scale

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"

	"github.com/tarm/serial"
)

type SerialScale struct {
	cfg          *serial.Config
	scaleID      string
	last         domain.WeightReading
	mu           sync.RWMutex
	clients      map[chan domain.WeightReading]struct{}
	cmu          sync.Mutex
	connected    bool
	lastError    string
	recentErrors []string
	errorMu      sync.RWMutex
}

func NewSerialScale(port string, baud int, scaleID string) *SerialScale {
	return &SerialScale{
		cfg:          &serial.Config{Name: port, Baud: baud},
		scaleID:      scaleID,
		clients:      make(map[chan domain.WeightReading]struct{}),
		connected:    false,
		recentErrors: make([]string, 0, 10), // Mantener últimos 10 errores
	}
}

func (s *SerialScale) StartReading() {
	appLogger := logger.Get()
	appLogger.Info("Iniciando lectura del puerto serial %s @ %d baud", s.cfg.Name, s.cfg.Baud)

	for {
		ser, err := serial.OpenPort(s.cfg)
		if err != nil {
			errorMsg := fmt.Sprintf("No puedo abrir %s: %v", s.cfg.Name, err)
			s.recordError(errorMsg)
			appLogger.SerialError("abrir puerto", s.cfg.Name, err)
			appLogger.Error("No puedo abrir %s: %v. Reintentando en 2s...", s.cfg.Name, err)
			time.Sleep(2 * time.Second)
			continue
		}

		s.setConnected(true)
		s.clearError()
		appLogger.Info("Conectado a la báscula en %s @ %d baud", s.cfg.Name, s.cfg.Baud)
		appLogger.SerialInfo("Conexión establecida exitosamente en %s @ %d baud", s.cfg.Name, s.cfg.Baud)

		buf := make([]byte, 256)

		for {
			n, err := ser.Read(buf)
			if err != nil {
				errorMsg := fmt.Sprintf("Error leyendo del puerto %s: %v", s.cfg.Name, err)
				s.recordError(errorMsg)
				s.setConnected(false)
				appLogger.SerialError("leer datos", s.cfg.Name, err)
				appLogger.Error("Error leyendo del puerto: %v. Reintentando conexión en 1s...", err)
				break
			}
			if n == 0 {
				continue
			}

			line := strings.TrimSpace(string(buf[:n]))
			if line == "" {
				continue
			}

			appLogger.SerialData(s.cfg.Name, line)
			s.update(line)
		}

		if err := ser.Close(); err != nil {
			errorMsg := fmt.Sprintf("Error cerrando puerto %s: %v", s.cfg.Name, err)
			s.recordError(errorMsg)
			appLogger.SerialError("cerrar puerto", s.cfg.Name, err)
		} else {
			appLogger.SerialInfo("Puerto %s cerrado", s.cfg.Name)
		}
		s.setConnected(false)
		time.Sleep(1 * time.Second)
	}
}

func (s *SerialScale) update(raw string) {
	reading := domain.WeightReading{
		Weight:  raw,
		Time:    time.Now(),
		ScaleID: s.scaleID,
	}

	s.mu.Lock()
	s.last = reading
	s.mu.Unlock()

	appLogger := logger.Get()
	appLogger.Debug("Actualizando lectura: peso=%s, scale_id=%s, tiempo=%s", reading.Weight, reading.ScaleID, reading.Time.Format(time.RFC3339))

	s.broadcast(reading)
}

func (s *SerialScale) LastReading() domain.WeightReading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.last
}

// --- pub/sub para SSE ---

func (s *SerialScale) Subscribe() chan domain.WeightReading {
	ch := make(chan domain.WeightReading, 10)

	s.cmu.Lock()
	s.clients[ch] = struct{}{}
	clientCount := len(s.clients)
	s.cmu.Unlock()

	appLogger := logger.Get()
	appLogger.Debug("Nuevo cliente suscrito. Total de clientes: %d", clientCount)

	return ch
}

func (s *SerialScale) Unsubscribe(c chan domain.WeightReading) {
	s.cmu.Lock()
	delete(s.clients, c)
	clientCount := len(s.clients)
	close(c)
	s.cmu.Unlock()

	appLogger := logger.Get()
	appLogger.Debug("Cliente desuscrito. Total de clientes: %d", clientCount)
}

func (s *SerialScale) broadcast(r domain.WeightReading) {
	s.cmu.Lock()
	clientCount := len(s.clients)
	clients := make([]chan domain.WeightReading, 0, clientCount)
	for ch := range s.clients {
		clients = append(clients, ch)
	}
	s.cmu.Unlock()

	appLogger := logger.Get()
	appLogger.Debug("Transmitiendo lectura a %d clientes suscritos", clientCount)

	for _, ch := range clients {
		select {
		case ch <- r:
			// Enviado exitosamente
		default:
			appLogger.Error("Canal de cliente lleno, descartando lectura")
		}
	}
}

// GetStatus implementa domain.ScaleStatusProvider
func (s *SerialScale) GetStatus() domain.ScaleStatusInfo {
	s.errorMu.RLock()
	defer s.errorMu.RUnlock()

	errors := make([]string, len(s.recentErrors))
	copy(errors, s.recentErrors)

	return domain.ScaleStatusInfo{
		Connected:    s.connected,
		Error:        s.lastError,
		RecentErrors: errors,
	}
}

// setConnected actualiza el estado de conexión
func (s *SerialScale) setConnected(connected bool) {
	s.errorMu.Lock()
	defer s.errorMu.Unlock()
	s.connected = connected
}

// recordError registra un error reciente
func (s *SerialScale) recordError(errorMsg string) {
	s.errorMu.Lock()
	defer s.errorMu.Unlock()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullError := fmt.Sprintf("[%s] %s", timestamp, errorMsg)
	
	s.lastError = fullError
	
	// Mantener solo los últimos 10 errores
	s.recentErrors = append(s.recentErrors, fullError)
	if len(s.recentErrors) > 10 {
		s.recentErrors = s.recentErrors[len(s.recentErrors)-10:]
	}
}

// clearError limpia el último error
func (s *SerialScale) clearError() {
	s.errorMu.Lock()
	defer s.errorMu.Unlock()
	s.lastError = ""
}