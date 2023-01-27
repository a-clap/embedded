package max31865

import "sync/atomic"

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

type pollType int

const (
	sync pollType = iota
	async
)

type config struct {
	id       string
	wiring   Wiring
	refRes   float32
	rNominal float32
	ready    Ready
	polling  atomic.Bool
	pollType pollType
}

func newConfig() *config {
	return &config{
		id:       "",
		wiring:   ThreeWire,
		refRes:   430.0,
		rNominal: 100.0,
		ready:    nil,
		polling:  atomic.Bool{},
		pollType: sync,
	}
}

func (c *config) reg() uint8 {
	const value = uint8((1 << filter60Hz) | (1 << continuous) | (1 << vBias))
	const wireMsk = 1 << wire3
	if c.wiring == ThreeWire {
		return value | wireMsk
	}
	return value
}

func (c *config) clearFaults() uint8 {
	return c.reg() | (1 << clearFault)
}

func (c *config) faultDetect() uint8 {
	return 0b10000100 | (c.reg() & ((1 << filter60Hz) | (1 << wire3)))
}

func (c *config) faultDetectFinished(reg uint8) bool {
	mask := uint8(1<<faultDetect2 | 1<<faultDetect1)
	return reg&mask == 0
}
