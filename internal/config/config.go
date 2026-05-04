package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort        string
	SerialPort        string
	BaudRate          int
	AllowedOrigin     string
	UseMock           bool
	ScaleID           string
	EnabledDebugMode  bool
	LogDir            string
	StatusEndpointURL string
}

// ensureEnvFile crea un archivo .env por defecto si no existe.
func ensureEnvFile(envPath string) {
	_ = ensureEnvFileWithErrors(envPath)
}

// ensureEnvFileWithErrors crea un archivo .env por defecto si no existe y retorna errores
func ensureEnvFileWithErrors(envPath string) error {
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		tmpl := `# Configuración del conector de báscula
SERVER_PORT=8080
SERIAL_PORT=COM1
BAUD_RATE=9600
ALLOWED_ORIGIN=*
MOCK_SCALE=true
SCALE_ID=bascula-1
ENABLED_DEBUG_MODE=true
# LOG_DIR vacío = logs en la misma carpeta del ejecutable
# LOG_DIR=logs
STATUS_ENDPOINT_URL=https://cfc.fresa.com.ar/pushBasculaStatus
			`
		if err := os.WriteFile(envPath, []byte(tmpl), 0644); err != nil {
			return fmt.Errorf("no se pudo crear .env por defecto en %s: %v", envPath, err)
		}
		log.Printf(".env no existía, se creó uno por defecto en %s", envPath)
	}
	return nil
}

// Load carga la configuración desde .env (ubicado al lado del ejecutable)
func Load() Config {
	cfg, _ := LoadWithErrors()
	return cfg
}

// LogWriter es una función para escribir logs tempranos (antes de que el logger esté inicializado)
type LogWriter func(logDir, level, message string)

// LoadWithErrors carga la configuración y retorna también los errores encontrados durante la carga
// Si se proporciona un logWriter, los errores se escriben directamente al archivo de log
func LoadWithErrors(logWriter ...LogWriter) (Config, []string) {
	var errors []string
	var earlyLog LogWriter
	if len(logWriter) > 0 {
		earlyLog = logWriter[0]
	}
	// Ruta del ejecutable (para buscar el .env junto al .exe/bin)
	exePath, err := os.Executable()
	if err != nil {
		errMsg := fmt.Sprintf("No se pudo obtener ruta del ejecutable: %v, usando directorio actual", err)
		errors = append(errors, errMsg)
		if earlyLog != nil {
			earlyLog(".", "ERROR", errMsg)
		}
		exePath = "." // fallback
	}
	exeDir := filepath.Dir(exePath)

	// Si estamos en go run (build temp)
	if strings.Contains(exePath, os.TempDir()) {
		wd, _ := os.Getwd()
		exeDir = wd
	}

	envPath := filepath.Join(exeDir, ".env")

	// Crear .env si no existe (usar exeDir como logDir temporal)
	if err := ensureEnvFileWithErrors(envPath); err != nil {
		errMsg := fmt.Sprintf("Error creando archivo .env: %v", err)
		errors = append(errors, errMsg)
		if earlyLog != nil {
			earlyLog(exeDir, "ERROR", errMsg)
		}
	}

	// Cargar variables
	if err := godotenv.Load(envPath); err != nil {
		errMsg := fmt.Sprintf("Error cargando archivo .env desde %s: %v", envPath, err)
		errors = append(errors, errMsg)
		if earlyLog != nil {
			earlyLog(exeDir, "ERROR", errMsg)
		}
	}

	// LogDir: calcularlo después de cargar el .env
	// Por defecto usa el directorio del ejecutable directamente
	logDir := getEnv("LOG_DIR", "")
	if logDir == "" {
		// Si no se especifica, usar el directorio del ejecutable directamente
		logDir = exeDir
	} else if !filepath.IsAbs(logDir) {
		// Si es relativo, unirlo al directorio del ejecutable
		logDir = filepath.Join(exeDir, logDir)
	}

	cfg := Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		AllowedOrigin:     getEnv("ALLOWED_ORIGIN", "*"),
		UseMock:           getEnv("MOCK_SCALE", "false") == "true",
		ScaleID:           getEnv("SCALE_ID", "bascula-1"),
		EnabledDebugMode:  getEnv("ENABLED_DEBUG_MODE", "true") == "true",
		LogDir:            logDir,
		StatusEndpointURL: getEnv("STATUS_ENDPOINT_URL", "https://cfc.fresa.com.ar/pushBasculaStatus"),
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
		errMsg := fmt.Sprintf("BAUD_RATE inválido (%s), usando valor por defecto 9600", baudStr)
		errors = append(errors, errMsg)
		if earlyLog != nil {
			earlyLog(logDir, "WARN", errMsg)
		}
		baud = 9600
	}
	cfg.BaudRate = baud

	return cfg, errors
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
