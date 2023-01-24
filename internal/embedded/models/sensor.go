package models

import "time"

type PollData interface {
	ID() string
	Temperature() float32
	Stamp() time.Time
	Error() error
}

type Temperature struct {
	ID          string    `json:"id"`
	Enabled     bool      `json:"enabled"`
	Temperature float32   `json:"temperature"`
	Stamp       time.Time `json:"stamp"`
}
