/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import "math"

type rtd struct {
	r   uint16
	err error
}

func newRtd() *rtd {
	return &rtd{}
}

func (r *rtd) update(msb byte, lsb byte) error {
	// first bit in lsb is information about error
	if lsb&0x1 == 0x1 {
		r.err = ErrRtd
		return r.err
	}
	r.r = uint16(msb)<<8 | uint16(lsb)
	// rtd need to be shifted
	r.r >>= 1
	r.err = nil

	return nil
}

func (r *rtd) rtd() uint16 {
	return r.r
}

func (r *rtd) toTemperature(refRes float32, rNominal float32) float32 {
	const (
		RtdA float32 = 3.9083e-3
		RtdB float32 = -5.775e-7
	)
	Rt := float32(r.r)
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
