package max

import "fmt"

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
	ReadWrite(write []byte) (read []byte, err error)
}

type Dev struct {
	Transfer
	c *config
}

func New(t Transfer, wiring Wiring) (*Dev, error) {
	if err := checkTransfer(t); err != nil {
		return nil, err
	}
	return &Dev{Transfer: t, c: newConfig(wiring)}, nil
}

func (d *Dev) write(addr byte, w []byte) error {
	buf := []byte{addr | 0x80}
	buf = append(buf, w...)
	_, err := d.ReadWrite(buf)
	return err
}

func (d *Dev) read(addr byte, len int) ([]byte, error) {
	// We need to create slice with 1 byte more
	w := make([]byte, 0, len+1)
	w[0] = addr
	r, err := d.ReadWrite(w)
	if err != nil {
		return nil, err
	}
	// First byte is useless
	return r[1:], nil
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
