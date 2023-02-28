/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embeddedmock

import (
	"github.com/a-clap/iot/internal/embedded/gpio"
)

type GPIO struct {
	state bool
	cfg   gpio.Config
}

func NewGPIO(id string, state bool, direction gpio.Direction) *GPIO {
	return &GPIO{
		state: state,
		cfg: gpio.Config{
			ID:          id,
			Direction:   direction,
			ActiveLevel: gpio.High,
			Value:       false,
		},
	}
}

func (g *GPIO) ID() string {
	return g.cfg.ID
}

func (g *GPIO) Get() (bool, error) {
	return g.state, nil
}

func (g *GPIO) Configure(config gpio.Config) error {
	g.cfg = config
	return nil
}

func (g *GPIO) GetConfig() (gpio.Config, error) {
	return g.cfg, nil
}
