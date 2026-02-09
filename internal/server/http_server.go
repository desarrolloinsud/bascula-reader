package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"
	"bascula-connector/internal/status"
)

type HTTPServer struct {
	scale         domain.Scale
	addr          string
	allowedOrigin string
	serialPort    string
	baudRate      int
	useMock       bool
	statusSender  *status.StatusSender
}

func NewHTTPServer(scale domain.Scale, port string, allowedOrigin string, serialPort string, baudRate int, useMock bool, statusSender *status.StatusSender) *HTTPServer {
	return &HTTPServer{
		scale:         scale,
		addr:          "127.0.0.1:" + port,
		allowedOrigin: allowedOrigin,
		serialPort:    serialPort,
		baudRate:      baudRate,
		useMock:       useMock,
		statusSender:  statusSender,
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", s.loggingMiddleware(s.handleStatus))
	mux.HandleFunc("/weight", s.loggingMiddleware(s.handleWeight))
	mux.HandleFunc("/stream", s.loggingMiddleware(s.handleStream))

	appLogger := logger.Get()
	appLogger.Info("Servidor HTTP escuchando en http://%s ...", s.addr)
	return http.ListenAndServe(s.addr, mux)
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

	reading := s.scale.LastReading()

	resp := map[string]interface{}{
		"status":       "running",
		"serial_port":  s.serialPort,
		"baud_rate":    s.baudRate,
		"use_mock":     s.useMock,
		"last_weight":  reading.Weight,
		"last_read_at": reading.Time.Format(time.RFC3339),
		"scale_id":     reading.ScaleID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		appLogger := logger.Get()
		appLogger.Error("Error codificando respuesta JSON en /status: %v", err)
	}
}

func (s *HTTPServer) handleWeight(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	reading := s.scale.LastReading()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reading); err != nil {
		appLogger := logger.Get()
		appLogger.Error("Error codificando respuesta JSON en /weight: %v", err)
	}
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

