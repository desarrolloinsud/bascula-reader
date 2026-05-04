package domain

// Port de entrada: cómo el resto del sistema ve una báscula
type Scale interface {
	StartReading()                 // arranca el loop (goroutine afuera o adentro)
	StopReading()                  // detiene el loop limpiamente
	LastReading() WeightReading    // último peso leído
	Subscribe() chan WeightReading // stream para SSE
	Unsubscribe(ch chan WeightReading)
}

// ScaleStatusProvider proporciona información de estado y errores
type ScaleStatusProvider interface {
	GetStatus() ScaleStatusInfo
}

// ScaleStatusInfo contiene información del estado de la báscula
type ScaleStatusInfo struct {
	Connected   bool     // Si el puerto serial está conectado
	Error       string   // Último error o mensaje de error
	RecentErrors []string // Lista de errores recientes
}