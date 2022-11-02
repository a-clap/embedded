package max31865

import (
	"fmt"
	"github.com/a-clap/iot/pkg/spidev"
	"io"
	"math"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

const (
	REG_CONF = iota
	REG_RTD_MSB
	REG_RTD_LSB
	REG_H_FAULT_MSB
	REG_H_FAULT_LSB
	REG_L_FAULT_MSB
	REG_L_FAULT_LSB
	REG_FAULT
)

var (
	ErrReadWrite  = fmt.Errorf("error on ReadWrite")
	ErrReadZeroes = fmt.Errorf("read only zeroes from device")
	ErrReadFF     = fmt.Errorf("read only 0xFF from device")
	ErrRtd        = fmt.Errorf("rtd error")
)

type Transfer interface {
	io.Closer
	ReadWrite(write []byte) (read []byte, err error)
}

type maxSpidevTransfer struct {
	*spidev.Spidev
}

func (m maxSpidevTransfer) ReadWrite(write []byte) (read []byte, err error) {
	read = make([]byte, len(write))
	err = m.Spidev.Tx(write, read)
	return read, err
}

type Dev struct {
	Transfer
	io.Closer
	cfg Config
	c   *regConfig
	r   *rtd
}

type Config struct {
	Wiring   Wiring
	RefRes   float32
	RNominal float32
}

func NewDefault(devFile string, c Config) (*Dev, error) {
	t, err := spidev.New(devFile, 5*physic.MegaHertz, spi.Mode1, 8)
	if err != nil {
		return nil, err
	}
	return New(&maxSpidevTransfer{t}, c)
}

func New(t Transfer, c Config) (*Dev, error) {
	if err := checkTransfer(t); err != nil {
		return nil, err
	}
	d := &Dev{
		Transfer: t,
		c:        newConfig(c.Wiring),
		r:        newRtd(),
		cfg:      c,
	}
	// Do initial regConfig
	err := d.config()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Dev) Temperature() (tmp float32, err error) {
	r, err := d.read(REG_CONF, REG_FAULT+1)
	if err != nil {
		//	can't do much about it
		return
	}
	err = d.r.update(r[REG_RTD_MSB], r[REG_RTD_LSB])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = d.clearFaults()
		// make error more specific
		err = fmt.Errorf("%w: errorReg: %v, posibble causes: %v", err, r[REG_FAULT], errorCauses(r[REG_FAULT], d.cfg.Wiring))
		return
	}
	rtd := d.r.rtd()
	return rtdToTemperature(rtd, d.cfg.RefRes, d.cfg.RNominal), nil
}

func (d *Dev) clearFaults() error {
	return d.write(REG_CONF, []byte{d.c.clearFaults()})
}

func (d *Dev) config() error {
	err := d.write(REG_CONF, []byte{d.c.reg()})
	return err
}

func (d *Dev) read(addr byte, len int) ([]byte, error) {
	// We need to create slice with 1 byte more
	w := make([]byte, len+1)
	w[0] = addr
	r, err := d.ReadWrite(w)
	if err != nil {
		return nil, err
	}
	// First byte is useless
	return r[1:], nil
}

func (d *Dev) write(addr byte, w []byte) error {
	buf := []byte{addr | 0x80}
	buf = append(buf, w...)
	_, err := d.ReadWrite(buf)
	return err
}

func checkTransfer(t Transfer) error {
	const size = REG_FAULT + 1
	buf := make([]byte, size)
	buf[0] = REG_CONF
	r, err := t.ReadWrite(buf)
	if err != nil {
		return ErrReadWrite
	}
	checkReadings := func(expected byte) bool {
		for _, elem := range r {
			if elem != expected {
				return false
			}
		}
		return true
	}

	if onlyZeroes := checkReadings(0); onlyZeroes {
		return ErrReadZeroes
	}

	if onlyFF := checkReadings(0xff); onlyFF {
		return ErrReadFF
	}
	return nil
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

func (d *Dev) Close() error {
	return d.Transfer.Close()
}
