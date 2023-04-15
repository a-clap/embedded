/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	"time"
	
	"github.com/a-clap/iot/pkg/embedded/gpio"
)

func main() {
	out, err := gpio.Output(gpio.GetBananaPin(gpio.PWR_LED), "", false)
	if err != nil {
		panic(err)
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
