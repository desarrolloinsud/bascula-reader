package main

import (
	"log"

	"bascula-connector/internal/config"
	"bascula-connector/internal/domain"
	"bascula-connector/internal/scale"
	"bascula-connector/internal/server"
)

func main() {
	cfg := config.Load()

	log.Printf(
		"Configuración: puerto HTTP=%s, serial=%s @ %d, mock=%v, origin=%s, scale_id=%s",
		cfg.ServerPort, cfg.SerialPort, cfg.BaudRate, cfg.UseMock, cfg.AllowedOrigin, cfg.ScaleID,
	)

	// Elegimos implementación de la báscula según la config
	var s domain.Scale
	if cfg.UseMock {
		s = scale.NewMockScale(cfg.ScaleID)
	} else {
		s = scale.NewSerialScale(cfg.SerialPort, cfg.BaudRate, cfg.ScaleID)
	}

	// Loop de lectura (goroutine)
	go s.StartReading()

	// HTTP server
	httpSrv := server.New(s, cfg.ServerPort, cfg.AllowedOrigin)
	if err := httpSrv.Start(); err != nil {
		log.Fatalf("Error iniciando HTTP server: %v", err)
	}
}