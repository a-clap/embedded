package dest

import (
	"errors"
)

type Config interface {
	Heaters() ConfigHeaters
}

var (
	_ Config = (*config)(nil)
)

var (
	ErrHeaterNotFound = errors.New("heater with specified hardware ID wasn't found")
)

type config struct {
	heaters *Heaters
}

func newConfig() *config {
	c := &config{heaters: new(Heaters)}
	return c
}

func (g *config) Heaters() ConfigHeaters {
	return g.heaters
}
