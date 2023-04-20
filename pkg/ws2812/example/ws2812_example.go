/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"log"
	"time"
	
	"github.com/a-clap/embedded/pkg/ws2812"
)

const (
	zero byte = 0b11000000
	one  byte = 0b11111000
)

type Color struct {
	r, g, b uint8
}

func (c Color) color() []byte {
	red := c.generic(c.r)
	green := c.generic(c.g)
	blue := c.generic(c.b)
	
	full := make([]byte, 0, 24)
	full = append(full, green...)
	full = append(full, red...)
	full = append(full, blue...)
	return full
}

func (c Color) generic(u uint8) []byte {
	gen := make([]byte, 8)
	for i, k := 0, 7; k >= 0; k-- {
		if (u & (1 << k)) == 0 {
			gen[i] = zero
		} else {
			gen[i] = one
		}
		i++
	}
	return gen
}

const LEDS = 8

func main() {
	w, err := ws2812.NewDefault("/dev/spidev1.0", LEDS)
	if err != nil {
		panic(err)
	}
	for {
		for i := 0; i < LEDS; i++ {
			_ = w.SetColor(uint(i), 50, 0, 0)
			
			err := w.Refresh()
			if err != nil {
				log.Println(err)
			}
			_ = w.SetColor(uint(i), 0, 0, 0)
			<-time.After(30 * time.Millisecond)
		}
		for i := 6; i > 0; i-- {
			_ = w.SetColor(uint(i), 50, 0, 0)
			
			err := w.Refresh()
			if err != nil {
				log.Println(err)
			}
			_ = w.SetColor(uint(i), 0, 0, 0)
			<-time.After(30 * time.Millisecond)
		}
	}
	
}
