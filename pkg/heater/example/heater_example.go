/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	"time"

	"github.com/a-clap/embedded/pkg/gpio"
	"github.com/a-clap/embedded/pkg/heater"
)

func main() {
	// Get predefined gpio pin
	pin := gpio.GetBananaPin(gpio.PWR_LED)
	// Create 'heater' on specified PIN
	heater, err := heater.New(heater.WithGpioHeating(pin, "my_heater", gpio.Low))
	if err != nil {
		log.Fatalln(err)
	}

	// Set heater power (range 0 - 100%)
	err = heater.SetPower(95)
	if err != nil {
		log.Fatalln(err)
	}

	// Create channel on which we can receive errors
	errChan := make(chan error, 10)
	// Enable heater in background
	heater.Enable(errChan)
	// Disable on leave
	defer heater.Disable()

	<-time.After(1 * time.Second)

	// Change power on fly
	err = heater.SetPower(13)
	if err != nil {
		log.Fatalln(err)
	}
}
