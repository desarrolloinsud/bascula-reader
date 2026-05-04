package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bascula-connector/internal/config"
	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"
	"bascula-connector/internal/scale"
	"bascula-connector/internal/server"
	"bascula-connector/internal/status"
)

// writeEarlyLog escribe directamente al archivo de log (antes de que el logger esté inicializado)
// Se usa para errores de instalación/inicialización que DEBEN registrarse siempre
func writeEarlyLog(logDir, level, message string) {
	// Intentar crear el directorio si no existe
	_ = os.MkdirAll(logDir, 0755)

	// Nombre del archivo de log con timestamp
	logFile := filepath.Join(logDir, fmt.Sprintf("bascula-%s.log", time.Now().Format("2006-01-02")))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Si no se puede escribir al archivo, al menos intentar stdout
		log.Printf("[%s] %s", level, message)
		return
	}
	defer file.Close()

	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(file, "%s [%s] %s\n", timestamp, level, message)
}

// writeEmergencyLog escribe un log de emergencia cuando falla la inicialización del logger
func writeEmergencyLog(logDir string, err error) string {
	// Intentar crear el directorio si no existe
	_ = os.MkdirAll(logDir, 0755)

	// Archivo de emergencia
	emergencyFile := filepath.Join(logDir, fmt.Sprintf("bascula-emergency-%s.log", time.Now().Format("20060102-150405")))

	file, fileErr := os.OpenFile(emergencyFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if fileErr != nil {
		return ""
	}
	defer file.Close()

	fmt.Fprintf(file, "[%s] [EMERGENCY] Error inicializando logger: %v\n", time.Now().Format(time.RFC3339), err)
	return emergencyFile
}

func main() {
	// Cargar configuración y capturar errores de inicialización
	// Los errores se escribirán directamente al archivo de log (mandatorios)
	cfg, initErrors := config.LoadWithErrors(func(logDir, level, message string) {
		writeEarlyLog(logDir, level, message)
	})

	// Registrar errores de instalación/inicialización ANTES de inicializar el logger
	// Estos son mandatorios y deben registrarse siempre
	for _, errMsg := range initErrors {
		writeEarlyLog(cfg.LogDir, "ERROR", fmt.Sprintf("Inicialización: %s", errMsg))
	}

	// Inicializar logger
	if err := logger.Init(cfg.LogDir, cfg.EnabledDebugMode); err != nil {
		// Si falla la inicialización del logger, intentar escribir a un archivo de emergencia
		logFile := writeEmergencyLog(cfg.LogDir, err)
		writeEarlyLog(cfg.LogDir, "FATAL", fmt.Sprintf("Error inicializando logger: %v", err))
		log.Fatalf("Error inicializando logger: %v (log de emergencia en: %s)", err, logFile)
	}
	defer logger.Close()

	appLogger := logger.Get()

	// Registrar errores de inicialización en el logger también (si los hubo)
	if len(initErrors) > 0 {
		appLogger.Error("=== Errores durante la inicialización ===")
		for _, errMsg := range initErrors {
			appLogger.Error("Inicialización: %s", errMsg)
		}
	}

	// Elegimos implementación de la báscula según la config
	var s domain.Scale
	if cfg.UseMock {
		appLogger.Info("Usando modo MOCK para la báscula")
		s = scale.NewMockScale(cfg.ScaleID)
	} else {
		appLogger.Info("Inicializando báscula serial en puerto %s @ %d baud", cfg.SerialPort, cfg.BaudRate)
		s = scale.NewSerialScale(cfg.SerialPort, cfg.BaudRate, cfg.ScaleID)
	}

	// Loop de lectura (goroutine)
	appLogger.Info("Iniciando lectura de la báscula...")
	go s.StartReading()

	// Inicializar y iniciar envío periódico de status
	appLogger.Info("Inicializando sistema de envío de status...")
	statusSender := status.NewStatusSender(cfg.StatusEndpointURL, s, cfg)
	go statusSender.StartPeriodicSender()

	// HTTP server
	appLogger.Info("Iniciando servidor HTTP en puerto %s...", cfg.ServerPort)
	httpSrv := server.NewHTTPServer(
		s,
		cfg.ScaleID,
		cfg.ServerPort,
		cfg.AllowedOrigin,
		cfg.SerialPort,
		cfg.BaudRate,
		cfg.UseMock,
		statusSender,
	)

	// Mensaje de inicio exitoso con toda la información relevante
	serialInfo := fmt.Sprintf("%s @ %d baud", cfg.SerialPort, cfg.BaudRate)
	if cfg.UseMock {
		serialInfo = "MODO MOCK (sin puerto serial)"
	}

	appLogger.Info("========================================")
	appLogger.Info("✅ APLICACIÓN INICIADA CORRECTAMENTE")
	appLogger.Info("========================================")
	appLogger.Info("Servidor HTTP: http://127.0.0.1:%s", cfg.ServerPort)
	appLogger.Info("Puerto Serial: %s", serialInfo)
	appLogger.Info("Scale ID: %s", cfg.ScaleID)
	appLogger.Info("Modo Debug: %v", cfg.EnabledDebugMode)
	appLogger.Info("========================================")

	if err := httpSrv.Start(); err != nil {
		appLogger.Error("Error crítico iniciando HTTP server: %v", err)
		appLogger.Error("La aplicación no puede continuar sin el servidor HTTP")
		os.Exit(1)
	}
}
