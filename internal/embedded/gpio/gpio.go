/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package gpio

import (
	"strconv"

	"github.com/warthog618/gpiod"
)

type ActiveLevel int
type Direction int

const (
	Low ActiveLevel = iota
	High
)
const (
	DirInput Direction = iota
	DirOutput
)

type Config struct {
	ID          string      `json:"id"`
	Direction   Direction   `json:"direction"`
	ActiveLevel ActiveLevel `json:"active_level"`
	Value       bool        `json:"value"`
}

type Pin struct {
	Chip string `json:"chip"`
	Line uint   `json:"line"`
}

type In struct {
	pin Pin
	Config
	level ActiveLevel
	*gpiod.Line
}

type Out struct {
	pin Pin
	Config
	level ActiveLevel
	*gpiod.Line
}

type Error struct {
	Pin Pin    `json:"pin"`
	Op  string `json:"op"`
	Err string `json:"error"`
}

func (e *Error) Error() string {
	if e.Err == "" {
		return "<nil>"
	}
	s := e.Op
	if e.Pin.Chip != "" {
		s += ":" + e.Pin.Chip
		s += ":" + strconv.FormatInt(int64(e.Pin.Line), 10)
	}
	s += ": " + e.Err
	return s
}

// init checks if there are any gpiochips available
func init() {
	chips := gpiod.Chips()
	if chips == nil {
		panic("gpiochips not found!")
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
var (
	_    Writer = (*Out)(nil)
	_, _ Reader = (*Out)(nil), (*In)(nil)
	_, _ Closer = (*Out)(nil), (*In)(nil)
)

func getLine(pin Pin, options ...gpiod.LineReqOption) (*gpiod.Line, error) {
	return gpiod.RequestLine(pin.Chip, int(pin.Line), options...)
}

func getActiveLevel(line *gpiod.Line) (ActiveLevel, error) {
	info, err := line.Info()
	if err != nil {
		return Low, err
	}

	if info.Config.ActiveLow {
		return Low, nil
	}
	return High, nil
}

func setActiveLevel(line *gpiod.Line, level ActiveLevel) error {
	opt := gpiod.AsActiveLow
	if level == High {
		opt = gpiod.AsActiveHigh
	}
	return line.Reconfigure(opt)
}

func Input(pin Pin, options ...gpiod.LineReqOption) (*In, error) {
	options = append(options, gpiod.AsInput)
	line, err := getLine(pin, options...)
	if err != nil {
		return nil, &Error{Pin: pin, Op: "input.getLine", Err: err.Error()}
	}
	lvl, err := getActiveLevel(line)
	if err != nil {
		return nil, &Error{Pin: pin, Op: "input.getActiveLevel", Err: err.Error()}
	}
	return &In{pin: pin, level: lvl, Line: line}, nil
}

func Output(pin Pin, initValue bool, options ...gpiod.LineReqOption) (*Out, error) {
	startValue := 0
	if initValue {
		startValue = 1
	}

	options = append(options, gpiod.AsOutput(startValue))
	line, err := getLine(pin, options...)
	if err != nil {
		return nil, &Error{Pin: pin, Op: "output.getLine", Err: err.Error()}
	}

	lvl, err := getActiveLevel(line)
	if err != nil {
		return nil, &Error{Pin: pin, Op: "output.getActiveLevel", Err: err.Error()}
	}
	return &Out{pin: pin, level: lvl, Line: line}, nil
}

func (in *In) ID() string {
	return in.Chip() + ":" + strconv.FormatInt(int64(in.Offset()), 32)
}

func (in *In) Get() (bool, error) {
	var value bool
	val, err := in.Line.Value()
	if err != nil {
		return value, &Error{Pin: in.pin, Op: "in.Get", Err: err.Error()}
	}

	if val > 0 {
		value = true
	}
	return value, err
}

func (in *In) Configure(new Config) error {
	last := in.Config
	if last.ActiveLevel != new.ActiveLevel {
		if err := setActiveLevel(in.Line, new.ActiveLevel); err != nil {
			return &Error{Pin: in.pin, Op: "Configure.setActiveLevel", Err: err.Error()}
		}
	}
	last.ActiveLevel = new.ActiveLevel
	return nil
}

func (in *In) GetConfig() (Config, error) {
	var err error
	in.Config.Value, err = in.Get()
	return in.Config, &Error{Pin: in.pin, Op: "GetConfig.Get", Err: err.Error()}
}

func (o *Out) Set(value bool) error {
	var setValue int
	if value {
		setValue = 1
	}
	err := o.SetValue(setValue)
	return &Error{Pin: o.pin, Op: "Set.SetValue", Err: err.Error()}

}

func (o *Out) Get() (bool, error) {
	var value bool
	val, err := o.Line.Value()
	if err != nil {
		return value, &Error{Pin: o.pin, Op: "Get", Err: err.Error()}
	}

	if val > 0 {
		value = true
	}

	return value, nil
}

func (o *Out) ID() string {
	return o.Chip() + ":" + strconv.FormatInt(int64(o.Offset()), 32)
}

func (o *Out) Configure(new Config) error {
	last := o.Config
	if last.ActiveLevel != new.ActiveLevel {
		if err := setActiveLevel(o.Line, new.ActiveLevel); err != nil {
			return &Error{Pin: o.pin, Op: "Configure.setActiveLevel", Err: err.Error()}
		}
	}
	last.ActiveLevel = new.ActiveLevel

	if last.Value != new.Value {
		if err := o.Set(new.Value); err != nil {
			return err
		}
	}
	return nil
}

func (o *Out) GetConfig() (Config, error) {
	var err error
	o.Config.Value, err = o.Get()
	return o.Config, err
}
