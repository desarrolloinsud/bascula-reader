package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bascula-connector/internal/app"
	"bascula-connector/internal/config"
	"bascula-connector/internal/logger"

	fyneapp "fyne.io/fyne/v2/app"
)

func writeEarlyLog(logDir, level, message string) {
	_ = os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, fmt.Sprintf("bascula-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("[%s] %s", level, message)
		return
	}
	defer file.Close()
	fmt.Fprintf(file, "%s [%s] %s\n", time.Now().Format("2006/01/02 15:04:05"), level, message)
}

func main() {
	cfg, initErrors := config.LoadWithErrors(func(logDir, level, message string) {
		writeEarlyLog(logDir, level, message)
	})

	if err := logger.Init(cfg.LogDir, cfg.EnabledDebugMode); err != nil {
		log.Fatalf("Error inicializando logger: %v", err)
	}
	defer logger.Close()

	appLog := logger.Get()
	for _, e := range initErrors {
		appLog.Error("Inicialización: %s", e)
	}

	connector := app.New(cfg, appLog)

	fa := fyneapp.New()

	if shouldShowWizard(cfg) {
		showWizard(fa, connector, cfg)
	} else {
		connector.StartAll()
		showDashboard(fa, connector)
	}

	fa.Run()
}

// shouldShowWizard retorna true si la config es la de primer uso (default sin modificar).
func shouldShowWizard(cfg config.Config) bool {
	return cfg.ScaleID == "bascula-1"
}
