package models

type Heater struct {
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
	Power  uint   `json:"power"`
}
