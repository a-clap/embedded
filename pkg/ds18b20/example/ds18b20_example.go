package main

import (
	"fmt"
	"github.com/a-clap/iot/internal/models"
	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/a-clap/logger"
	"go.uber.org/zap/zapcore"
	"time"
)

func main() {
	log := logger.NewDefaultZap(zapcore.DebugLevel)
	ds, err := ds18b20.NewBus(ds18b20.WithOnewire())
	if err != nil {
		log.Panic(err)
	}

	reads := make(chan models.SensorReadings)

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
		fmt.Printf("id: %s, Temperature: %s. Time: %s, err: %v \n", readings.ID, readings.Temperature, readings.Stamp, readings.Error)
	}

	fmt.Println("finished")
}
