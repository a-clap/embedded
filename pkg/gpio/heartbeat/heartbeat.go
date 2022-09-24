package main

import (
	"github.com/a-clap/beaglebone/pkg/gpio"
	"github.com/a-clap/logger"
	"go.uber.org/zap/zapcore"
	"time"
)

const (
	USR3_LED = 24
)

func main() {
	log := logger.NewDefaultZap(zapcore.DebugLevel)
	gpio.Log = log

	out, err := gpio.Output(USR3_LED, false)
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
				log.Info(err)
			}
			<-time.After(state.delay)
		}
	}
}
