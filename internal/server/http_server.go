package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bascula-connector/internal/domain"
)

type HTTPServer struct {
	scale         domain.Scale
	addr          string
	allowedOrigin string
}

func New(scale domain.Scale, port string, allowedOrigin string) *HTTPServer {
	return &HTTPServer{
		scale:         scale,
		addr:          "127.0.0.1:" + port,
		allowedOrigin: allowedOrigin,
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/weight", s.handleWeight)
	mux.HandleFunc("/stream", s.handleStream)

	log.Printf("Servidor HTTP escuchando en http://%s ...", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	reading := s.scale.LastReading()
	resp := map[string]interface{}{
		"status":       "running",
		"last_weight":  reading.Weight,
		"last_read_at": reading.Time.Format(time.RFC3339),
		"scale_id":     reading.ScaleID,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) handleWeight(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	reading := s.scale.LastReading()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reading)
}

func (s *HTTPServer) handleStream(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
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
		b, _ := json.Marshal(lr)
		fmt.Fprintf(w, "data: %s\n\n", string(b))
		flusher.Flush()
	}

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case reading := <-ch:
			b, _ := json.Marshal(reading)
			fmt.Fprintf(w, "data: %s\n\n", string(b))
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

