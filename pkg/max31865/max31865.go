package max31865

import (
	"errors"
	"fmt"
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
	ErrReadWrite        = errors.New("error on ReadWrite")
	ErrReadZeroes       = errors.New("read only zeroes from device")
	ErrReadFF           = errors.New("read only 0xFF from device")
	ErrRtd              = errors.New("rtd error")
	ErrAlreadyPolling   = errors.New("sensor is already polling")
	ErrWrongArgs        = errors.New("wrong args passed to callback")
	ErrNoReadyInterface = errors.New("lack of ready interface")
)

type Transfer interface {
	io.Closer
	ReadWrite(write []byte) (read []byte, err error)
}

type Sensor interface {
	io.Closer
	ID() string
	Temperature() (float32, error)
}

// Ready is an interface which allows to register a callback
// max31865 has a pin DRDY, which goes low, when new conversion is ready, this interface should rely on that pin
type Ready interface {
	Open(callback func(any) error, args any) error
	Close()
}

type Readings interface {
	Get() (id string, temperature string, timestamp time.Time)
}

type sensor struct {
	Transfer
	cfg    config
	regCfg *regConfig
	r      *rtd
}

func NewDefault(devFile string, args ...any) (Sensor, error) {
	dev, err := newMaxSpidev(devFile)
	if err != nil {
		return nil, err
	}
	args = append([]any{ID(devFile)}, args...)
	return New(dev, args...)
}

func New(t Transfer, args ...any) (Sensor, error) {
	if err := checkTransfer(t); err != nil {
		return nil, err
	}

	s := &sensor{
		Transfer: t,
		regCfg:   newRegConfig(),
		r:        newRtd(),
		cfg:      newConfig(),
	}

	s.parse(args...)
	// Do initial regConfig
	err := s.config()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *sensor) parse(args ...any) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case Ready:
			s.cfg.ready = arg
		case ID:
			s.cfg.id = arg
		case Wiring:
			s.cfg.wiring = arg
			s.regCfg.setWiring(arg)
		case RefRes:
			s.cfg.refRes = arg
		case RNominal:
			s.cfg.rNominal = arg
		}
	}

}

func (s *sensor) ID() string {
	return string(s.cfg.id)
}

func (s *sensor) Temperature() (tmp float32, err error) {
	r, err := s.read(regConf, regFault+1)
	if err != nil {
		//	can't do much about it
		return
	}
	err = s.r.update(r[regRtdMsb], r[regRtdLsb])
	if err != nil {
		// Not handling error here, should have happened on previous call
		_ = s.clearFaults()
		// make error more specific
		err = fmt.Errorf("%w: errorReg: %v, posibble causes: %v", err, r[regFault], errorCauses(r[regFault], s.cfg.wiring))
		return
	}
	rtd := s.r.rtd()
	return rtdToTemperature(rtd, s.cfg.refRes, s.cfg.rNominal), nil
}

func (s *sensor) Close() error {
	return s.Transfer.Close()
}

func (s *sensor) clearFaults() error {
	return s.write(regConf, []byte{s.regCfg.clearFaults()})
}

func (s *sensor) config() error {
	err := s.write(regConf, []byte{s.regCfg.reg()})
	return err
}

func (s *sensor) read(addr byte, len int) ([]byte, error) {
	// We need to create slice with 1 byte more
	w := make([]byte, len+1)
	w[0] = addr
	r, err := s.ReadWrite(w)
	if err != nil {
		return nil, err
	}
	// First byte is useless
	return r[1:], nil
}

func (s *sensor) write(addr byte, w []byte) error {
	buf := []byte{addr | 0x80}
	buf = append(buf, w...)
	_, err := s.ReadWrite(buf)
	return err
}

func checkTransfer(t Transfer) error {
	const size = regFault + 1
	buf := make([]byte, size)
	buf[0] = regConf
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

func rtdToTemperature(rtd uint16, refResT RefRes, rNominalT RNominal) float32 {
	refRes := float32(refResT)
	rNominal := float32(rNominalT)
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
