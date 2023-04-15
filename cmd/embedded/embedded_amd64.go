/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	
	"github.com/a-clap/iot/cmd/embedded/embeddedmock"
	"github.com/a-clap/iot/pkg/embedded"
	"github.com/a-clap/iot/pkg/embedded/gpio"
)

func main() {
	handler := getMockedEmbedded()
	err := handler.Run()
	log.Println(err)
}

func getMockedEmbedded() *embedded.Handler {
	ptIds := []string{"PT_1", "PT_2", "PT_3"}
	pts := make([]embedded.PTSensor, len(ptIds))
	for i, id := range ptIds {
		pts[i] = embeddedmock.NewPT(id)
	}
	
	dsIds := []struct {
		bus, id string
	}{
		{
			bus: "1",
			id:  "ds_1",
		},
		{
			bus: "1",
			id:  "ds_2",
		},
	}
	dss := make([]embedded.DSSensor, len(dsIds))
	for i, id := range dsIds {
		dss[i] = embeddedmock.NewDS(id.bus, id.id)
	}
	
	heaterIds := []string{"heater_1", "heater_2", "heater_3"}
	heaters := make(map[string]embedded.Heater, len(heaterIds))
	for _, id := range heaterIds {
		heaters[id] = embeddedmock.NewHeater()
	}
	
	gpioIds := []struct {
		id    string
		state bool
		dir   gpio.Direction
	}{
		{
			id:    "gpio_1",
			state: false,
			dir:   gpio.DirInput,
		}, {
			id:    "gpio_2",
			state: true,
			dir:   gpio.DirOutput,
		},
	}
	gpios := make([]embedded.GPIO, len(gpioIds))
	for i, id := range gpioIds {
		gpios[i] = embeddedmock.NewGPIO(id.id, id.state, id.dir)
	}
	
	handler, err := embedded.New(
		embedded.WithPT(pts),
		embedded.WithDS18B20(dss),
		embedded.WithHeaters(heaters),
		embedded.WithGPIOs(gpios),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return handler
}
