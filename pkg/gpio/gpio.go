package gpio

import (
	"github.com/a-clap/logger"
	"github.com/warthog618/gpiod"
	"strconv"
)

var (
	pins int = 0 // max
)

// init read gpiochips from user space, so we can set detect how many gpios are available
func init() {
	chips := gpiod.Chips()
	if chips == nil {
		panic("gpiochips not found!")
	}

	Log.Debug("detected chips: ", chips)
	for _, c := range chips {
		chip, err := gpiod.NewChip(c)
		if err != nil {
			Log.Error("error getting chip ", chip, ", err: ", err)
			continue
		}
		// add lines to max pins
		pins += chip.Lines()

		if err := chip.Close(); err != nil {
			Log.Errorf("error on closing chip %v: %v", chip, err)
		}
	}
}

var Log logger.Logger = logger.NewNop()

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

type In struct {
	*gpiod.Line
}

type Out struct {
	*gpiod.Line
}

func getLine(pin int, options ...gpiod.LineReqOption) (*gpiod.Line, error) {
	chip, pin := parsePin(pin)
	return gpiod.RequestLine(chip, pin, options...)
}

func Input(pin int, options ...gpiod.LineReqOption) (*In, error) {
	options = append(options, gpiod.AsInput)
	line, err := getLine(pin, options...)
	if err != nil {
		return nil, err
	}
	return &In{Line: line}, nil
}

func Output(pin int, initValue bool, options ...gpiod.LineReqOption) (*Out, error) {
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

func parsePin(pin int) (chip string, pinNumber int) {
	// Each gpiochip has at most 32 pins (0 - 31)
	chip = "gpiochip" + strconv.FormatInt(int64(pin/32), 10)

	pinNumber = pin % 32
	Log.Debugf("for pin %v, chip: %v, pinNumber: %v", pin, chip, pinNumber)
	return
}
