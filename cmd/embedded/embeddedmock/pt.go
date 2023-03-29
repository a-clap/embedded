/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embeddedmock

import (
	"errors"
	"math/rand"
	"time"

	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/a-clap/iot/pkg/avg"
)

type PT struct {
	id      string
	cfg     max31865.SensorConfig
	polling bool
	r       max31865.Readings
	average *avg.Avg
}

func NewPT(id string) *PT {
	pt := &PT{
		id: id,
		cfg: max31865.SensorConfig{
			ID:           id,
			Correction:   0,
			ASyncPoll:    false,
			PollInterval: 0,
			Samples:      10,
		},
		polling: false,
		r:       max31865.Readings{},
		average: nil,
	}
	pt.average = avg.New(pt.cfg.Samples)
	return pt
}

func (p *PT) ID() string {
	return p.id
}

func (p *PT) Poll() (err error) {
	if p.polling {
		return errors.New("already polling")
	}
	p.polling = true
	return nil
}

func (p *PT) Configure(config max31865.SensorConfig) error {
	p.cfg = config
	p.average.Resize(p.cfg.Samples)
	return nil
}

func (p *PT) GetConfig() max31865.SensorConfig {
	return p.cfg
}

func (p *PT) Average() float64 {
	return p.average.Average()
}

func (p *PT) Temperature() (actual float64, average float64, err error) {
	return p.r.Temperature, p.Average(), nil
}

func (p *PT) GetReadings() []max31865.Readings {

	const min = 75.0
	const max = 76.0

	if p.polling {
		t := min + rand.Float64()*(max-min)
		t += p.cfg.Correction

		p.average.Add(t)

		p.r = max31865.Readings{
			ID:          p.id,
			Temperature: t,
			Average:     p.Average(),
			Stamp:       time.Now(),
			Error:       "",
		}
		return []max31865.Readings{p.r}
	}
	return nil
}

func (p *PT) Close() error {
	p.polling = false
	return nil
}
