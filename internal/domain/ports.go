package domain

// Port de entrada: cómo el resto del sistema ve una báscula
type Scale interface {
	StartReading()                // arranca el loop (goroutine afuera o adentro)
	LastReading() WeightReading   // último peso leído
	Subscribe() chan WeightReading // stream para SSE
	Unsubscribe(ch chan WeightReading)
}