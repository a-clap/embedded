/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"io"
)

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

type ReadWriteCloser interface {
	io.Closer
	ReadWrite(write []byte) (read []byte, err error)
}

// Ready is an interface which allows to register a callback
// max31865 has a pin DRDY, which goes low, when new conversion is ready, this interface should rely on that pin
type Ready interface {
	Open(callback func(any) error, args any) error
	Close()
}
