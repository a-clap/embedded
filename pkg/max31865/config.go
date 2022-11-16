package max31865

import "sync/atomic"

type Wiring string
type RefRes float32
type RNominal float32
type ID string

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

type pollType int

const (
	sync pollType = iota
	async
)

type config struct {
	id       ID
	wiring   Wiring
	refRes   RefRes
	rNominal RNominal
	ready    Ready
	polling  atomic.Bool
	pollType pollType
}

type regConfig struct {
	value uint8
}

func newConfig() config {
	return config{
		id:       "",
		wiring:   ThreeWire,
		refRes:   430.0,
		rNominal: 100.0,
		ready:    nil,
		polling:  atomic.Bool{},
		pollType: sync,
	}
}

func newRegConfig() *regConfig {
	// Default values
	value := uint8((1 << filter60Hz) | (1 << continuous) | (1 << vBias))

	c := &regConfig{
		value: value,
	}
	c.setWiring(ThreeWire)
	return c
}
func (c *regConfig) setWiring(w Wiring) {
	const wireMsk = 1 << wire3
	if w == ThreeWire {
		c.value |= wireMsk
	} else {
		c.value &^= wireMsk
	}
}

func (c *regConfig) reg() uint8 {
	return c.value
}

func (c *regConfig) clearFaults() uint8 {
	return c.reg() | (1 << clearFault)
}

func (c *regConfig) faultDetect() uint8 {
	return 0b10000100 | (c.reg() & ((1 << filter60Hz) | (1 << wire3)))
}

func (c *regConfig) faultDetectFinished(reg uint8) bool {
	mask := uint8(1<<faultDetect2 | 1<<faultDetect1)
	return reg&mask == 0
}
