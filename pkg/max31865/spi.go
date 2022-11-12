package max31865

import (
	"github.com/a-clap/iot/pkg/spidev"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

type maxSpidevTransfer struct {
	*spidev.Spidev
}

func newMaxSpidev(devFile string) (*maxSpidevTransfer, error) {
	maxSpi, err := spidev.New(devFile, 5*physic.MegaHertz, spi.Mode1, 8)
	if err != nil {
		return nil, err
	}
	return &maxSpidevTransfer{maxSpi}, nil
}

func (m *maxSpidevTransfer) ReadWrite(write []byte) (read []byte, err error) {
	read = make([]byte, len(write))
	err = m.Spidev.Tx(write, read)
	return read, err
}
