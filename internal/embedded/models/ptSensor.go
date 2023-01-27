package models

type PTWiring int

const (
	TwoWire PTWiring = iota + 2
	ThreeWire
	FourWire
)

type PTConfig struct {
	ID      string   `json:"id"`
	Enabled bool     `json:"enabled"`
	Wiring  PTWiring `json:"wiring"`
	Samples uint     `json:"samples"`
}

type PTSensor interface {
	Temperature() Temperature
	Poll() error
	StopPoll() error
	Config() PTConfig
	SetConfig(cfg PTConfig) (err error)
}
