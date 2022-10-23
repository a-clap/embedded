package max31865

import (
	"log"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"sync"
)

type spiHandler struct {
	init    sync.Once
	count   map[string]int
	handler map[string]spi.PortCloser
}

type spiTransfer struct {
	name string
	conn spi.Conn
}

var spiHandle = spiHandler{
	init:    sync.Once{},
	count:   make(map[string]int),
	handler: make(map[string]spi.PortCloser),
}

func newSpi(devFile string) (Transfer, error) {
	spiHandle.init.Do(func() {
		if _, err := host.Init(); err != nil {
			panic(err)
		}

		if _, err := driverreg.Init(); err != nil {
			panic(err)
		}

	})
	_, ok := spiHandle.count[devFile]
	if !ok {
		// There is not such device, create one
		p, err := spireg.Open(devFile)
		if err != nil {
			return nil, err
		}
		spiHandle.handler[devFile] = p
		spiHandle.count[devFile] = 1
	} else {
		// Increment count
		spiHandle.count[devFile]++
	}

	p, _ := spiHandle.handler[devFile]
	conn, err := p.Connect(5*physic.MegaHertz, spi.Mode1, 8)
	if err != nil {
		return nil, err
	}
	return spiTransfer{conn: conn, name: devFile}, nil
}

func (s *spiHandler) Close(devFile string) error {
	count, ok := spiHandle.count[devFile]
	if !ok {
		log.Println(devFile, "doesn't exist")
		return nil
	}

	if count--; count == 0 {
		err := spiHandle.handler[devFile].Close()
		delete(spiHandle.count, devFile)
		delete(spiHandle.handler, devFile)
		return err
	}

	return nil
}

func (s spiTransfer) ReadWrite(write []byte) (read []byte, err error) {
	read = make([]byte, len(write))
	err = s.conn.Tx(write, read)
	return read, err
}

func (s spiTransfer) Close() error {
	return spiHandle.Close(s.name)
}
