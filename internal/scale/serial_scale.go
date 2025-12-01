package scale

import (
	"log"
	"strings"
	"sync"
	"time"

	"bascula-connector/internal/domain"

	"github.com/tarm/serial"
)

type SerialScale struct {
	cfg      *serial.Config
	scaleID  string
	last     domain.WeightReading
	mu       sync.RWMutex
	clients  map[chan domain.WeightReading]struct{}
	cmu      sync.Mutex
}

func NewSerialScale(port string, baud int, scaleID string) *SerialScale {
	return &SerialScale{
		cfg:     &serial.Config{Name: port, Baud: baud},
		scaleID: scaleID,
		clients: make(map[chan domain.WeightReading]struct{}),
	}
}

func (s *SerialScale) StartReading() {
	for {
		ser, err := serial.OpenPort(s.cfg)
		if err != nil {
			log.Printf("No puedo abrir %s: %v. Reintentando en 2s...", s.cfg.Name, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Conectado a la báscula en %s @ %d baud", s.cfg.Name, s.cfg.Baud)

		buf := make([]byte, 256)

		for {
			n, err := ser.Read(buf)
			if err != nil {
				log.Printf("Error leyendo del puerto: %v. Reintentando en 1s...", err)
				break
			}
			if n == 0 {
				continue
			}

			line := strings.TrimSpace(string(buf[:n]))
			if line == "" {
				continue
			}

			s.update(line)
		}

		_ = ser.Close()
		time.Sleep(1 * time.Second)
	}
}

func (s *SerialScale) update(raw string) {
	reading := domain.WeightReading{
		Weight:  raw,
		Time:    time.Now(),
		ScaleID: s.scaleID,
	}

	s.mu.Lock()
	s.last = reading
	s.mu.Unlock()

	s.broadcast(reading)
}

func (s *SerialScale) LastReading() domain.WeightReading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.last
}

// --- pub/sub para SSE ---

func (s *SerialScale) Subscribe() chan domain.WeightReading {
	ch := make(chan domain.WeightReading, 10)

	s.cmu.Lock()
	s.clients[ch] = struct{}{}
	s.cmu.Unlock()

	return ch
}

func (s *SerialScale) Unsubscribe(c chan domain.WeightReading) {
	s.cmu.Lock()
	delete(s.clients, c)
	close(c)
	s.cmu.Unlock()
}

func (s *SerialScale) broadcast(r domain.WeightReading) {
	s.cmu.Lock()
	defer s.cmu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- r:
		default:
		}
	}
}