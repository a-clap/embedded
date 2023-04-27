/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/a-clap/embedded/pkg/max31865"
)

func main() {
	// Create new sensor with initial configuration
	sensor, err := max31865.NewSensor(
		max31865.WithSpidev("/dev/spidev0.0"),
		max31865.WithWiring(max31865.ThreeWire),
		max31865.WithRefRes(430.0),
		max31865.WithNominalRes(100.0),
	)
	if err != nil {
		log.Fatalln(err)
	}

	// Get few readings - on demand
	for i := 0; i < 5; i++ {
		actual, average, err := sensor.Temperature()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Actual: %v, average: %v\n", actual, average)
		<-time.After(1 * time.Second)
	}

	// Handle sensor readings in background
	err = sensor.Poll()
	if err != nil {
		log.Fatalln(err)
	}

	// Get few readings
	<-time.After(3 * time.Second)
	// Disable background polling
	sensor.Close()

	// Read stored readings
	reads := sensor.GetReadings()
	// Print it to user
	for _, readings := range reads {
		fmt.Printf("id: %s, Temperature: %v. Time: %s, err: %v \n", readings.ID, readings.Temperature, readings.Stamp, readings.Error)
	}

	fmt.Println("All good!")
}
