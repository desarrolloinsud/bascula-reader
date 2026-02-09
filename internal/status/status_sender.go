package status

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bascula-connector/internal/config"
	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"
)

type StatusSender struct {
	endpointURL string
	scale       domain.Scale
	cfg         config.Config
	client      *http.Client
}

type ScaleStatus struct {
	ScaleID           string                 `json:"scale_id"`
	Status            string                 `json:"status"`
	SerialPort        string                 `json:"serial_port"`
	BaudRate          int                    `json:"baud_rate"`
	UseMock           bool                   `json:"use_mock"`
	LastWeight        string                 `json:"last_weight"`
	LastReadAt        string                 `json:"last_read_at,omitempty"`
	SerialConnected   bool                   `json:"serial_connected"`
	SerialError       string                 `json:"serial_error,omitempty"`
	RecentErrors      []string               `json:"recent_errors,omitempty"`
	Config            map[string]interface{} `json:"config"`
	Timestamp         string                 `json:"timestamp"`
}

func NewStatusSender(endpointURL string, scale domain.Scale, cfg config.Config) *StatusSender {
	return &StatusSender{
		endpointURL: endpointURL,
		scale:       scale,
		cfg:         cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendStatus envía el status actual al endpoint remoto
// Solo se ejecuta si ENABLED_DEBUG_MODE está activado (modo debug)
func (s *StatusSender) SendStatus() error {
	// Verificar si está en modo debug
	if !s.cfg.EnabledDebugMode {
		return nil // No hacer nada si no está en modo debug
	}

	appLogger := logger.Get()

	// Obtener información de la báscula
	lastReading := s.scale.LastReading()

	// Obtener estado del puerto serial y errores (si la escala lo soporta)
	var serialConnected bool
	var serialError string
	var recentErrors []string

	if statusProvider, ok := s.scale.(domain.ScaleStatusProvider); ok {
		status := statusProvider.GetStatus()
		serialConnected = status.Connected
		serialError = status.Error
		recentErrors = status.RecentErrors
	} else {
		// Si no implementa la interfaz, asumimos estado básico
		serialConnected = !s.cfg.UseMock
		if s.cfg.UseMock {
			serialError = "Modo MOCK activo"
		}
	}

	// Preparar datos de configuración
	configData := map[string]interface{}{
		"server_port":         s.cfg.ServerPort,
		"serial_port":         s.cfg.SerialPort,
		"baud_rate":           s.cfg.BaudRate,
		"allowed_origin":      s.cfg.AllowedOrigin,
		"use_mock":            s.cfg.UseMock,
		"scale_id":            s.cfg.ScaleID,
		"enabled_debug_mode": s.cfg.EnabledDebugMode,
		"log_dir":             s.cfg.LogDir,
	}

	// Determinar estado general
	status := "running"
	if serialError != "" && !s.cfg.UseMock {
		status = "error"
	} else if !serialConnected && !s.cfg.UseMock {
		status = "disconnected"
	}

	// Crear payload
	payload := ScaleStatus{
		ScaleID:         s.cfg.ScaleID,
		Status:          status,
		SerialPort:      s.cfg.SerialPort,
		BaudRate:        s.cfg.BaudRate,
		UseMock:         s.cfg.UseMock,
		LastWeight:      lastReading.Weight,
		SerialConnected: serialConnected,
		Config:          configData,
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	if !lastReading.Time.IsZero() {
		payload.LastReadAt = lastReading.Time.Format(time.RFC3339)
	}

	if serialError != "" {
		payload.SerialError = serialError
	}

	if len(recentErrors) > 0 {
		payload.RecentErrors = recentErrors
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		appLogger.Error("Error serializando status para envío: %v", err)
		return fmt.Errorf("error serializando status: %w", err)
	}

	// Enviar al endpoint
	req, err := http.NewRequest("POST", s.endpointURL, bytes.NewBuffer(jsonData))
	if err != nil {
		appLogger.Error("Error creando request de status: %v", err)
		return fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		appLogger.Error("Error enviando status al endpoint %s: %v", s.endpointURL, err)
		return fmt.Errorf("error enviando status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		appLogger.Error("Endpoint de status respondió con código %d", resp.StatusCode)
		return fmt.Errorf("endpoint respondió con código %d", resp.StatusCode)
	}

	appLogger.Debug("Status enviado exitosamente al endpoint %s", s.endpointURL)
	return nil
}

// StartPeriodicSender inicia el envío periódico de status cada 10 minutos
// Solo se ejecuta si ENABLED_DEBUG_MODE está activado (modo debug)
func (s *StatusSender) StartPeriodicSender() {
	// Verificar si está en modo debug
	if !s.cfg.EnabledDebugMode {
		appLogger := logger.Get()
		appLogger.Debug("Envío periódico de status desactivado (ENABLED_DEBUG_MODE=false)")
		return // No iniciar el cron si no está en modo debug
	}

	appLogger := logger.Get()
	appLogger.Info("Iniciando envío periódico de status cada 10 minutos a %s", s.endpointURL)

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Enviar inmediatamente al inicio
	if err := s.SendStatus(); err != nil {
		appLogger.Error("Error en envío inicial de status: %v", err)
	}

	for range ticker.C {
		if err := s.SendStatus(); err != nil {
			appLogger.Error("Error en envío periódico de status: %v", err)
		}
	}
}
