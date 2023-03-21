/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
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

type output struct {
	Output  Output
	inRange bool
}

type Process struct {
	sensors            map[string]Sensor
	heaters            map[string]Heater
	outputs            map[string]*output
	clock              Clock
	config             Config
	currentPhaseConfig *PhaseConfig
	currentStamp       int64
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
		outputs: map[string]*output{},
		running: atomic.Bool{},
	}

	for _, opt := range option {
		opt(p)
	}

	p.config.Phases = make([]PhaseConfig, initPhases, initPhasesCapacity)
	p.setPhases(initPhases)
	return p, nil
}

func (p *Process) ConfigureSensors(sensors []Sensor) {
	p.sensors = make(map[string]Sensor)
	for _, sensor := range sensors {
		p.sensors[sensor.ID()] = sensor
	}
}
func (p *Process) ConfigureHeaters(sensors []Heater) {
	p.heaters = make(map[string]Heater)
	for _, heater := range sensors {
		p.heaters[heater.ID()] = heater
	}
}
func (p *Process) ConfigureOutputs(outputs []Output) {
	p.outputs = make(map[string]*output)
	for _, o := range outputs {
		p.outputs[o.ID()] = &output{
			Output:  o,
			inRange: false,
		}
	}
}

// Run enable processing
func (p *Process) Run() (Status, error) {
	if err := p.Validate(); err != nil {
		return Status{}, err
	}

	if p.Running() {
		return Status{}, ErrAlreadyRunning
	}

	p.run()
	return p.process()
}

func (p *Process) Running() bool {
	return p.running.Load()
}

func (p *Process) run() {
	p.running.Store(true)
	p.status.Running = true
	p.status.Done = false
	p.currentStamp = p.clock.Unix()
	p.status.StartTime = time.Unix(p.currentStamp, 0)

	p.moveToPhase(0)
}

func (p *Process) finishRun() {
	p.status.Done = true
	p.status.Running = false
	p.status.EndTime = time.Unix(p.currentStamp, 0)
	p.running.Store(false)

	// build standard status, but because of finished Process it will disable all outputs and heaters
	p.handleTemperatures()
	p.handleGpio()
	p.handleHeaters()
	if p.currentPhaseConfig.Next.Type == ByTime {
		p.status.Next.Time.TimeLeft = 0
	} else {
		p.status.Next.Temperature.TimeLeft = 0
	}

}

func (p *Process) Process() (Status, error) {
	if p.Running() == false {
		return Status{Running: false}, ErrNotRunning
	}
	p.currentStamp = p.clock.Unix()
	return p.process()
}

func (p *Process) Configure(cfg Config) error {
	if cfg.PhaseNumber != len(cfg.Phases) {
		// Major verification, rest is validated in Validate
		return errors.New("PhaseNumber differs from length of Phases")
	}

	p.setPhases(cfg.PhaseNumber)

	for i, phaseConfig := range cfg.Phases {
		p.configurePhase(i, phaseConfig)
	}

	return nil
}

func (p *Process) ConfigurePhase(phase int, config PhaseConfig) error {
	if phase >= p.config.PhaseNumber {
		return ErrNoSuchPhase
	}

	if p.running.Load() && p.status.PhaseNumber == phase {
		// If we are during process and configuring current phase we need to check config
		if err := p.validatePhaseConfig(phase, config); err != nil {
			return err
		}
	}

	p.configurePhase(phase, config)
	return nil
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
func (p *Process) SetPhases(number int) error {
	if err := validatePhaseNumber(number); err != nil {
		return err
	}
	p.setPhases(number)
	return nil
}

// Validate checks if needed interfaces are fulfilled and config is correct
func (p *Process) Validate() error {
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

	if len(p.config.Phases) != p.config.PhaseNumber {
		return errors.New("PhaseNumber differs from length of Phases")
	}

	for _, cfg := range p.config.Phases {
		// Validate, if len of each []Config match number of objects
		if len(cfg.Heaters) != len(p.heaters) {
			return ErrHeaterConfigDiffersFromHeatersLen
		}
		if len(cfg.GPIO) != len(p.outputs) {
			return ErrDifferentGPIOSConfig
		}
	}

	if err := validatePhaseNumber(p.config.PhaseNumber); err != nil {
		return err
	}

	for i, phaseConfig := range p.config.Phases {
		if err := p.validatePhaseConfig(i, phaseConfig); err != nil {
			return fmt.Errorf("validatePhaseConfig failed on %d: %w", i, err)
		}
	}

	return nil
}

func (p *Process) process() (Status, error) {
	// Handle each component and build status inside callbacks
	// Order is important, as GPIO may use temperature from updated Status
	p.status.Errors = nil
	if p.moveToNext.next(p.currentStamp) {
		p.moveToPhase(p.status.PhaseNumber + 1)
	} else {
		p.handleTemperatures()
		p.handleHeaters()
		p.handleGpio()

		timeleft := p.moveToNext.timeleft(p.currentStamp)
		if p.currentPhaseConfig.Next.Type == ByTime {
			p.status.Next.Time.TimeLeft = timeleft
		} else {
			p.status.Next.Temperature.TimeLeft = timeleft
		}

	}

	return p.status, nil
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
		p.status.Next.Temperature = MoveToNextStatusTemperature{}
		p.status.Next.Type = ByTime
		p.status.Next.Time.TimeLeft = p.currentPhaseConfig.Next.SecondsToMove

		p.moveToNext = newByTime(p.currentStamp, p.currentPhaseConfig.Next.SecondsToMove)
	} else {
		p.status.Next.Time = MoveToNextStatusTime{}
		p.status.Next.Type = ByTemperature
		p.status.Next.Temperature.SensorID = p.currentPhaseConfig.Next.SensorID
		p.status.Next.Temperature.TimeLeft = p.currentPhaseConfig.Next.TemperatureHoldSeconds
		p.status.Next.Temperature.SensorThreshold = p.currentPhaseConfig.Next.SensorThreshold
		// It is not possible - as ID are verified during Configure
		s, _ := p.sensors[p.currentPhaseConfig.Next.SensorID]
		p.moveToNext = newByTemperature(p.currentStamp, s, p.currentPhaseConfig.Next.SensorThreshold,
			p.currentPhaseConfig.Next.TemperatureHoldSeconds)
	}
	// If we moved, we need to update
	// Order is important, as GPIO may use temperature from updated Status
	p.handleTemperatures()
	p.handleHeaters()
	p.handleGpio()
}

func (p *Process) handleTemperatures() {
	p.status.Temperature = make([]TemperaturePhaseStatus, 0, len(p.sensors))
	for id, sensor := range p.sensors {
		p.status.Temperature = append(p.status.Temperature, TemperaturePhaseStatus{
			ID:          id,
			Temperature: sensor.Temperature(),
		})
	}
}

func (p *Process) handleHeaters() {
	p.status.Heaters = make([]HeaterPhaseStatus, 0, len(p.currentPhaseConfig.Heaters))
	for _, config := range p.currentPhaseConfig.Heaters {
		heater := p.heaters[config.ID]
		pwr := 0

		if p.status.Running {
			pwr = config.Power
		}

		if err := heater.SetPower(pwr); err != nil {
			err = fmt.Errorf("%w: on ID: %v, SetPower: %v", err, config.ID, config.Power)
			p.status.Errors = append(p.status.Errors, err.Error())
			continue
		}
		p.status.Heaters = append(p.status.Heaters, HeaterPhaseStatus{HeaterPhaseConfig{
			ID:    config.ID,
			Power: pwr,
		}})
	}
}

func (p *Process) handleGpio() {
	p.status.GPIO = make([]GPIOPhaseStatus, 0, len(p.currentPhaseConfig.GPIO))
	for _, config := range p.currentPhaseConfig.GPIO {
		gpio, ok := p.outputs[config.ID]
		if !ok {
			err := fmt.Errorf("output with ID: %v not found", config.ID)
			p.status.Errors = append(p.status.Errors, err.Error())
			continue
		}
		gpioValue := false
		if p.status.Running {
			s, ok := p.sensors[config.SensorID]
			if !ok {
				err := fmt.Errorf("sensor with ID: %v not found", config.SensorID)
				p.status.Errors = append(p.status.Errors, err.Error())
				continue
			}

			t := s.Temperature()
			if gpio.inRange {
				// Last time was in range
				gpio.inRange = t >= (config.TLow-config.Hysteresis) && t <= (config.THigh+config.Hysteresis)
			} else {
				// Out of range, need to hit tlow or thigh
				gpio.inRange = t >= config.TLow && t <= config.THigh
			}
			gpioValue = gpio.inRange

			if config.Inverted {
				gpioValue = !gpioValue
			}
		}

		if err := gpio.Output.Set(gpioValue); err != nil {
			err := fmt.Errorf("%w: on gpio ID: %v, on Set with value: %v", err, config.ID, gpioValue)
			p.status.Errors = append(p.status.Errors, err.Error())
			continue
		}

		p.status.GPIO = append(p.status.GPIO, GPIOPhaseStatus{
			ID:    config.ID,
			State: gpioValue,
		})
	}
}

func (p *Process) configurePhase(phase int, config PhaseConfig) {
	if p.running.Load() {
		// If we are during process and configuring current phase
		if p.status.PhaseNumber == phase {
			// Heaters, GPIO and Thresholds will be taken care just by next Process(), as they write values each time
			// We need to reconfigure ByTime or ByTemperature with proper time
			if p.status.Next.Type == ByTime {
				if b, ok := p.moveToNext.(*byTime); ok {
					b.duration = config.Next.SecondsToMove
				}
			} else {
				if b, ok := p.moveToNext.(*byTemperature); ok {
					b.duration = config.Next.TemperatureHoldSeconds
					b.byTime.duration = config.Next.TemperatureHoldSeconds
				}
			}
		}
	}

	p.config.Phases[phase] = config
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
			p.config.Phases = p.config.Phases[:number]
		}
	}
}

func (p *Process) validatePhaseConfig(phase int, config PhaseConfig) error {
	if phase >= p.config.PhaseNumber {
		return ErrNoSuchPhase
	}

	if err := p.validateConfigHeaters(config.Heaters); err != nil {
		return err
	}

	if err := p.validateConfigGPIO(config.GPIO); err != nil {
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

func (p *Process) validateConfigGPIO(gpio []GPIOPhaseConfig) error {
	if len(gpio) != len(p.outputs) {
		return ErrDifferentGPIOSConfig
	}
	configuredIos := make(map[string]int)
	for _, gpioConfig := range gpio {
		// Requested gpio exist?
		if _, ok := p.outputs[gpioConfig.ID]; !ok {
			return fmt.Errorf("%w. ID: %v", ErrWrongGpioID, gpioConfig.ID)
		}

		if _, ok := p.sensors[gpioConfig.SensorID]; !ok {
			return fmt.Errorf("%w. ID: %v", ErrWrongSensorID, gpioConfig.SensorID)
		}

		configuredIos[gpioConfig.ID]++
	}

	if len(configuredIos) != len(p.outputs) {
		return fmt.Errorf("not all gpios configured in phase")
	}

	// Is it needed? If we verified that len of configuredIos is equal...
	for id, times := range configuredIos {
		if times > 1 {
			return fmt.Errorf("gpio with id %v configured %v times", id, times)
		}
	}
	return nil
}

// validateConfigHeaters check values in passed HeaterPhaseConfig
func (p *Process) validateConfigHeaters(heaters []HeaterPhaseConfig) error {
	if len(heaters) == 0 {
		return ErrHeaterConfigDiffersFromHeatersLen
	}
	if len(heaters) != len(p.heaters) {
		return fmt.Errorf("different number of configs and heaters")
	}

	configuredHeaters := make(map[string]int)
	for _, heaterConfig := range heaters {
		// Requested heater exist?
		if _, ok := p.heaters[heaterConfig.ID]; !ok {
			return fmt.Errorf("%w. ID: %v", ErrWrongHeaterID, heaterConfig.ID)
		}
		// Is power in range 0-100?
		if heaterConfig.Power > 100 || heaterConfig.Power < 0 {
			return fmt.Errorf("%w: ID: %v, Value: %v", ErrWrongHeaterPower, heaterConfig.ID, heaterConfig.Power)
		}

		configuredHeaters[heaterConfig.ID]++
	}

	if len(configuredHeaters) != len(p.heaters) {
		return fmt.Errorf("not all heaters configured in phase")
	}

	// Is it needed? If we verified that len of configuredHeaters is equal
	for id, times := range configuredHeaters {
		if times > 1 {
			return fmt.Errorf("heater with id %v configured %v times", id, times)
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
