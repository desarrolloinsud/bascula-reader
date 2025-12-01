package scale

import (
	"fmt"
	"time"

	"bascula-connector/internal/domain"
)

type MockScale struct {
	*SerialScale 
}

func NewMockScale(scaleID string) *MockScale {
	ms := &MockScale{
		SerialScale: &SerialScale{
			scaleID: scaleID,
			clients: make(map[chan domain.WeightReading]struct{}),
		},
	}
	return ms
}

func (m *MockScale) StartReading() {
	weight := 0.0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		weight += 0.5
		raw := fmt.Sprintf("%.2f kg", weight)
		m.update(raw) // reutiliza update de SerialScale
	}
}