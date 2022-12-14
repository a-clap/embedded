package max31865

import (
	"github.com/a-clap/iot/pkg/gpio"
	"github.com/warthog618/gpiod"
	"log"
)

type gpioReady struct {
	pin    gpio.Pin
	in     *gpio.In
	cb     func(any) error
	cbArgs any
}

func newGpioReady(in gpio.Pin) (Ready, error) {
	var err error
	m := &gpioReady{pin: in}
	m.in, err = gpio.Input(m.pin, gpiod.WithPullUp, gpiod.WithFallingEdge, gpiod.WithEventHandler(m.eventHandler))
	return m, err
}

func (m *gpioReady) Open(callback func(any) error, args any) error {
	m.cb = callback
	m.cbArgs = args

	return nil
}

func (m *gpioReady) Close() {
	_ = m.in.Close()
}

func (m *gpioReady) eventHandler(event gpiod.LineEvent) {
	if m.cb != nil {
		if err := m.cb(m.cbArgs); err != nil {
			log.Print(err, event)
		}
	}
}
