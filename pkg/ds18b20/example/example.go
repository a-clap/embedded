package main

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/ds18b20"
	"github.com/a-clap/logger"
	"go.uber.org/zap/zapcore"
	"time"
)

func main() {
	log := logger.NewDefaultZap(zapcore.WarnLevel)
	ds18b20.Log = log

	devices, err := ds18b20.SensorsIDs()
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Found devices ", devices)

	reads := make(chan ds18b20.Sensor)
	exitCh := make(chan struct{})

	finCh, errCh, err := ds18b20.Poll(devices, reads, exitCh, 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	ch := make(chan bool)

	// Just to end this after time
	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				ch <- true
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ch:
				exitCh <- struct{}{}
				return
			case sensor := <-reads:
				tmp, err := sensor.ParseTemperature()
				if err != nil {
					log.Warn("couldn't parse temp ", err)
				}
				fmt.Printf("ID: %s\nTemperature %v\n", sensor.ID, tmp)
			case err := <-errCh:
				fmt.Println("Error from ds18b20", err)
			}
		}
	}()
	<-finCh
	fmt.Println("finished")
}
