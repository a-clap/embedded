package max

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

type Transfer interface {
	WriteRead(write []byte) (read []byte, err error)
}

type Dev struct {
	Transfer
}

func New(t Transfer) (*Dev, error) {
	return &Dev{t}, nil
}
