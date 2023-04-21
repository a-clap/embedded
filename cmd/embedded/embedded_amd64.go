/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	
	"github.com/a-clap/embedded/cmd/embedded/embeddedmock"
	"github.com/a-clap/embedded/pkg/embedded"
	"github.com/a-clap/embedded/pkg/gpio"
)

func main() {
	var err error
	opts := getOpts()
	if false {
		handler, err := embedded.NewRest(opts...)
		if err != nil {
			log.Fatalln(err)
		}
		err = handler.Run()
	} else {
		handler, err := embedded.NewRPC(opts...)
		if err != nil {
			log.Fatalln(err)
		}
		err = handler.Run()
		log.Fatalln(err)
	}
	log.Println(err)
}

func getOpts() []embedded.Option {
	
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
	var opts []embedded.Option
	opts = append(opts, embedded.WithPT(pts))
	opts = append(opts, embedded.WithDS18B20(dss))
	opts = append(opts, embedded.WithHeaters(heaters))
	opts = append(opts, embedded.WithGPIOs(gpios))
	return opts
}
