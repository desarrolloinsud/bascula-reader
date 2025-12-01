package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	SerialPort    string
	BaudRate      int
	AllowedOrigin string
	UseMock       bool
	ScaleID       string
}

// ensureEnvFile crea un archivo .env por defecto si no existe.
func ensureEnvFile(envPath string) {
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		tmpl := `# Configuración del conector de báscula
SERVER_PORT=7070
SERIAL_PORT=COM1
BAUD_RATE=9600
ALLOWED_ORIGIN=*
MOCK_SCALE=true
SCALE_ID=bascula-1
			`
		err := os.WriteFile(envPath, []byte(tmpl), 0644)
		if err != nil {
			log.Printf("No se pudo crear .env por defecto: %v", err)
		} else {
			log.Printf(".env no existía, se creó uno por defecto en %s", envPath)
		}
	}
}

// Load carga la configuración desde .env (ubicado al lado del ejecutable)
func Load() Config {
	// Ruta del ejecutable (para buscar el .env junto al .exe/bin)
	exePath := getEnvPath()
	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	// Crear .env si no existe
	ensureEnvFile(envPath)

	// Cargar variables
	_ = godotenv.Load(envPath)

	cfg := Config{
		ServerPort:    getEnv("SERVER_PORT", "7070"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "*"),
		UseMock:       getEnv("MOCK_SCALE", "false") == "true",
		ScaleID:       getEnv("SCALE_ID", "bascula-1"),
	}

	// Puerto serie por defecto según SO
	defaultPort := "COM1"
	if runtime.GOOS == "darwin" {
		defaultPort = "/dev/tty.usbserial"
	} else if runtime.GOOS == "linux" {
		defaultPort = "/dev/ttyUSB0"
	}
	cfg.SerialPort = getEnv("SERIAL_PORT", defaultPort)

	// Parseo de baudios
	baudStr := getEnv("BAUD_RATE", "9600")
	baud, err := strconv.Atoi(baudStr)
	if err != nil {
		log.Printf("BAUD_RATE inválido (%s), usando 9600", baudStr)
		baud = 9600
	}
	cfg.BaudRate = baud

	return cfg
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getEnvPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ".env" // fallback
	}

	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	// Si estamos en go run (build temp)
	if strings.Contains(exePath, os.TempDir()) {
		wd, _ := os.Getwd()
		return filepath.Join(wd, ".env")
	}

	return envPath
}