/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"io"
)

// Internal registers in Max
const (
	regConf = iota
	regRtdMsb
	regRtdLsb
	regHFaultMsb
	regHFaultLsb
	regLFaultMsb
	regLFaultLsb
	regFault
)

// ReadWriteCloser is full duplex communication with Max31865
type ReadWriteCloser interface {
	io.Closer
	ReadWrite(write []byte) (read []byte, err error)
}

// Ready is an interface which allows to register a callback
// max31865 has a pin DRDY, which goes low, when new conversion is ready, this interface should rely on that pin
type Ready interface {
	Open(callback func()) error
	Close()
}
