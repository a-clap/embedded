/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ws2812

import (
	"errors"
	"fmt"
	
	"github.com/a-clap/iot/pkg/embedded/spidev"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

const (
	zero       byte = 0b11000000
	one        byte = 0b11111000
	bitsPerLed      = 24
)

var (
	ErrInterface   = errors.New("interface error")
	ErrLedNotExist = errors.New("specified led number doesn't exist")
)

type Writer interface {
	Write([]byte) error
}

type WS2812 struct {
	size      uint
	writer    Writer
	ledBuffer []byte
}
type wsSpidevWriter struct {
	spi.Conn
}

func (w wsSpidevWriter) Write(p []byte) (err error) {
	r := make([]byte, len(p))
	return w.Tx(p, r)
}

func NewDefault(filename string, size uint) (*WS2812, error) {
	s, err := spidev.New(filename, 6400*physic.KiloHertz, spi.Mode1, 8)
	if err != nil {
		return nil, err
	}
	return New(size, wsSpidevWriter{s}), nil
}

func New(size uint, w Writer) *WS2812 {
	led := &WS2812{
		size:      size,
		writer:    w,
		ledBuffer: make([]byte, size*bitsPerLed+3),
	}
	// turn off all
	led.SetAll(0, 0, 0)
	
	return led
}

func (w *WS2812) SetColor(idx uint, r, g, b uint8) error {
	if idx >= w.size {
		return ErrLedNotExist
	}
	ledPos := idx*bitsPerLed + 3
	parseColor(g, w.ledBuffer, &ledPos)
	parseColor(r, w.ledBuffer, &ledPos)
	parseColor(b, w.ledBuffer, &ledPos)
	return nil
}

func (w *WS2812) SetAll(r, g, b uint8) {
	for i := uint(0); i < w.size; i++ {
		_ = w.SetColor(i, r, g, b)
	}
}

// Refresh update leds
func (w *WS2812) Refresh() error {
	return w.write(w.ledBuffer)
}

// write is a wrapper for interface Writer
func (w *WS2812) write(buf []byte) error {
	if err := w.writer.Write(buf); err != nil {
		return fmt.Errorf("%w: %v", ErrInterface, err)
	}
	return nil
}

func parseColor(u uint8, buf []byte, pos *uint) {
	for k := 7; k >= 0; k-- {
		if (u & (1 << k)) == 0 {
			buf[*pos] = zero
		} else {
			buf[*pos] = one
		}
		*pos++
	}
}
