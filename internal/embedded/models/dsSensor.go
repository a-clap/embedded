package models

import "time"

type Resolution int

const (
	Resolution9BIT  Resolution = 9
	Resolution10BIT Resolution = 10
	Resolution11BIT Resolution = 11
	Resolution12BIT Resolution = 12
	DefaultSamples  uint       = 5
)

type DSConfig struct {
	ID             string     `json:"id"`
	Enabled        bool       `json:"enabled"`
	Resolution     Resolution `json:"resolution"`
	PollTimeMillis uint       `json:"poll_time_millis"`
	Samples        uint       `json:"samples"`
}

type DSStatus struct {
	ID          string    `json:"id"`
	Enabled     bool      `json:"enabled"`
	Temperature float32   `json:"temperature"`
	Stamp       time.Time `json:"stamp"`
}
