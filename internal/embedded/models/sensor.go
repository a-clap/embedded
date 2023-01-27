package models

import "time"

type Temperature struct {
	ID          string    `json:"id"`
	Enabled     bool      `json:"enabled"`
	Temperature float32   `json:"temperature"`
	Stamp       time.Time `json:"stamp"`
}
