/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/a-clap/embedded/pkg/ds18b20"
)

func main() {
	// Get bus handler
	ds, err := ds18b20.NewBus(ds18b20.WithOnewire())
	if err != nil {
		log.Fatalln(err)
	}

	// Find sensors on Bus
	ids, err := ds.IDs()
	if err != nil && len(ids) == 0 {
		log.Fatalln(err)
	}

	// Create sensor from first ID
	sensor, err := ds.NewSensor(ids[0])
	if err != nil {
		log.Fatalln(err)
	}

	// Abillity to configure sensor
	cfg := sensor.GetConfig()
	cfg.Resolution = ds18b20.Resolution10Bit
	err = sensor.Configure(cfg)
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

	// Start background polling
	sensor.Poll()

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
