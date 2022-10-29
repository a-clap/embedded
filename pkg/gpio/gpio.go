package gpio

import (
	"github.com/warthog618/gpiod"
)

type Pin struct {
	Chip string
	Line uint
}

type In struct {
	*gpiod.Line
}

type Out struct {
	*gpiod.Line
}

// init read gpiochips from user space, so we can set detect how many gpios are available
func init() {
	chips := gpiod.Chips()
	if chips == nil {
		panic("gpiochips not found!")
	}

	for _, c := range chips {
		chip, err := gpiod.NewChip(c)
		if err != nil {
			continue
		}
		_ = chip.Close()
	}
}

// Writer provides access to set value on digital output
type Writer interface {
	Set(bool) error
}

// Reader returns current value of gpio, input or output
type Reader interface {
	Get() (bool, error)
}

// Closer closes line
type Closer interface {
	Close() error
}

// Fulfill interfaces
var _ Writer = &Out{}
var _, _ Reader = &Out{}, &In{}
var _, _ Closer = &Out{}, &In{}

func getLine(pin Pin, options ...gpiod.LineReqOption) (*gpiod.Line, error) {
	return gpiod.RequestLine(pin.Chip, int(pin.Line), options...)
}

func Input(pin Pin, options ...gpiod.LineReqOption) (*In, error) {
	options = append(options, gpiod.AsInput)
	line, err := getLine(pin, options...)
	if err != nil {
		return nil, err
	}
	return &In{Line: line}, nil
}

func Output(pin Pin, initValue bool, options ...gpiod.LineReqOption) (*Out, error) {
	startValue := 0
	if initValue {
		startValue = 1
	}
	options = append(options, gpiod.AsOutput(startValue))
	line, err := getLine(pin, options...)
	if err != nil {
		return nil, err
	}
	return &Out{Line: line}, nil
}

func (o *Out) Set(value bool) error {
	var setValue int
	if value {
		setValue = 1
	}
	return o.SetValue(setValue)
}

func (o *Out) Get() (bool, error) {
	var value bool
	val, err := o.Value()
	if val > 0 {
		value = true
	}

	return value, err
}

func (in *In) Get() (bool, error) {
	var value bool
	val, err := in.Value()
	if val > 0 {
		value = true
	}

	return value, err
}
