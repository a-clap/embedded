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

	ds := ds18b20.NewDefault()

	ids, err := ds.IDs()
	if err != nil && len(ids) == 0 {
		log.Fatal(err)
	}
	sensor, _ := ds.NewSensor(ids[0])

	errs := sensor.Poll(reads, 750*time.Millisecond)
	if errs != nil {
		log.Fatal(err)
	}

	// Just to end this after time
	go func() {
		for {
			select {
			case <-time.After(10 * time.Second):
				_ = sensor.Close()
			}
		}
	}()

	for readings := range reads {
		id := readings.ID()
		tmp, stamp, err := readings.Get()
		fmt.Printf("id: %s, Temperature: %s. Time: %s, err: %v \n", id, tmp, stamp, err)
	}

	fmt.Println("finished")
}
