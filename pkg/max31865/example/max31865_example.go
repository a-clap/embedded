package main

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/max31865"
	"log"
	"time"
)

func main() {
	dev, err := max31865.NewDefault("/dev/spidev0.0", max31865.Config{
		Wiring:   max31865.ThreeWire,
		RefRes:   430.0,
		RNominal: 100.0,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {

		t, err := dev.Temperature()
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
