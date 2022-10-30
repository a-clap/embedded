package ws2812

import (
	"errors"
	"fmt"
	"github.com/a-clap/iot/pkg/spidev"
	"io"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

const (
	zero byte = 0b11000000
	one  byte = 0b11111000
)

var (
	ErrInterface   = errors.New("interface error")
	ErrLedNotExist = errors.New("specified led number doesn't exist")
)

type led struct {
	r, g, b uint8
}

type WS2812 struct {
	size   uint
	writer io.Writer
	leds   []led
}
type wsSpidevWriter struct {
	spi.Conn
}

func (w wsSpidevWriter) Write(p []byte) (n int, err error) {
	r := make([]byte, len(p))
	if err := w.Tx(p, r); err != nil {
		return 0, err
	}
	return len(p), nil
}

func NewDefault(filename string, size uint) (*WS2812, error) {
	s, err := spidev.New(filename, 6400*physic.KiloHertz, spi.Mode1, 8)
	if err != nil {
		return nil, err
	}
	return New(size, wsSpidevWriter{s}), nil
}

func New(size uint, w io.Writer) *WS2812 {
	return &WS2812{
		size:   size,
		writer: w,
		leds:   make([]led, size),
	}
}

func (w *WS2812) prepare() []byte {
	buf := make([]byte, len(w.leds)*3*8+3)
	buf[0] = 0
	buf[1] = 0
	buf[2] = 0
	pos := 3
	for _, led := range w.leds {
		led.append(buf, &pos)
	}

	return buf
}

func (w *WS2812) SetColor(idx uint, r, g, b uint8) error {
	if idx >= w.size {
		return ErrLedNotExist
	}
	w.leds[idx].r = r
	w.leds[idx].g = g
	w.leds[idx].b = b
	return nil
}

func (w *WS2812) SetAll(r, g, b uint8) {
	for i := 0; i < len(w.leds); i++ {
		_ = w.SetColor(uint(i), r, g, b)
	}
}

// Refresh update leds
func (w *WS2812) Refresh() error {
	b := w.prepare()
	return w.write(b)
}

// write is a wrapper for interface Writer
func (w *WS2812) write(buf []byte) error {
	if n, err := w.writer.Write(buf); err != nil && n != len(buf) {
		return fmt.Errorf("%w: %v", ErrInterface, err)
	}
	return nil
}

func (l led) append(buf []byte, pos *int) {
	l.generic(l.g, buf, pos)
	l.generic(l.r, buf, pos)
	l.generic(l.b, buf, pos)
}

func (l led) generic(u uint8, buf []byte, pos *int) {
	for k := 7; k >= 0; k-- {
		if (u & (1 << k)) == 0 {
			buf[*pos] = zero
		} else {
			buf[*pos] = one
		}
		*pos++
	}

}
