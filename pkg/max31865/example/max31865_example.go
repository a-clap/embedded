package main

import (
	"fmt"
	"github.com/a-clap/iot/pkg/max31865"
	"log"
	"time"
)

func main() {
	//dev, err := max31865.New("/dev/spidev0.0", max31865.ThreeWire, max31865.RefRes(430.0), max31865.RNominal(100.0))
	dev, err := max31865.New(
		max31865.WithSpidev("/dev/spidev0.0"),
		max31865.WithWiring(max31865.ThreeWire),
		max31865.WithRefRes(430.0),
		max31865.WithRNominal(100.0),
	)
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
