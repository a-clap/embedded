package main

import (
	"github.com/a-clap/beaglebone/pkg/gpio"
	"log"
	"time"
)

func main() {
	out, err := gpio.Output(gpio.USR3_LED, false)
	if err != nil {
		log.Fatal(err)
	}

	states := []struct {
		delay time.Duration
		value bool
	}{
		{
			delay: 120 * time.Millisecond,
			value: true,
		},
		{
			delay: 60 * time.Millisecond,
			value: false,
		},
		{
			delay: 160 * time.Millisecond,
			value: true,
		},
		{
			delay: 300 * time.Millisecond,
			value: false,
		},
	}

	for {
		for _, state := range states {
			err = out.Set(state.value)
			if err != nil {
				log.Println(err)
			}
			<-time.After(state.delay)
		}
	}
}
