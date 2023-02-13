/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

type Wiring string

const (
	TwoWire   Wiring = "twoWire"
	ThreeWire Wiring = "threeWire"
	FourWire  Wiring = "fourWire"
)

const (
	filter60Hz uint8 = iota
	clearFault
	faultDetect1
	faultDetect2
	wire3
	oneShot
	continuous
	vBias
)

type configReg struct {
	id       string
	wiring   Wiring
	refRes   float32
	rNominal float32
}

func newConfig() *configReg {
	return &configReg{
		id:       "",
		wiring:   ThreeWire,
		refRes:   430.0,
		rNominal: 100.0,
	}
}

func (c *configReg) reg() uint8 {
	const value = uint8((1 << filter60Hz) | (1 << continuous) | (1 << vBias))
	const wireMsk = 1 << wire3
	if c.wiring == ThreeWire {
		return value | wireMsk
	}
	return value
}

func (c *configReg) clearFaults() uint8 {
	return c.reg() | (1 << clearFault)
}

func (c *configReg) faultDetect() uint8 {
	return 0b10000100 | (c.reg() & ((1 << filter60Hz) | (1 << wire3)))
}

func (c *configReg) faultDetectFinished(reg uint8) bool {
	mask := uint8(1<<faultDetect2 | 1<<faultDetect1)
	return reg&mask == 0
}
