/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process_test

import (
	"errors"
	"testing"
	"time"

	"github.com/a-clap/iot/internal/distillation/process"
	"github.com/stretchr/testify/suite"
)

type ProcessTestSuite struct {
	suite.Suite
}

func TestProcess(t *testing.T) {
	suite.Run(t, new(ProcessTestSuite))
}
func (pts *ProcessTestSuite) TestHappyPath_ConfigureOnFly() {
	t := pts.Require()

	heaterMock := new(HeaterMock)
	heaterMock.On("ID").Return("h1")
	heaters := []process.Heater{heaterMock}

	sensorMock := new(SensorMock)
	sensorMock.On("ID").Return("s1")
	sensors := []process.Sensor{sensorMock}

	outputMock := new(OutputMock)
	outputMock.On("ID").Return("o1")
	outputs := []process.Output{outputMock}

	clockMock := new(ClockMock)

	p, err := process.New(
		process.WithSensors(sensors),
		process.WithHeaters(heaters),
		process.WithOutputs(outputs),
		process.WithClock(clockMock),
	)
	t.Nil(err)
	t.NotNil(p)

	cfg := process.Config{
		PhaseNumber: 2,
		Phases: []process.PhaseConfig{
			{
				Next: process.MoveToNextConfig{
					Type:                   process.ByTime,
					SensorID:               "",
					SensorThreshold:        0,
					TemperatureHoldSeconds: 0,
					SecondsToMove:          100,
				},
				Heaters: []process.HeaterPhaseConfig{
					{
						ID:    "h1",
						Power: 13,
					},
				},
				GPIO: []process.GPIOPhaseConfig{
					{
						ID:         "o1",
						SensorID:   "s1",
						TLow:       10,
						THigh:      20,
						Hysteresis: 1,
						Inverted:   false,
					},
				},
			},
			{
				Next: process.MoveToNextConfig{
					Type:                   process.ByTemperature,
					SensorID:               "s1",
					SensorThreshold:        75.0,
					TemperatureHoldSeconds: 10,
					SecondsToMove:          0,
				},
				Heaters: []process.HeaterPhaseConfig{
					{
						ID:    "h1",
						Power: 40,
					},
				},
				GPIO: []process.GPIOPhaseConfig{
					{
						ID:         "o1",
						SensorID:   "s1",
						TLow:       50,
						THigh:      51,
						Hysteresis: 1,
						Inverted:   true,
					},
				},
			},
		},
	}
	t.Nil(p.Configure(cfg))
	retTime := int64(0)
	// We expect:
	// 1. Single call to Unix()
	clockMock.On("Unix").Return(retTime).Once()
	// 2. Twice Get temperature from sensor, so it can be placed in status
	sensorMock.On("Temperature").Return(1.1).Times(4)
	// 3. Set Power on heater
	heaterMock.On("SetPower", 13).Return(nil).Twice()
	// 4. Handle GPIO
	outputMock.On("Set", false).Return(nil).Twice()
	s, err := p.Run()
	t.Nil(err)
	expectedStatus := process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(retTime, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 100,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: false,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Second time, just to move time
	retTime = 5
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(1.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", false).Return(nil).Twice()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 95,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: false,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Now, lets modify config and see what happens
	phase := cfg.Phases[0]
	phase.GPIO[0].Inverted = true
	phase.Heaters[0].Power = 56
	phase.Next.SecondsToMove = 500
	t.Nil(p.ConfigurePhase(0, phase))

	// We expect that GPIO should be set now - it is inverted
	// Power of heater should be 56
	retTime = 10
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(1.1).Twice()
	heaterMock.On("SetPower", 56).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Twice()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 490,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 56,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Move to second phase
	retTime = 501
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(1.1).Twice()
	heaterMock.On("SetPower", 40).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Twice()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 1,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 40,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Now, lets modify config - TemperatureHoldSeconds and see what happens
	phase = cfg.Phases[1]
	phase.Next.TemperatureHoldSeconds = 500
	t.Nil(p.ConfigurePhase(1, phase))
	retTime = 600
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(77.1).Times(3)
	heaterMock.On("SetPower", 40).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Twice()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 1,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        500,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 40,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 77.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Now some time elapsed, and we are changing hold again
	phase = cfg.Phases[1]
	phase.Next.TemperatureHoldSeconds = 800
	t.Nil(p.ConfigurePhase(1, phase))
	retTime = 700
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(77.1).Times(3)
	heaterMock.On("SetPower", 40).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Twice()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 1,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        700,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 40,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 77.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

}
func (pts *ProcessTestSuite) TestHappyPath_VerifyGPIOHandlingSinglePhase() {
	t := pts.Require()

	heaterMock := new(HeaterMock)
	heaterMock.On("ID").Return("h1")
	heaters := []process.Heater{heaterMock}

	sensorMock := new(SensorMock)
	sensorMock.On("ID").Return("s1")
	sensors := []process.Sensor{sensorMock}

	outputMock := new(OutputMock)
	outputMock.On("ID").Return("o1")
	outputs := []process.Output{outputMock}

	clockMock := new(ClockMock)

	p, err := process.New(
		process.WithSensors(sensors),
		process.WithHeaters(heaters),
		process.WithOutputs(outputs),
		process.WithClock(clockMock),
	)
	t.Nil(err)
	t.NotNil(p)

	cfg := process.Config{
		PhaseNumber: 1,
		Phases: []process.PhaseConfig{
			{
				Next: process.MoveToNextConfig{
					Type:                   process.ByTime,
					SensorID:               "",
					SensorThreshold:        0,
					TemperatureHoldSeconds: 0,
					SecondsToMove:          100,
				},
				Heaters: []process.HeaterPhaseConfig{
					{
						ID:    "h1",
						Power: 13,
					},
				},
				GPIO: []process.GPIOPhaseConfig{
					{
						ID:         "o1",
						SensorID:   "s1",
						TLow:       10,
						THigh:      20,
						Hysteresis: 1,
						Inverted:   false,
					},
				},
			},
		},
	}

	t.Nil(p.Configure(cfg))
	retTime := int64(0)
	// We expect:
	// 1. Single call to Unix()
	clockMock.On("Unix").Return(retTime).Once()
	// 2. Twice Get temperature from sensor, so it can be placed in status
	sensorMock.On("Temperature").Return(1.1).Times(4)
	// 3. Set Power on heater
	heaterMock.On("SetPower", 13).Return(nil).Twice()
	// 4. Handle GPIO
	outputMock.On("Set", false).Return(nil).Twice()
	s, err := p.Run()
	t.Nil(err)
	expectedStatus := process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(retTime, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 100,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: false,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Second: gpio should be set
	retTime = 5
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(15.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 95,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 15.1,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Third call, gpio should still be set
	retTime = 17
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(21.0).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 83,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 21,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Fourth call, gpio should still be set
	retTime = 17
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(9.0).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 83,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 9,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Fifth call, gpio should be off
	retTime = 17
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(8.9).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", false).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 83,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 8.9,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: false,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Sixth call, gpio still off
	retTime = 17
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(9.9).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", false).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 83,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 9.9,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: false,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Seventh call, gpio ON
	retTime = 17
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(10.0).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	outputMock.On("Set", true).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 83,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 10.0,
			},
		},
		GPIO: []process.GPIOPhaseStatus{
			{
				ID:    "o1",
				State: true,
			},
		},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	//
	// FourthCall - finished
	// retTime = 101
	// clockMock.On("Unix").Return(retTime).Once()
	// sensorMock.On("Temperature").Return(100.1).Once()
	// // Disable heater
	// heaterMock.On("SetPower", 0).Return(nil).Once()
	// s, err = p.Process()
	// t.Nil(err)
	// expectedStatus = process.Status{
	// 	Done:        true,
	// 	PhaseNumber: 0,
	// 	StartTime:   time.Unix(0, 0),
	// 	EndTime:     time.Unix(retTime, 0),
	// 	Next: process.MoveToNextStatus{
	// 		Type: process.ByTime,
	// 		Time: process.MoveToNextStatusTime{
	// 			TimeLeft: 1,
	// 		},
	// 		Temperature: process.MoveToNextStatusTemperature{},
	// 	},
	// 	Heaters: []process.HeaterPhaseStatus{},
	// 	Temperature: []process.TemperaturePhaseStatus{
	// 		{
	// 			ID:          "s1",
	// 			Temperature: 150.1,
	// 		},
	// 	},
	// 	GPIO:   []process.GPIOPhaseStatus{},
	// 	Errors: nil,
	// }
	// t.EqualValues(expectedStatus, s)
}
func (pts *ProcessTestSuite) TestHappyPath_SinglePhaseByTemperature() {
	t := pts.Require()

	heaterMock := new(HeaterMock)
	heaterMock.On("ID").Return("h1")
	heaters := []process.Heater{heaterMock}

	sensorMock := new(SensorMock)
	sensorMock.On("ID").Return("s1")
	sensors := []process.Sensor{sensorMock}

	clockMock := new(ClockMock)

	p, err := process.New(
		process.WithSensors(sensors),
		process.WithHeaters(heaters),
		process.WithClock(clockMock),
	)
	t.Nil(err)
	t.NotNil(p)

	cfg := process.Config{
		PhaseNumber: 1,
		Phases: []process.PhaseConfig{
			{
				Next: process.MoveToNextConfig{
					Type:                   process.ByTemperature,
					SensorID:               "s1",
					SensorThreshold:        75.0,
					TemperatureHoldSeconds: 10,
					SecondsToMove:          0,
				},
				Heaters: []process.HeaterPhaseConfig{
					{
						ID:    "h1",
						Power: 13,
					},
				},
				GPIO: nil,
			},
		},
	}

	t.Nil(p.Configure(cfg))
	retTime := int64(0)
	// We expect:
	// 1. Single call to Unix()
	clockMock.On("Unix").Return(retTime).Once()
	// 2. Three times get temperature from sensor: built first Status, next() from byTemperature and built status again
	sensorMock.On("Temperature").Return(1.1).Times(3)
	// 3. Set Power on heater
	heaterMock.On("SetPower", 13).Return(nil).Twice()
	s, err := p.Run()
	t.Nil(err)
	expectedStatus := process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(retTime, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// On second Process we just expect single calls, except Temperature, which is called via ByTemperature
	retTime = 5
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(74.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 74.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Third call, temperature over threshold
	retTime = 99
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(75.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 75.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// FourthCall - temperature over threshold, time not elapsed
	retTime = 101
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(100.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        8,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 100.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Fifth - temperature under threshold, timeleft should be back to 10
	retTime = 103
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(74.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 74.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Sixth - temperature over threshold
	retTime = 150
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(75.1).Twice()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 75.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Seventh - temperature over threshold, time elapsed
	retTime = 160
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(75.1).Twice()
	// Disable heater
	heaterMock.On("SetPower", 0).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        true,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(retTime, 0),
		Next: process.MoveToNextStatus{
			Type: process.ByTemperature,
			Time: process.MoveToNextStatusTime{},
			Temperature: process.MoveToNextStatusTemperature{
				SensorID:        "s1",
				SensorThreshold: 75.0,
				TimeLeft:        10,
			},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 75.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)
}
func (pts *ProcessTestSuite) TestHappyPath_SinglePhaseByTime() {
	t := pts.Require()

	heaterMock := new(HeaterMock)
	heaterMock.On("ID").Return("h1")
	heaters := []process.Heater{heaterMock}

	sensorMock := new(SensorMock)
	sensorMock.On("ID").Return("s1")
	sensors := []process.Sensor{sensorMock}

	clockMock := new(ClockMock)

	p, err := process.New(
		process.WithSensors(sensors),
		process.WithHeaters(heaters),
		process.WithClock(clockMock),
	)
	t.Nil(err)
	t.NotNil(p)

	cfg := process.Config{
		PhaseNumber: 1,
		Phases: []process.PhaseConfig{
			{
				Next: process.MoveToNextConfig{
					Type:                   process.ByTime,
					SensorID:               "",
					SensorThreshold:        0,
					TemperatureHoldSeconds: 0,
					SecondsToMove:          100,
				},
				Heaters: []process.HeaterPhaseConfig{
					{
						ID:    "h1",
						Power: 13,
					},
				},
				GPIO: nil,
			},
		},
	}

	t.Nil(p.Configure(cfg))
	retTime := int64(0)
	// We expect:
	// 1. Single call to Unix()
	clockMock.On("Unix").Return(retTime).Once()
	// 2. Twice Get temperature from sensor, so it can be placed in status
	sensorMock.On("Temperature").Return(1.1).Twice()
	// 3. Set Power on heater
	heaterMock.On("SetPower", 13).Return(nil).Twice()
	s, err := p.Run()
	t.Nil(err)
	expectedStatus := process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(retTime, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 100,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 1.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// On second Process we just expect single calls
	retTime = 5
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(100.1).Once()
	heaterMock.On("SetPower", 13).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 95,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			{
				process.HeaterPhaseConfig{
					ID:    "h1",
					Power: 13,
				},
			},
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 100.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)

	// Third call, err on SetPower
	retTime = 99
	pwrErr := errors.New("hello there")
	clockMock.On("Unix").Return(retTime).Once()
	sensorMock.On("Temperature").Return(150.1).Once()
	heaterMock.On("SetPower", 13).Return(pwrErr).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        false,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Time{},
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 1,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{
			// Empty, because error happened
		},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 150.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	// Check without errors,
	errs := s.Errors
	s.Errors = nil
	expectedStatus.Errors = nil
	t.EqualValues(expectedStatus, s)
	t.Len(errs, 1)
	t.Contains(errs[0], pwrErr.Error())

	// FourthCall - finished
	retTime = 101
	clockMock.On("Unix").Return(retTime).Once()
	// sensorMock.On("Temperature").Return(100.1).Once()
	// Disable heater
	heaterMock.On("SetPower", 0).Return(nil).Once()
	s, err = p.Process()
	t.Nil(err)
	expectedStatus = process.Status{
		Done:        true,
		PhaseNumber: 0,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(retTime, 0),
		Next: process.MoveToNextStatus{
			Type: process.ByTime,
			Time: process.MoveToNextStatusTime{
				TimeLeft: 1,
			},
			Temperature: process.MoveToNextStatusTemperature{},
		},
		Heaters: []process.HeaterPhaseStatus{},
		Temperature: []process.TemperaturePhaseStatus{
			{
				ID:          "s1",
				Temperature: 150.1,
			},
		},
		GPIO:   []process.GPIOPhaseStatus{},
		Errors: nil,
	}
	t.EqualValues(expectedStatus, s)
}
