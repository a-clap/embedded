package models

type PTConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Samples uint   `json:"samples"`
}

type PTSensor interface {
	ID() string
	Temperature() Temperature
	Config() PTConfig
	SetConfig(cfg PTConfig) (err error)
}
