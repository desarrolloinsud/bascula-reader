package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"
	"bascula-connector/internal/status"
)

type HTTPServer struct {
	scale         domain.Scale
	scaleID       string
	addr          string
	port          string
	httpPort      int
	allowedOrigin string
	serialPort    string
	baudRate      int
	useMock       bool
	statusSender  *status.StatusSender
	srv           *http.Server
}

func NewHTTPServer(scale domain.Scale, scaleID string, port string, allowedOrigin string, serialPort string, baudRate int, useMock bool, statusSender *status.StatusSender) *HTTPServer {
	httpPort, _ := strconv.Atoi(port)

	return &HTTPServer{
		scale:         scale,
		scaleID:       scaleID,
		addr:          "127.0.0.1:" + port,
		port:          port,
		httpPort:      httpPort,
		allowedOrigin: allowedOrigin,
		serialPort:    serialPort,
		baudRate:      baudRate,
		useMock:       useMock,
		statusSender:  statusSender,
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", s.loggingMiddleware(s.handlePing))
	mux.HandleFunc("/status", s.loggingMiddleware(s.handleStatus))
	mux.HandleFunc("/weight", s.loggingMiddleware(s.handleWeight))
	mux.HandleFunc("/stream", s.loggingMiddleware(s.handleStream))

	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	appLogger := logger.Get()
	appLogger.Info("Servidor HTTP escuchando en http://%s ...", s.addr)
	return s.srv.ListenAndServe()
}

func (s *HTTPServer) Stop() error {
	if s.srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// loggingMiddleware registra todas las peticiones HTTP
func (s *HTTPServer) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		appLogger := logger.Get()

		// Crear un ResponseWriter que capture el status code
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Ejecutar el handler
		next(lw, r)

		// Calcular duración
		duration := time.Since(start)

		// Registrar la petición
		appLogger.HTTPRequest(r.Method, r.URL.Path, r.RemoteAddr, lw.statusCode, duration)

		// También registrar errores si los hay
		if lw.statusCode >= 400 {
			appLogger.Error("HTTP Error: %s %s desde %s - Status: %d", r.Method, r.URL.Path, r.RemoteAddr, lw.statusCode)
		}
	}
}

// loggingResponseWriter envuelve http.ResponseWriter para capturar el status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

type runtimeStatus struct {
	Connected    bool
	Error        string
	RecentErrors []string
}

type statusResponse struct {
	Status          string   `json:"status"`
	ScaleID         string   `json:"scale_id"`
	SerialPort      string   `json:"serial_port"`
	BaudRate        int      `json:"baud_rate"`
	HTTPPort        int      `json:"http_port"`
	UseMock         bool     `json:"use_mock"`
	LastWeight      string   `json:"last_weight"`
	LastReadAt      string   `json:"last_read_at,omitempty"`
	SerialConnected bool     `json:"serial_connected"`
	SerialError     string   `json:"serial_error,omitempty"`
	RecentErrors    []string `json:"recent_errors,omitempty"`
	Timestamp       string   `json:"timestamp"`
}

func (s *HTTPServer) currentRuntimeStatus() runtimeStatus {
	if statusProvider, ok := s.scale.(domain.ScaleStatusProvider); ok {
		statusInfo := statusProvider.GetStatus()
		return runtimeStatus{
			Connected:    statusInfo.Connected,
			Error:        statusInfo.Error,
			RecentErrors: statusInfo.RecentErrors,
		}
	}

	return runtimeStatus{
		Connected: s.useMock,
	}
}

func (s *HTTPServer) buildStatusResponse(status string) statusResponse {
	reading := s.scale.LastReading()
	runtimeStatus := s.currentRuntimeStatus()

	resp := statusResponse{
		Status:          status,
		ScaleID:         s.scaleID,
		SerialPort:      s.serialPort,
		BaudRate:        s.baudRate,
		HTTPPort:        s.httpPort,
		UseMock:         s.useMock,
		LastWeight:      reading.Weight,
		SerialConnected: runtimeStatus.Connected,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	if !reading.Time.IsZero() {
		resp.LastReadAt = reading.Time.Format(time.RFC3339)
	}

	if reading.ScaleID != "" {
		resp.ScaleID = reading.ScaleID
	}

	if runtimeStatus.Error != "" {
		resp.SerialError = runtimeStatus.Error
	}

	if len(runtimeStatus.RecentErrors) > 0 {
		resp.RecentErrors = runtimeStatus.RecentErrors
	}

	return resp
}

func (s *HTTPServer) writeJSON(w http.ResponseWriter, endpoint string, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		appLogger := logger.Get()
		appLogger.Error("Error codificando respuesta JSON en %s: %v", endpoint, err)
	}
}

func (s *HTTPServer) handlePing(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.writeJSON(w, "/ping", s.buildStatusResponse("pong"))
}

func (s *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Enviar status al endpoint remoto cuando se consulta (solo en modo debug)
	// El SendStatus() ya verifica internamente si está en modo debug
	if s.statusSender != nil {
		go func() {
			if err := s.statusSender.SendStatus(); err != nil {
				appLogger := logger.Get()
				appLogger.Error("Error enviando status al consultar /status: %v", err)
			}
		}()
	}

	runtimeStatus := s.currentRuntimeStatus()
	serviceStatus := "running"
	if runtimeStatus.Error != "" && !s.useMock {
		serviceStatus = "error"
	} else if !runtimeStatus.Connected && !s.useMock {
		serviceStatus = "disconnected"
	}

	s.writeJSON(w, "/status", s.buildStatusResponse(serviceStatus))
}

func (s *HTTPServer) handleWeight(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	reading := s.scale.LastReading()
	s.writeJSON(w, "/weight", reading)
}

func (s *HTTPServer) handleStream(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	appLogger := logger.Get()

	flusher, ok := w.(http.Flusher)
	if !ok {
		appLogger.Error("Streaming no soportado para cliente %s", r.RemoteAddr)
		http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.scale.Subscribe()
	defer s.scale.Unsubscribe(ch)

	// enviar último valor si queremos
	if lr := s.scale.LastReading(); lr.Weight != "" {
		b, err := json.Marshal(lr)
		if err != nil {
			appLogger.Error("Error serializando lectura en /stream: %v", err)
		} else {
			fmt.Fprintf(w, "data: %s\n\n", string(b))
			flusher.Flush()
		}
	}

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case reading := <-ch:
			b, err := json.Marshal(reading)
			if err != nil {
				appLogger.Error("Error serializando lectura en stream SSE: %v", err)
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", string(b)); err != nil {
				appLogger.Error("Error escribiendo en stream SSE: %v", err)
				return
			}
			flusher.Flush()
		}
	}
}

func (s *HTTPServer) enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", s.allowedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Headers",
		"Content-Type, X-Requested-With, X-CSRF-TOKEN",
	)
}
