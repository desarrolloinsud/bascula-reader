package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	fileLogger *log.Logger
	file       *os.File
	detailed   bool
	mu         sync.Mutex
}

var (
	instance *Logger
	once     sync.Once
)

// Init inicializa el logger global
func Init(logDir string, detailed bool) error {
	var err error
	once.Do(func() {
		// Crear directorio de logs si no existe
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return
		}

		// Nombre del archivo de log con timestamp
		logFile := filepath.Join(logDir, fmt.Sprintf("bascula-%s.log", time.Now().Format("2006-01-02")))

		// Abrir archivo en modo append
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}

		instance = &Logger{
			fileLogger: log.New(file, "", log.LstdFlags),
			file:       file,
			detailed:   detailed,
		}

		instance.Info("=== Inicio de sesión de logging ===")
		if detailed {
			instance.Info("Logging detallado: ACTIVADO")
		} else {
			instance.Info("Logging detallado: DESACTIVADO")
		}

		go purgeOldLogs(logDir, 30)
	})
	return err
}

// purgeOldLogs elimina archivos bascula-*.log con más de maxDays días de antigüedad.
func purgeOldLogs(logDir string, maxDays int) {
	cutoff := time.Now().AddDate(0, 0, -maxDays)
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if len(name) < 16 || name[:8] != "bascula-" || name[len(name)-4:] != ".log" {
			continue
		}
		// nombre: bascula-2006-01-02.log → fecha en posición 8..18
		dateStr := name[8 : len(name)-4]
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			_ = os.Remove(filepath.Join(logDir, name))
		}
	}
}

// Get retorna la instancia del logger
func Get() *Logger {
	if instance == nil {
		// Fallback: crear logger básico si no se inicializó
		once.Do(func() {
			instance = &Logger{
				fileLogger: log.New(os.Stderr, "", log.LstdFlags),
				detailed:   false,
			}
		})
	}
	return instance
}

// Close cierra el archivo de log
func Close() {
	if instance != nil && instance.file != nil {
		instance.Info("=== Fin de sesión de logging ===")
		_ = instance.file.Close()
	}
}

// Info escribe un mensaje informativo
func (l *Logger) Info(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[INFO] "+format, args...)
	l.fileLogger.Println(msg)
	log.Println(msg) // También a stdout
}

// Error escribe un mensaje de error
func (l *Logger) Error(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[ERROR] "+format, args...)
	l.fileLogger.Println(msg)
	log.Println(msg) // También a stdout
}

// Debug escribe un mensaje de debug (solo si detailed está activado)
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.detailed {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[DEBUG] "+format, args...)
	l.fileLogger.Println(msg)
	log.Println(msg) // También a stdout
}

// HTTPRequest registra una petición HTTP
func (l *Logger) HTTPRequest(method, path, remoteAddr string, statusCode int, duration time.Duration) {
	if !l.detailed {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[HTTP] %s %s desde %s - Status: %d - Duración: %v", method, path, remoteAddr, statusCode, duration)
	l.fileLogger.Println(msg)
	log.Println(msg)
}

// SerialError registra un error del puerto serial
func (l *Logger) SerialError(operation, port string, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[SERIAL] Error en %s (puerto: %s): %v", operation, port, err)
	l.fileLogger.Println(msg)
	log.Println(msg)
}

// SerialInfo registra información del puerto serial
func (l *Logger) SerialInfo(format string, args ...interface{}) {
	if !l.detailed {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[SERIAL] "+format, args...)
	l.fileLogger.Println(msg)
	log.Println(msg)
}

// SerialData registra datos recibidos del puerto serial
func (l *Logger) SerialData(port string, data string) {
	if !l.detailed {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	msg := fmt.Sprintf("[SERIAL] Datos recibidos desde %s: %s", port, data)
	l.fileLogger.Println(msg)
	log.Println(msg)
}
