/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"log"
	"time"
	
	"github.com/a-clap/iot/pkg/ds18b20"
)

func main() {
	ds, err := ds18b20.NewBus(ds18b20.WithOnewire())
	if err != nil {
		log.Fatalln(err)
	}
	
	ids, err := ds.IDs()
	if err != nil && len(ids) == 0 {
		log.Fatalln(err)
	}
	
	sensor, _ := ds.NewSensor(ids[0])
	
	errs := sensor.Poll()
	if errs != nil {
		log.Fatalln(err)
	}
	
	// Just to end this after time
	select {
	case <-time.After(10 * time.Second):
		_ = sensor.Close()
	}
	reads := sensor.GetReadings()
	
	for _, readings := range reads {
		fmt.Printf("id: %s, Temperature: %v. Time: %s, err: %v \n", readings.ID, readings.Temperature, readings.Stamp, readings.Error)
	}
	
	fmt.Println("finished")
}
