package main

import (
	"fmt"
	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/a-clap/logger"
	"go.uber.org/zap/zapcore"
	"time"
)

func main() {
	log := logger.NewDefaultZap(zapcore.DebugLevel)

	reads := make(chan ds18b20.Readings)
	exitCh := make(chan struct{})

	ds := ds18b20.NewDefault()

	ids, err := ds.Devices()
	if err != nil {
		log.Fatal(err)
	}

	finCh, errCh, errs := ds.Poll(ids, reads, exitCh, 750*time.Millisecond)
	if errs != nil {
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
				id, tmp, stamp := sensor.Get()
				fmt.Printf("ID: %s, Temperature: %s. Time: %s\n", id, tmp, stamp)
			case err := <-errCh:
				fmt.Println("Error from ds18b20", err)
			}
		}
	}()
	<-finCh
	fmt.Println("finished")
}
