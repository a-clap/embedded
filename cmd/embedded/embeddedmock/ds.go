/*
 * Copyright (c) 2023 a-clad. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embeddedmock

import (
	"math/rand"
	"time"

	"github.com/a-clap/embedded/pkg/avg"
	"github.com/a-clap/embedded/pkg/ds18b20"
)

type DS struct {
	bus, id string
	polling bool
	cfg     ds18b20.SensorConfig
	average *avg.Avg
	r       ds18b20.Readings
}

func NewDS(bus, id string) *DS {
	d := &DS{
		bus: bus,
		id:  id,
		cfg: ds18b20.SensorConfig{
			ID:           id,
			Correction:   0,
			Resolution:   ds18b20.Resolution12Bit,
			PollInterval: 0,
			Samples:      10,
		},
		polling: false,
		r:       ds18b20.Readings{},
		average: nil,
	}
	d.average = avg.New(10)
	return d
}
func (d *DS) ID() string {
	return d.id
}

func (d *DS) Name() (bus string, id string) {
	return d.bus, d.id
}

func (d *DS) Poll() {
	d.polling = true
}

func (d *DS) Temperature() (actual, average float64, err error) {
	return d.r.Temperature, d.Average(), nil
}

func (d *DS) GetReadings() []ds18b20.Readings {
	const min = 75.0
	const max = 76.0

	if d.polling {
		t := min + rand.Float64()*(max-min)
		t += d.cfg.Correction

		d.average.Add(t)

		d.r = ds18b20.Readings{
			ID:          d.id,
			Temperature: t,
			Average:     d.Average(),
			Stamp:       time.Now(),
			Error:       "",
		}
		return []ds18b20.Readings{d.r}
	}
	return nil
}

func (d *DS) Average() float64 {
	return d.average.Average()
}

func (d *DS) Configure(config ds18b20.SensorConfig) error {
	d.cfg = config
	d.average.Resize(d.cfg.Samples)
	return nil
}

func (d *DS) GetConfig() ds18b20.SensorConfig {
	return d.cfg
}

func (d *DS) Close() {
	d.polling = false
}
