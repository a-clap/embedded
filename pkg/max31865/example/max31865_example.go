package main

import (
	"fmt"
	"github.com/a-clap/beaglebone/pkg/max31865"
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

func main() {
	//max31865.Log = logger.NewDefaultZap(zapcore.DebugLevel)
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
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
	for i := 0; i < 100; i++ {
		<-time.After(1 * time.Second)
		rtd, err := m.Temperature()
		fmt.Println("rtd =", rtd, ", err = ", err)

	}

}
