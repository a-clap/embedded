/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"errors"
	"io"
	"math"
	"time"
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

var (
	ErrInterface         = errors.New("error on interface usage")
	ErrReadZeroes        = errors.New("read only zeroes from device")
	ErrReadFF            = errors.New("read only 0xFF from device")
	ErrRtd               = errors.New("rtd error")
	ErrAlreadyPolling    = errors.New("max is already polling")
	ErrNotPolling        = errors.New("max is not polling")
	ErrWrongArgs         = errors.New("wrong args passed to callback")
	ErrNoReadyInterface  = errors.New("lack of Ready interface")
	ErrNoReadWriteCloser = errors.New("lack of ReadWriterCloser interface")
	ErrTooMuchTriggers   = errors.New("poll received too much triggers")
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

type Readings interface {
	ID() string
	Temperature() float32
	Stamp() time.Time
	Error() error
}

func New(options ...Option) (*Sensor, error) {
	return newSensor(options...)
}

func rtdToTemperature(rtd uint16, refRes float32, rNominal float32) float32 {
	const (
		RtdA float32 = 3.9083e-3
		RtdB float32 = -5.775e-7
	)
	Rt := float32(rtd)
	Rt /= 32768
	Rt *= refRes

	Z1 := -RtdA
	Z2 := RtdA*RtdA - (4 * RtdB)
	Z3 := (4 * RtdB) / rNominal
	Z4 := 2 * RtdB

	temp := Z2 + (Z3 * Rt)
	temp = (float32(math.Sqrt(float64(temp))) + Z1) / Z4

	if temp >= 0 {
		return temp
	}

	Rt /= rNominal
	Rt *= 100

	rpoly := Rt

	temp = -242.02
	temp += 2.2228 * rpoly
	rpoly *= Rt
	temp += 2.5859e-3 * rpoly
	rpoly *= rNominal
	temp -= 4.8260e-6 * rpoly
	rpoly *= Rt
	temp -= 2.8183e-8 * rpoly
	rpoly *= Rt
	temp += 1.5243e-10 * rpoly

	return temp
}
