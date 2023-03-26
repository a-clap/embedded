/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/a-clap/iot/internal/embedded/max31865"
)

func main() {
	dev, err := max31865.NewSensor(
		max31865.WithSpidev("/dev/spidev0.0"),
		max31865.WithWiring(max31865.ThreeWire),
		max31865.WithRefRes(430.0),
		max31865.WithRNominal(100.0),
	)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {

		t, _, err := dev.Temperature()
		if err != nil {
			panic(err)
		}
		fmt.Println(t)
		<-time.After(1 * time.Second)
	}
	err = dev.Close()
	if err != nil {
		log.Fatalln(err)
	}
}
