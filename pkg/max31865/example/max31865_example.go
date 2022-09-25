package main

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/gpio"
	"github.com/a-clap/beaglebone/pkg/max31865"
	"github.com/warthog618/gpiod"
	"log"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"time"
)

type maxTransfer struct {
	spi.Conn
}

func (m *maxTransfer) TxRx(w []byte) ([]byte, error) {
	buf := make([]byte, len(w))
	err := m.Tx(w, buf)
	return buf, err
}

var _ max31865.Transfer = &maxTransfer{}

var ch chan struct{}

func event(evt gpiod.LineEvent) {
	if evt.Type == gpiod.LineEventFallingEdge {
		ch <- struct{}{}
	}

}

func main() {
	// Make sure periph is initialized.
	_, err := host.Init()
	if err != nil {
		log.Fatal(err)
	}
	// go package import.
	if _, err := driverreg.Init(); err != nil {
		log.Fatal(err)
	}

	// Use spireg SPI port registry to find the first available SPI bus.
	p, err := spireg.Open("/dev/spidev0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Convert the spi.Port into a spi.Conn so it can be used for communication.
	c, err := p.Connect(physic.MegaHertz, spi.Mode1, 8)
	if err != nil {
		log.Fatal(err)
	}
	maxTransfer := &maxTransfer{c}

	m, err := max31865.NewDefault(maxTransfer)
	if err != nil {
		log.Fatal(err)
	}

	ch = make(chan struct{})

	in, err := gpio.Input(16,
		gpiod.WithPullUp,
		gpiod.WithFallingEdge,
		gpiod.WithDebounce(10*time.Microsecond),
		gpiod.WithEventHandler(event),
	)
	if err != nil {
		log.Fatal(err)
	}

	val, err := in.Get()
	if err != nil {
		log.Fatal(err)
	}
	// If there is already low level, edge will not be detected
	if !val {
		_, _ = m.Temperature()
	}

	for {
		select {
		case <-ch:
			rtd, _ := m.Temperature()
			fmt.Println("Temperature =", rtd)

		}
	}

}
