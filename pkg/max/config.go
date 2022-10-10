package max

type Wiring string

const (
	TwoWire   Wiring = "twoWire"
	ThreeWire        = "threeWire"
	FourWire         = "fourWire"
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

type Config struct {
	Wiring Wiring
	value  uint8
}

func NewConfig(w Wiring) *Config {
	// Default values
	value := uint8((1 << filter60Hz) | (1 << continuous) | (1 << vBias))

	if w == ThreeWire {
		value |= 1 << wire3
	}
	return &Config{
		Wiring: w,
		value:  value,
	}
}

func (c *Config) reg() uint8 {
	return c.value
}

func (c *Config) clearFaults() uint8 {
	return c.reg() | (1 << clearFault)
}

func (c *Config) faultDetect() uint8 {
	return 0b10000100 | (c.reg() & ((1 << filter60Hz) | (1 << wire3)))
}

func (c *Config) faultDetectFinished(reg uint8) bool {
	mask := uint8(1<<faultDetect2 | 1<<faultDetect1)
	return reg&mask == 0
}
