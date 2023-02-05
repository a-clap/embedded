/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package gpio

import (
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/warthog618/gpiod"
	"strconv"
)

type Pin struct {
	Chip string `json:"chip"`
	Line uint   `json:"line"`
}

type In struct {
	level models.ActiveLevel
	*gpiod.Line
}

type Out struct {
	level models.ActiveLevel
	*gpiod.Line
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

	_, _ models.GPIO = (*In)(nil), (*Out)(nil)
)

func getLine(pin Pin, options ...gpiod.LineReqOption) (*gpiod.Line, error) {
	return gpiod.RequestLine(pin.Chip, int(pin.Line), options...)
}

func getActiveLevel(line *gpiod.Line) (models.ActiveLevel, error) {
	info, err := line.Info()
	if err != nil {
		return models.Low, err
	}

	if info.Config.ActiveLow {
		return models.Low, nil
	}
	return models.High, nil
}
func setActiveLevel(line *gpiod.Line, level models.ActiveLevel) error {
	opt := gpiod.AsActiveLow
	if level == models.High {
		opt = gpiod.AsActiveHigh
	}
	return line.Reconfigure(opt)
}

func Input(pin Pin, options ...gpiod.LineReqOption) (*In, error) {
	options = append(options, gpiod.AsInput)
	line, err := getLine(pin, options...)
	if err != nil {
		return nil, err
	}
	lvl, err := getActiveLevel(line)
	if err != nil {
		return nil, err
	}
	return &In{level: lvl, Line: line}, nil
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
	lvl, err := getActiveLevel(line)
	if err != nil {
		return nil, err
	}
	return &Out{level: lvl, Line: line}, nil
}

func (in *In) Get() (bool, error) {
	var value bool
	val, err := in.Value()
	if val > 0 {
		value = true
	}

	return value, err
}

func (in *In) ID() string {
	return in.Chip() + ":" + strconv.FormatInt(int64(in.Offset()), 32)
}

func (in *In) Config() (models.GPIOConfig, error) {
	val, err := in.Get()
	if err != nil {
		return models.GPIOConfig{}, nil
	}
	return models.GPIOConfig{
		ID:          in.ID(),
		Direction:   models.Input,
		ActiveLevel: in.level,
		Value:       val,
	}, nil
}

func (in *In) SetConfig(cfg models.GPIOConfig) error {
	if cfg.ActiveLevel != in.level {
		if err := setActiveLevel(in.Line, cfg.ActiveLevel); err != nil {
			return err
		}
		in.level = cfg.ActiveLevel
	}
	return nil
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

func (o *Out) ID() string {
	return o.Chip() + ":" + strconv.FormatInt(int64(o.Offset()), 32)
}

func (o *Out) Config() (models.GPIOConfig, error) {
	val, err := o.Get()
	if err != nil {
		return models.GPIOConfig{}, nil
	}
	return models.GPIOConfig{
		ID:          o.ID(),
		Direction:   models.Input,
		ActiveLevel: o.level,
		Value:       val,
	}, nil
}

func (o *Out) SetConfig(cfg models.GPIOConfig) error {
	if cfg.ActiveLevel != o.level {
		if err := setActiveLevel(o.Line, cfg.ActiveLevel); err != nil {
			return err
		}
		o.level = cfg.ActiveLevel
	}

	value, err := o.Get()
	if err != nil {
		return err
	}

	if cfg.Value != value {
		return o.Set(cfg.Value)
	}

	return nil
}
