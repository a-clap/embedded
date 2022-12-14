package models

import "time"

type SensorType int

// Known SensorType
const (
	DS18B20 SensorType = iota
	PT100
)

type SensorInfo struct {
	Name string     `json:"name"`
	Type SensorType `json:"type"`
	ID   string     `json:"id"`
}

type SensorReadings struct {
	ID          string    `json:"id"`
	Temperature string    `json:"temperature"`
	Stamp       time.Time `json:"stamp"`
	Error       error     `json:"error"`
}
