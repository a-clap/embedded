package models

type ActiveLevel int
type Direction int

const (
	Low ActiveLevel = iota
	High
)
const (
	Input Direction = iota
	Output
)

type GPIOConfig struct {
	ID          string      `json:"id"`
	Direction   Direction   `json:"direction"`
	ActiveLevel ActiveLevel `json:"active_level"`
	Value       bool        `json:"value"`
}

type GPIO interface {
	ID() string
	Config() (GPIOConfig, error)
	SetConfig(cfg GPIOConfig) error
}
