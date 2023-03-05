/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process_test

import (
	"testing"

	"github.com/a-clap/iot/internal/distillation/process"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProcessConfigSuite struct {
	suite.Suite
}

type ClockMock struct {
	mock.Mock
}

type HeaterMock struct {
	mock.Mock
}

type OutputMock struct {
	mock.Mock
}

type SensorMock struct {
	mock.Mock
}

func TestProcessConfig(t *testing.T) {
	suite.Run(t, new(ProcessConfigSuite))
}

func (ps *ProcessConfigSuite) TestConfigurePhase_SensorsError() {
	t := ps.Require()

	args := []struct {
		name             string
		moveToNextConfig process.MoveToNextConfig
		sensorsIDs       []string
		err              error
	}{
		{
			name: "byTime - time can't be 0",
			moveToNextConfig: process.MoveToNextConfig{
				Type:                   process.ByTime,
				SensorID:               "",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          0,
			},
			sensorsIDs: []string{"s1"},
			err:        process.ErrByTimeWrongTime,
		},
		{
			name: "byTime - seconds under 0",
			moveToNextConfig: process.MoveToNextConfig{
				Type:                   process.ByTime,
				SensorID:               "",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          -1,
			},
			sensorsIDs: []string{"s1"},
			err:        process.ErrByTimeWrongTime,
		},
		{
			name: "byTime - all good",
			moveToNextConfig: process.MoveToNextConfig{
				Type:                   process.ByTime,
				SensorID:               "",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          1,
			},
			sensorsIDs: []string{"s1"},
			err:        nil,
		},
		{
			name: "byTemperature - wrong sensor",
			moveToNextConfig: process.MoveToNextConfig{
				Type:                   process.ByTemperature,
				SensorID:               "s2",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          0,
			},
			sensorsIDs: []string{"s1"},
			err:        process.ErrByTemperatureWrongID,
		},
		{
			name: "byTemperature - weird type",
			moveToNextConfig: process.MoveToNextConfig{
				Type:                   process.MoveToNextType(3),
				SensorID:               "s1",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          0,
			},
			sensorsIDs: []string{"s1"},
			err:        process.ErrUnknownType,
		},
		{
			name: "byTemperature - all good, threshold/hold can be 0",
			moveToNextConfig: process.MoveToNextConfig{
				Type:                   process.ByTemperature,
				SensorID:               "s1",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          0,
			},
			sensorsIDs: []string{"s1"},
			err:        nil,
		},
	}
	for _, arg := range args {
		// Always good config - except Next
		phaseConfig := process.PhaseConfig{
			Next: process.MoveToNextConfig{
				Type:                   0,
				SensorID:               "",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          0,
			},
			Heaters: []process.HeaterPhaseConfig{
				{ID: "h1", Power: 13},
			},
			GPIO: nil,
		}

		heaterMock := new(HeaterMock)
		heaterMock.On("ID").Return("h1")
		heaters := append(make([]process.Heater, 0, 1), heaterMock)

		sensors := make([]process.Sensor, len(arg.sensorsIDs))
		for i, sensor := range arg.sensorsIDs {
			h := new(SensorMock)
			h.On("ID").Return(sensor)
			sensors[i] = h
		}

		phaseConfig.Next = arg.moveToNextConfig
		p, err := process.New(
			process.WithHeaters(heaters),
			process.WithSensors(sensors))

		t.Nil(err)
		t.NotNil(p)

		t.Nil(p.SetPhases(5), arg.name)
		err = p.ConfigurePhase(3, phaseConfig)
		if arg.err != nil {
			t.NotNil(err, arg.name)
			t.ErrorContains(err, arg.err.Error(), arg.name)
			continue
		}
		t.Nil(err, arg.name)
		cfg := p.GetConfig()
		t.EqualValues(phaseConfig, cfg.Phases[3], arg.name)
	}
}

func (ps *ProcessConfigSuite) TestConfigurePhase_PhaseCount() {
	t := ps.Require()

	args := []struct {
		name                string
		phaseCount          int
		phaseNumberToConfig int
		err                 error
	}{
		{
			name:                "wrong phase number to config",
			phaseCount:          3,
			phaseNumberToConfig: 4,
			err:                 process.ErrNoSuchPhase,
		},
		{
			name:                "wrong phase number to config - count starts from 0",
			phaseCount:          3,
			phaseNumberToConfig: 3,
			err:                 process.ErrNoSuchPhase,
		},
	}
	for _, arg := range args {
		// Always good config - except phaseCounts
		phaseConfig := process.PhaseConfig{
			Next: process.MoveToNextConfig{
				Type:                   0,
				SensorID:               "",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          0,
			},
			Heaters: nil,
			GPIO:    nil,
		}

		heaterMock := new(HeaterMock)
		heaterMock.On("ID").Return("s1")
		heaters := append(make([]process.Heater, 0, 1), heaterMock)

		sensorMock := new(SensorMock)
		sensorMock.On("ID").Return("s1")
		sensors := append(make([]process.Sensor, 0, 1), sensorMock)

		p, err := process.New(
			process.WithHeaters(heaters),
			process.WithSensors(sensors))

		t.Nil(err)
		t.NotNil(p)

		t.Nil(p.SetPhases(arg.phaseCount), arg.name)
		err = p.ConfigurePhase(arg.phaseNumberToConfig, phaseConfig)
		t.NotNil(err, arg.name)
		t.ErrorContains(err, arg.err.Error())
	}
}
func (ps *ProcessConfigSuite) TestConfigurePhase_HeatersError() {
	t := ps.Require()

	args := []struct {
		name   string
		heater []struct {
			id string
		}
		heatersConfig []process.HeaterPhaseConfig
		err           error
	}{
		{
			name: "wrong id of heater",
			heater: []struct{ id string }{
				{
					id: "h1",
				},
			},
			heatersConfig: []process.HeaterPhaseConfig{
				{
					ID:    "h2",
					Power: 13,
				},
			},
			err: process.ErrWrongHeaterID,
		},
		{
			name: "power of heater over 100",
			heater: []struct{ id string }{
				{
					id: "h1",
				},
			},
			heatersConfig: []process.HeaterPhaseConfig{
				{
					ID:    "h1",
					Power: 101,
				},
			},
			err: process.ErrWrongHeaterPower,
		},
		{
			name: "lack of heater configuration",
			heater: []struct{ id string }{
				{
					id: "h1",
				},
			},
			heatersConfig: nil,
			err:           process.ErrNoHeatersInConfig,
		},

		{
			name: "all good",
			heater: []struct{ id string }{
				{
					id: "h1",
				},
			},
			heatersConfig: []process.HeaterPhaseConfig{
				{
					ID:    "h1",
					Power: 15,
				},
			},
			err: nil,
		},
	}
	for _, arg := range args {
		// Always good config - except heaters
		phaseConfig := process.PhaseConfig{
			Next: process.MoveToNextConfig{
				Type:                   process.ByTime,
				SensorID:               "",
				SensorThreshold:        0,
				TemperatureHoldSeconds: 0,
				SecondsToMove:          3,
			},
			Heaters: nil,
			GPIO:    nil,
		}

		heaters := make([]process.Heater, len(arg.heater))
		for i, heater := range arg.heater {
			h := new(HeaterMock)
			h.On("ID").Return(heater.id)
			heaters[i] = h
		}

		sensorMock := new(SensorMock)
		sensorMock.On("ID").Return("s1")
		sensors := append(make([]process.Sensor, 0, 1), sensorMock)

		phaseConfig.Heaters = arg.heatersConfig
		p, err := process.New(
			process.WithHeaters(heaters),
			process.WithSensors(sensors))

		t.Nil(err)
		t.NotNil(p)

		t.Nil(p.SetPhases(5), arg.name)
		err = p.ConfigurePhase(3, phaseConfig)
		if arg.err != nil {
			t.NotNil(err, arg.name)
			t.ErrorContains(err, arg.err.Error(), arg.name)
			continue
		}
		t.Nil(err, arg.name)
		cfg := p.GetConfig()
		t.EqualValues(phaseConfig, cfg.Phases[3], arg.name)
	}
}

func (ps *ProcessConfigSuite) TestSetPhases_ReflectsBuffer() {
	t := ps.Require()

	heaterMock := new(HeaterMock)
	heaterMock.On("ID").Return("h1")

	sensorMock := new(SensorMock)
	sensorMock.On("ID").Return("s1")

	heaters := []process.Heater{heaterMock}
	sensors := []process.Sensor{sensorMock}

	p, err := process.New(process.WithHeaters(heaters), process.WithSensors(sensors))
	t.Nil(err)
	t.NotNil(p)

	args := []struct {
		name       string
		newLen     int
		notLessCap int
	}{
		{
			name:       "single element",
			newLen:     1,
			notLessCap: 10,
		},
		{
			name:       "5 elements",
			newLen:     5,
			notLessCap: 10,
		},
		{
			name:       "10 elements",
			newLen:     10,
			notLessCap: 10,
		},
		{
			name:       "20 elements",
			newLen:     20,
			notLessCap: 20,
		},
		{
			name:       "15 elements",
			newLen:     15,
			notLessCap: 20,
		},
		{
			name:       "trim to 10 elements",
			newLen:     9,
			notLessCap: 10,
		},
	}
	for _, arg := range args {
		t.Nil(p.SetPhases(arg.newLen), arg.name)
		cfg := p.GetConfig()
		t.EqualValues(arg.newLen, cfg.PhaseNumber, arg.name)
		t.EqualValues(arg.newLen, len(cfg.Phases), arg.name)
		t.LessOrEqual(arg.notLessCap, cap(cfg.Phases), arg.name)
	}

}
func (ps *ProcessConfigSuite) TestSetPhases_Errors() {
	t := ps.Require()

	heaterMock := new(HeaterMock)
	heaterMock.On("ID").Return("h1")

	sensorMock := new(SensorMock)
	sensorMock.On("ID").Return("s1")

	heaters := []process.Heater{heaterMock}
	sensors := []process.Sensor{sensorMock}
	p, err := process.New(process.WithHeaters(heaters), process.WithSensors(sensors))
	t.Nil(err)
	t.NotNil(p)
	{
		err := p.SetPhases(0)
		t.ErrorContains(err, process.ErrPhasesBelowZero.Error(), "0 phases")
	}
	{
		err := p.SetPhases(-1)
		t.ErrorContains(err, process.ErrPhasesBelowZero.Error(), "-1 phases")
	}
	{
		err := p.SetPhases(1)
		t.Nil(err, "all good")
	}

}

func (ps *ProcessConfigSuite) TestNew() {
	t := ps.Require()

	args := []struct {
		name      string
		heaters   []*HeaterMock
		heatersID []string
		sensors   []*SensorMock
		sensorsID []string
		outputs   []*OutputMock
		outputsID []string
		clock     *ClockMock
		err       error
	}{
		{
			name:      "lack of any interface",
			heaters:   nil,
			heatersID: nil,
			sensors:   nil,
			sensorsID: nil,
			outputs:   nil,
			outputsID: nil,
			clock:     nil,
			err:       process.ErrNoTSensors,
		},
		{
			name:      "lack of heaters interface",
			heaters:   nil,
			heatersID: nil,
			sensors:   []*SensorMock{new(SensorMock)},
			sensorsID: []string{"s1"},
			outputs:   []*OutputMock{new(OutputMock)},
			outputsID: []string{"o1"},
			clock:     new(ClockMock),
			err:       process.ErrNoHeaters,
		},
		{
			name:      "lack of sensors interface",
			heaters:   []*HeaterMock{new(HeaterMock)},
			heatersID: []string{"h1"},
			sensors:   nil,
			sensorsID: nil,
			outputs:   []*OutputMock{new(OutputMock)},
			outputsID: []string{"o1"},
			clock:     new(ClockMock),
			err:       process.ErrNoTSensors,
		},
		{
			name:      "lack of outputs interface - it is okay",
			heaters:   []*HeaterMock{new(HeaterMock)},
			heatersID: []string{"h1"},
			sensors:   []*SensorMock{new(SensorMock)},
			sensorsID: []string{"s1"},
			outputs:   nil,
			outputsID: nil,
			clock:     new(ClockMock),
			err:       nil,
		},
		{
			name:      "clock interface isn't needed",
			heaters:   []*HeaterMock{new(HeaterMock)},
			heatersID: []string{"h1"},
			sensors:   []*SensorMock{new(SensorMock)},
			sensorsID: []string{"s1"},
			outputs:   []*OutputMock{new(OutputMock)},
			outputsID: []string{"o1"},
			clock:     nil,
			err:       nil,
		},
	}
	for _, arg := range args {
		t.EqualValues(len(arg.heaters), len(arg.heatersID), arg.name)
		t.EqualValues(len(arg.sensors), len(arg.sensorsID), arg.name)
		t.EqualValues(len(arg.outputs), len(arg.outputs), arg.name)

		heaters := make([]process.Heater, len(arg.heaters))
		for i := range arg.heaters {
			arg.heaters[i].On("ID").Return(arg.heatersID[i])
			heaters[i] = arg.heaters[i]
		}
		sensors := make([]process.Sensor, len(arg.sensors))
		for i := range arg.sensors {
			arg.sensors[i].On("ID").Return(arg.sensorsID[i])
			sensors[i] = arg.sensors[i]
		}
		outputs := make([]process.Output, len(arg.outputs))
		for i := range arg.outputs {
			arg.outputs[i].On("ID").Return(arg.outputsID[i])
			outputs[i] = arg.outputs[i]
		}

		options := []process.Option{
			process.WithHeaters(heaters),
			process.WithSensors(sensors),
			process.WithOutputs(outputs),
		}
		if arg.clock != nil {
			options = append(options, process.WithClock(arg.clock))
		}

		p, err := process.New(options...)

		if arg.err != nil {
			t.Nil(p, arg.name)
			t.NotNil(err, arg.name)
			t.ErrorContains(err, arg.err.Error(), arg.name)
			continue
		}
		t.NotNil(p, arg.name)
		t.Nil(err, arg.name)
	}

}

func (m *ClockMock) Unix() int64 {
	return m.Called().Get(0).(int64)
}

func (m *SensorMock) ID() string {
	return m.Called().String(0)
}

func (m *SensorMock) Temperature() float64 {
	return m.Called().Get(0).(float64)
}
func (m *OutputMock) ID() string {
	return m.Called().String(0)
}

func (m *OutputMock) Set(value bool) error {
	return m.Called(value).Error(0)
}

func (m *HeaterMock) ID() string {
	return m.Called().String(0)
}

func (m *HeaterMock) SetPower(pwr int) error {
	return m.Called(pwr).Error(0)
}
