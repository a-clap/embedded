/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	"os"

	"github.com/a-clap/iot/cmd/embedded/embeddedmock"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/spf13/viper"
)

func main() {

	var handler *embedded.Handler
	if _, ok := os.LookupEnv("TESTING"); ok {
		handler = getMockedEmbedded()
	} else {
		handler = getEmbeddedFromConfig()
	}
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

func getEmbeddedFromConfig() *embedded.Handler {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	cfg := embedded.Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	e, err := embedded.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}
	return e
}
