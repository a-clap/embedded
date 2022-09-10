package beaglebone

import "github.com/warthog618/gpiod"

// Interfaces from Gobot.gpio

// PwmWriter interface represents an Adaptor which has Pwm capabilities
type PwmWriter interface {
	PwmWrite(string, byte) (err error)
}

// ServoWriter interface represents an Adaptor which has Servo capabilities
type ServoWriter interface {
	ServoWrite(string, byte) (err error)
}

// DigitalWriter interface represents an Adaptor which has DigitalWrite capabilities
type DigitalWriter interface {
	DigitalWrite(string, byte) (err error)
}

// DigitalReader interface represents an Adaptor which has DigitalRead capabilities
type DigitalReader interface {
	DigitalRead(string) (val int, err error)
}

func NewPin() *digitalPin {
	p, err := gpiod.RequestLine("gpiochip0", 3)
	if err != nil {

	}
}
