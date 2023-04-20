/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"github.com/a-clap/embedded/pkg/embedded/spidev"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

type spidevTransfer struct {
	*spidev.Spidev
}

func newMaxSpidev(devFile string) (*spidevTransfer, error) {
	maxSpi, err := spidev.New(devFile, 5*physic.MegaHertz, spi.Mode1, 8)
	if err != nil {
		return nil, err
	}
	return &spidevTransfer{maxSpi}, nil
}

func (m *spidevTransfer) ReadWrite(write []byte) (read []byte, err error) {
	read = make([]byte, len(write))
	err = m.Spidev.Tx(write, read)
	return read, err
}
