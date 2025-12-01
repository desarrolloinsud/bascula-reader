package domain

import "time"

type WeightReading struct {
	Weight  string    `json:"weight"`
	Time    time.Time `json:"time"`
	ScaleID string    `json:"scale_id"`
}