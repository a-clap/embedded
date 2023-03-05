/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

import (
	"errors"
	"fmt"
	"sync/atomic"
)

type Heater interface {
	ID() string             // ID returns unique ID of interface
	SetPower(pwr int) error // SetPower set power (in %) for that Heater. 0% means Heater should be disabled
}

type Sensor interface {
	ID() string           // ID returns unique ID of interface
	Temperature() float64 // Temperature returns latest temperature read from sensor
}

type Output interface {
	ID() string           // ID returns unique ID of interface
	Set(value bool) error // Set applies value to output
}

type Clock interface {
	Unix() int64 // Unix returns seconds since 01.01.1970 UTC
}

type Process struct {
	sensors            map[string]Sensor
	heaters            map[string]Heater
	outputs            map[string]Output
	clock              Clock
	config             Config
	currentPhaseConfig *PhaseConfig
	running            atomic.Bool
	status             Status
	moveToNext         moveToNext
}

const (
	initPhases         = 3
	initPhasesCapacity = 10
)

func New(option ...Option) (*Process, error) {
	p := &Process{
		sensors: map[string]Sensor{},
		heaters: map[string]Heater{},
		outputs: map[string]Output{},
		running: atomic.Bool{},
	}

	for _, opt := range option {
		opt(p)
	}

	if err := p.verify(); err != nil {
		return nil, err
	}

	p.config.Phases = make([]PhaseConfig, initPhases, initPhasesCapacity)
	p.setPhases(initPhases)
	return p, nil
}

// Run enable processing
func (p *Process) Run() (Status, error) {
	if p.running.Load() {
		return Status{}, ErrAlreadyRunning
	}

	// Check if our current config is correct, i.e. just after ctor
	// If there is no error, it will do unnecessary copy, like p.config = config
	if err := p.Configure(p.config); err != nil {
		return Status{}, err
	}

	p.run()
	return p.process()
}

func (p *Process) run() {
	p.running.Store(true)
	p.status.Done = false
	p.moveToPhase(0)
}

func (p *Process) finishRun() {
	p.running.Store(false)
	p.status.Done = true
}

func (p *Process) Process() (Status, error) {
	if !p.running.Load() {
		return Status{}, ErrNotRunning
	}
	return p.process()
}

func (p *Process) process() (Status, error) {
	// move to next
	if p.moveToNext.next() {
		p.moveToPhase(p.status.PhaseNumber + 1)
	}
	p.handleGpio()
	// Build status

	return Status{}, nil
}
func (p *Process) moveToPhase(next int) {
	if next >= p.config.PhaseNumber {
		// Looks like we are done
		p.finishRun()
		return
	}

	p.status.PhaseNumber = next
	p.currentPhaseConfig = &p.config.Phases[next]

	if p.currentPhaseConfig.Next.Type == ByTime {
		p.moveToNext = newByTime(p.clock, int64(p.currentPhaseConfig.Next.SecondsToMove))
	} else {
		// TODO: handle error
		s, _ := p.sensors[p.currentPhaseConfig.Next.SensorID]
		p.moveToNext = newByTemperature(p.clock, s, p.currentPhaseConfig.Next.SensorThreshold,
			int64(p.currentPhaseConfig.Next.TemperatureHoldSeconds))
	}
}

func (p *Process) handleGpio() {
	// Verify, if we still need to handle gpios

}

func (p *Process) GetConfig() Config {
	// Config holds probably more phaseConfigs than needed
	// Return only relevant
	cfg := Config{
		PhaseNumber: p.config.PhaseNumber,
		Phases:      p.config.Phases[:p.config.PhaseNumber],
	}
	return cfg
}

func (p *Process) Configure(cfg Config) error {
	if len(cfg.Phases) != cfg.PhaseNumber {
		return errors.New("PhaseNumber differs from length of Phases")
	}

	if err := validatePhaseNumber(cfg.PhaseNumber); err != nil {
		return err
	}

	for i, phaseConfig := range cfg.Phases {
		if err := p.validateConfig(i, phaseConfig); err != nil {
			return fmt.Errorf("validateConfig failed on %d: %w", i, err)
		}
	}
	// All should be good now
	p.setPhases(cfg.PhaseNumber)
	for i, phaseConfig := range cfg.Phases {
		p.configurePhase(i, phaseConfig)
	}

	return nil
}

func (p *Process) ConfigurePhase(phase int, config PhaseConfig) error {
	if err := p.validateConfig(phase, config); err != nil {
		return err
	}

	p.configurePhase(phase, config)
	return nil
}

func (p *Process) configurePhase(phase int, config PhaseConfig) {
	// TODO: what if configuring phase during Run?
	p.config.Phases[phase] = config
}

func (p *Process) SetPhases(number int) error {
	if err := validatePhaseNumber(number); err != nil {
		return err
	}
	p.setPhases(number)
	return nil
}

func (p *Process) setPhases(number int) {
	// Nothing to do
	if number == p.config.PhaseNumber {
		return
	}
	// Update value at the end
	defer func() {
		p.config.PhaseNumber = number
	}()

	// Do we need to add elements?
	if number > p.config.PhaseNumber {
		// Increase capacity and len - if needed
		for i := len(p.config.Phases); i < number; i++ {
			p.config.Phases = append(p.config.Phases, PhaseConfig{})
		}
	} else {
		// Remove unused elements over initPhasesCapacity
		if number < initPhasesCapacity {
			p.config.Phases = p.config.Phases[:initPhasesCapacity]
		}
	}
}

func (p *Process) validateConfig(phase int, config PhaseConfig) error {
	if phase >= p.config.PhaseNumber {
		return ErrNoSuchPhase
	}

	if err := p.validateConfigHeaters(config.Heaters); err != nil {
		return err
	}

	if err := p.validateNext(config.Next); err != nil {
		return err
	}

	return nil
}

// validateNext check values in passed MoveToNextConfig
func (p *Process) validateNext(next MoveToNextConfig) error {
	switch next.Type {
	case ByTime:
		if next.SecondsToMove <= 0 {
			return ErrByTimeWrongTime
		}
	case ByTemperature:
		if _, ok := p.sensors[next.SensorID]; !ok {
			return ErrByTemperatureWrongID
		}
	default:
		return ErrUnknownType
	}

	return nil
}

// validateConfigHeaters check values in passed HeaterPhaseConfig
func (p *Process) validateConfigHeaters(heaters []HeaterPhaseConfig) error {
	if len(heaters) == 0 {
		return ErrNoHeatersInConfig
	}

	for _, heaterConfig := range heaters {
		if _, ok := p.heaters[heaterConfig.ID]; !ok {
			return fmt.Errorf("%w. ID: %v", ErrWrongHeaterID, heaterConfig.ID)
		}
		if heaterConfig.Power > 100 || heaterConfig.Power < 0 {
			return fmt.Errorf("%w: ID: %v, Value: %v", ErrWrongHeaterPower, heaterConfig.ID, heaterConfig.Power)
		}
	}
	return nil
}

func validatePhaseNumber(number int) error {
	if number <= 0 {
		return ErrPhasesBelowZero
	}
	return nil
}

func (p *Process) verify() error {
	if p.clock == nil {
		// that's ok - just use std
		p.clock = new(clock)
	}

	if len(p.sensors) == 0 {
		return ErrNoTSensors
	}

	if len(p.heaters) == 0 {
		return ErrNoHeaters
	}

	return nil
}
