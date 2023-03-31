/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

type Wiring int

// Wiring options
const (
	TwoWire Wiring = iota + 2
	ThreeWire
	FourWire
)

// Registers in Max31865
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

// configReg holds most important Max values - wiring and re
type configReg struct {
	wiring     Wiring  // chosen by user
	refRes     float64 // reference resistor (usually 400 or 430)
	nominalRes float64 // nominal resistor (100 for PT100)
}

func newConfig() *configReg {
	return &configReg{
		wiring:     ThreeWire,
		refRes:     430.0,
		nominalRes: 100.0,
	}
}

// reg returns current value of register - based on config
func (c *configReg) reg() uint8 {
	const value = uint8((1 << filter60Hz) | (1 << continuous) | (1 << vBias))
	const wireMsk = 1 << wire3
	if c.wiring == ThreeWire {
		return value | wireMsk
	}
	return value
}

// clearFaults returns command which execute internal Max command - faults reset
func (c *configReg) clearFaults() uint8 {
	return c.reg() | (1 << clearFault)
}

// faultDetect returns command which executes internal Max command - fault detect
func (c *configReg) faultDetect() uint8 {
	return 0b10000100 | (c.reg() & ((1 << filter60Hz) | (1 << wire3)))
}

// faultDetectFinished returns whether faultDetect is done
func (c *configReg) faultDetectFinished(reg uint8) bool {
	mask := uint8(1<<faultDetect2 | 1<<faultDetect1)
	return reg&mask == 0
}
