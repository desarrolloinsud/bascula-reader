package scale

import (
	"fmt"
	"time"

	"bascula-connector/internal/domain"
	"bascula-connector/internal/logger"
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
	appLogger := logger.Get()
	appLogger.Info("Iniciando modo MOCK de báscula (scale_id=%s)", m.scaleID)
	appLogger.Debug("Modo MOCK: generando lecturas simuladas cada 1 segundo")

	weight := 0.0
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		weight += 0.5
		raw := fmt.Sprintf("%.2f kg", weight)
		appLogger.Debug("Modo MOCK: generando lectura simulada: %s", raw)
		m.update(raw) // reutiliza update de SerialScale
	}
}

// GetStatus implementa domain.ScaleStatusProvider para MockScale
func (m *MockScale) GetStatus() domain.ScaleStatusInfo {
	return domain.ScaleStatusInfo{
		Connected:    true, // En modo mock siempre está "conectado"
		Error:        "",
		RecentErrors: []string{},
	}
}