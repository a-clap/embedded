package max31865

import (
	"errors"
	"fmt"
	"github.com/a-clap/logger"
	"math"
)

var Log logger.Logger = logger.NewNop()

type Transfer interface {
	TxRx(w []byte) (r []byte, err error)
}

var (
	ErrFault     = errors.New("sensor fault")
	ErrInterface = errors.New("interface fault")
)

type Device struct {
	t    Transfer
	cfg  Config
	regs regs
}

type Config struct {
	Bias       bool
	AutoMode   bool
	Wire3      bool
	Filter50Hz bool
	RefRes     int
	RNominal   int
	reg        byte
}

type regs struct {
	rtd      uint16
	faults   uint8
	rtdFault bool
}

func DefaultConfig() Config {
	return Config{
		Bias:       true,
		AutoMode:   true,
		Wire3:      true,
		Filter50Hz: false,
		RefRes:     430,
		RNominal:   100,
	}
}

func NewDefault(t Transfer) (*Device, error) {
	c := DefaultConfig()
	return New(t, c)
}

func New(t Transfer, config Config) (*Device, error) {
	d := &Device{
		t:    t,
		cfg:  config,
		regs: regs{},
	}

	if err := d.doConfig(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Device) Temperature() (float32, error) {
	if err := d.update(); err != nil {
		return 0, err
	}

	if d.regs.faults != 0 {
		d.clearFaults()
		return 0, ErrFault
	}

	return d.parseRtd(), nil
}

func (d *Device) doConfig() error {
	if err := d.update(); err != nil {
		return err
	}
	d.cfg.parse()
	return d.write(d.cfg.get())
}

func (d *Device) write(value byte) error {
	buf := []byte{uint8(0x80), value}
	_, err := d.t.TxRx(buf)
	if err != nil {
		return fmt.Errorf("%w %v", ErrInterface, err)
	}
	return nil
}

func (d *Device) read(addr byte, length int) ([]byte, error) {
	buf := make([]byte, length+1)
	buf[0] = addr
	r, err := d.t.TxRx(buf)
	if err != nil {
		return nil, fmt.Errorf("%w %v", ErrInterface, err)
	}
	return r[1:], nil
}

func (d *Device) clearFaults() {
	d.cfg.clearFaults()
	_ = d.write(d.cfg.get())
}

func (d *Device) update() error {
	const (
		conf         uint8 = 0x0
		rtdMsb             = 0x1
		rtdLsb             = 0x2
		faultHighMSB       = 0x3 // Not used
		faultHighLSB       = 0x4 // Not used
		faultLowMSB        = 0x5 // Not used
		faultLowLSB        = 0x6 // Not used
		faultStatus        = 0x7
	)
	// Reading all data, then map to regs
	// Supposedly it will be faster than multiple reads for single bytes
	r, err := d.read(conf, 8)
	if err != nil {
		return err
	}
	d.cfg.set(r[conf])
	// LSB of RTD is signification of fault
	if d.regs.rtdFault = (r[rtdLsb] >> 7) == 0x1; !d.regs.rtdFault {
		d.regs.rtd = (uint16(r[rtdMsb])<<8 | uint16(r[rtdLsb])) >> 1
	}
	d.regs.faults = r[faultStatus]
	Log.Info(d.regs)

	return nil
}

func (d *Device) parseRtd() float32 {
	const (
		RtdA float32 = 3.9083e-3
		RtdB float32 = -5.775e-7
	)
	refResistor := float32(d.cfg.RefRes)
	RTDnominal := float32(d.cfg.RNominal)

	Rt := float32(d.regs.rtd)
	Rt /= 32768
	Rt *= refResistor

	Z1 := -RtdA
	Z2 := RtdA*RtdA - (4 * RtdB)
	Z3 := (4 * RtdB) / RTDnominal
	Z4 := 2 * RtdB

	temp := Z2 + (Z3 * Rt)
	temp = (float32(math.Sqrt(float64(temp))) + Z1) / Z4

	if temp >= 0 {
		return temp
	}

	Rt /= RTDnominal
	Rt *= 100

	rpoly := Rt

	temp = -242.02
	temp += 2.2228 * rpoly
	rpoly *= Rt
	temp += 2.5859e-3 * rpoly
	rpoly *= RTDnominal
	temp -= 4.8260e-6 * rpoly
	rpoly *= Rt
	temp -= 2.8183e-8 * rpoly
	rpoly *= Rt
	temp += 1.5243e-10 * rpoly

	return temp
}
func (c *Config) parse() {
	c.setBias(c.Bias)
	c.setFilter(c.Filter50Hz)
	c.setMode(c.AutoMode)
	c.setWires(c.Wire3)
}

func (c *Config) setBias(bias bool) {
	const BiasPos = 7
	c.modify(bias, BiasPos)
}

func (c *Config) setMode(value bool) {
	const AutoMode = 6
	c.modify(value, AutoMode)
}

func (c *Config) setWires(value bool) {
	const Wires = 4
	c.modify(value, Wires)
}

func (c *Config) setFilter(value bool) {
	const Filter = 0
	c.modify(value, Filter)
}

func (c *Config) clearFaults() {
	const FaultClear = 1
	c.modify(true, FaultClear)
}

func (c *Config) modify(value bool, pos byte) {
	if value {
		c.reg |= 1 << pos
	} else {
		c.reg &= ^(1 << pos)
	}
}

func (c *Config) get() byte {
	return c.reg
}

func (c *Config) set(b byte) {
	c.reg = b

}
