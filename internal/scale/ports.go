package scale

import (
	"fmt"
	"runtime"

	"github.com/tarm/serial"
)

// ListAvailablePorts detecta puertos serie disponibles en el sistema.
// En Windows prueba COM1-COM20; en macOS/Linux usa patrones de nombre comunes.
func ListAvailablePorts() []string {
	switch runtime.GOOS {
	case "windows":
		return listWindowsPorts()
	default:
		return listUnixPorts()
	}
}

func listWindowsPorts() []string {
	var found []string
	for i := 1; i <= 20; i++ {
		name := fmt.Sprintf("COM%d", i)
		cfg := &serial.Config{Name: name, Baud: 9600}
		port, err := serial.OpenPort(cfg)
		if err == nil {
			port.Close()
			found = append(found, name)
			continue
		}
		// "Access denied" significa que existe pero está en uso: igualmente lo incluimos.
		errStr := err.Error()
		if contains(errStr, "Access is denied") || contains(errStr, "access denied") {
			found = append(found, name)
		}
	}
	return found
}

func listUnixPorts() []string {
	candidates := []string{
		"/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyUSB2",
		"/dev/ttyS0", "/dev/ttyS1", "/dev/ttyS2",
		"/dev/tty.usbserial", "/dev/tty.usbmodem",
		"/dev/tty.usbserial-0001", "/dev/tty.SLAB_USBtoUART",
	}
	var found []string
	for _, name := range candidates {
		cfg := &serial.Config{Name: name, Baud: 9600}
		port, err := serial.OpenPort(cfg)
		if err == nil {
			port.Close()
			found = append(found, name)
			continue
		}
		// "resource busy" significa que existe pero está en uso.
		if contains(err.Error(), "resource busy") || contains(err.Error(), "permission denied") {
			found = append(found, name)
		}
	}
	return found
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
