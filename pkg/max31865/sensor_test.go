/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/a-clap/iot/pkg/max31865"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SensorSuite struct {
	suite.Suite
}

type SensorTransferMock struct {
	mock.Mock
}

type SensorTriggerMock struct {
	mock.Mock
	cb func()
}

var (
	sensorMock  *SensorTransferMock
	triggerMock *SensorTriggerMock
	maxInitCall = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	maxPORState = []byte{0x0, 0x0, 0x0, 0x0, 0xFF, 0xFF, 0x0, 0x0, 0x0}
)

func TestMaxSensor(t *testing.T) {
	suite.Run(t, new(SensorSuite))
}

func (s *SensorSuite) SetupTest() {
	sensorMock = new(SensorTransferMock)
	triggerMock = new(SensorTriggerMock)
}

func (s *SensorSuite) TestNew() {
	args := []struct {
		newArgs    []max31865.Option
		call       []byte
		returnArgs []byte
	}{
		{
			newArgs:    nil,
			call:       []byte{0x80, 0xd1},
			returnArgs: []byte{0x00, 0x00},
		},
		{
			newArgs:    []max31865.Option{max31865.WithWiring(max31865.TwoWire)},
			call:       []byte{0x80, 0xc1},
			returnArgs: []byte{0x00, 0x00},
		},
		{
			newArgs:    []max31865.Option{max31865.WithWiring(max31865.ThreeWire)},
			call:       []byte{0x80, 0xd1},
			returnArgs: []byte{0x00, 0x00},
		},
		{
			newArgs:    []max31865.Option{max31865.WithWiring(max31865.FourWire)},
			call:       []byte{0x80, 0xc1},
			returnArgs: []byte{0x00, 0x00},
		},
	}
	for _, arg := range args {
		sensorMock = new(SensorTransferMock)
		// Initial configReg call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
		// Configuration call
		sensorMock.On("ReadWrite", arg.call).Return(arg.returnArgs, nil)
		arg.newArgs = append(arg.newArgs, max31865.WithReadWriteCloser(sensorMock))
		max, _ := max31865.NewSensor(arg.newArgs...)
		s.NotNil(max)
	}
}

func (s *SensorSuite) TestTemperature() {
	// Data based on max datasheet
	args := []struct {
		returnArgs []byte
		tmp        float32
		err        error
	}{
		{
			returnArgs: []byte{0x0, 0xd1, 0x0B, 0xDA, 0xFF, 0xFF, 0x0, 0x0, 0x0},
			err:        nil,
			tmp:        -200.0,
		},
		{
			returnArgs: []byte{0x0, 0xd1, 0x12, 0xB4, 0xFF, 0xFF, 0x0, 0x0, 0x0},
			err:        nil,
			tmp:        -175.0,
		},
		{
			returnArgs: []byte{0x0, 0xd1, 0x33, 0x66, 0xFF, 0xFF, 0x0, 0x0, 0x0},
			err:        nil,
			tmp:        -50.0,
		},
		{
			returnArgs: []byte{0x0, 0xd1, 0x40, 0x00, 0xFF, 0xFF, 0x0, 0x0, 0x0},
			err:        nil,
			tmp:        0.0,
		},
		{
			returnArgs: []byte{0x0, 0xd1, 0x51, 0x54, 0xFF, 0xFF, 0x0, 0x0, 0x0},
			err:        nil,
			tmp:        70.0,
		},
	}
	for _, arg := range args {
		sensorMock = new(SensorTransferMock)
		// Initial configReg call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
		// Configuration call
		sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
		max, _ := max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
		s.NotNil(max)

		sensorMock.On("ReadWrite", maxInitCall).Return(arg.returnArgs, nil).Once()
		tmp, _, err := max.Temperature()
		s.Equal(arg.err, err)
		s.InDelta(arg.tmp, tmp, 1)
	}
}

func (s *SensorSuite) TestTemperatureError() {
	// Initial configReg call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	var max, _ = max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
	s.NotNil(max)

	// Return error (lsb of rtd set to 1)
	errArgs := []byte{0x0, 0xd1, 0x51, 0x55, 0xFF, 0xFF, 0x0, 0x0, 0x0}
	sensorMock.On("ReadWrite", maxInitCall).Return(errArgs, nil).Once()
	// max will try to reset error
	sensorMock.On("ReadWrite", []byte{0x80, 0xd3}).Return([]byte{0x00, 0xd1}, nil).Once()
	tmp, _, err := max.Temperature()
	s.ErrorContains(err, max31865.ErrRtd.Error())
	s.InDelta(0.0, tmp, 1)
}

func (s *SensorSuite) TestConfigure() {
	args := []struct {
		name        string
		newCfg      max31865.SensorConfig
		expectedErr error
	}{
		{
			name: "basic",
			newCfg: max31865.SensorConfig{
				Name:         "hey",
				ID:           "",
				Correction:   11,
				ASyncPoll:    false,
				PollInterval: 100,
				Samples:      13,
			},
			expectedErr: nil,
		},
		{
			name: "another",
			newCfg: max31865.SensorConfig{
				ID:           "",
				Correction:   11.5,
				ASyncPoll:    false,
				PollInterval: 1033,
				Samples:      1,
			},
			expectedErr: nil,
		},
		{
			name: "lack of ready interface",
			newCfg: max31865.SensorConfig{
				ID:           "",
				Correction:   -12.0,
				ASyncPoll:    true,
				PollInterval: 103,
				Samples:      15,
			},
			expectedErr: max31865.ErrNoReadyInterface,
		},
	}
	r := s.Require()
	for _, arg := range args {
		sensorMock = new(SensorTransferMock)
		triggerMock = new(SensorTriggerMock)

		// Initial configReg call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
		// Configuration call
		sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)

		sensor, _ := max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock))
		r.NotNil(sensor, arg.name)
		err := sensor.Configure(arg.newCfg)
		if arg.expectedErr != nil {
			r.NotNil(err)
			r.ErrorContains(err, arg.expectedErr.Error())
			continue
		}
		r.Nil(err)
		newCfg := sensor.GetConfig()
		r.Equal(arg.newCfg, newCfg)
	}
}

func (s *SensorSuite) TestPollTime() {
	r := s.Require()
	// Initial configReg call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	id := "max"
	max, _ := max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0), max31865.WithID(id))
	r.NotNil(max)

	cfg := max.GetConfig()
	cfg.PollInterval = 5 * time.Millisecond
	cfg.ASyncPoll = false

	r.Nil(max.Configure(cfg))

	expectedTmp := []float32{
		-200.1,
		-175.3,
		-50.7,
		0.0,
		70.0,
	}
	buffers := [][]byte{
		{0x0, 0xd1, 0x0B, 0xDA, 0xFF, 0xFF, 0x0, 0x0, 0x0},
		{0x0, 0xd1, 0x12, 0xB4, 0xFF, 0xFF, 0x0, 0x0, 0x0},
		{0x0, 0xd1, 0x33, 0x66, 0xFF, 0xFF, 0x0, 0x0, 0x0},
		{0x0, 0xd1, 0x40, 0x00, 0xFF, 0xFF, 0x0, 0x0, 0x0},
		{0x0, 0xd1, 0x51, 0x54, 0xFF, 0xFF, 0x0, 0x0, 0x0},
	}

	for _, buf := range buffers {
		sensorMock.On("ReadWrite", maxInitCall).Return(buf, nil).Once()
	}
	sensorMock.On("Close").Return(nil)
	err := max.Poll()
	s.Nil(err)
	<-time.After(cfg.PollInterval*time.Duration(len(buffers)) + time.Millisecond)

	s.Nil(max.Close())
	readings := max.GetReadings()
	r.Len(readings, len(expectedTmp))

	for i := range readings {
		r.Equal("", readings[i].Error)
		r.InDelta(expectedTmp[i], readings[i].Temperature, 0.1)

		average := float32(0)
		for j := 0; j <= i; j++ {
			average += expectedTmp[j]
		}
		average /= float32(i + 1)
		r.InDelta(average, readings[i].Average, 0.1)

		if i == 0 {
			continue
		}

		diff := readings[i].Stamp.Sub(readings[i-1].Stamp)
		r.InDelta(cfg.PollInterval, diff, float64(1*time.Millisecond))
	}

}

func (s *SensorSuite) TestPollTwice() {
	// Initial configReg call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	max, _ := max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
	s.NotNil(max)

	// Will call once at least
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	err := max.Poll()
	s.Nil(err)
	err = max.Poll()
	s.ErrorContains(err, max31865.ErrAlreadyPolling.Error())

	sensorMock.On("Close").Return(nil)
	s.Nil(max.Close())
}

func (s *SensorSuite) TestPollTriggerReturnsCorrectErrors() {
	r := s.Require()
	// Initial configReg call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil).Once()
	max, _ := max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0), max31865.WithReady(triggerMock))
	r.NotNil(max)

	cfg := max.GetConfig()
	cfg.ASyncPoll = true
	r.Nil(max.Configure(cfg))

	// Error on Open
	triggerErr := errors.New("broken")
	triggerMock.On("Open", mock.Anything).Return(triggerErr).Once()
	err := max.Poll()
	r.NotNil(err)
	r.ErrorContains(err, triggerErr.Error())

	// Try to get too much triggers
	triggerMock.On("Open", mock.Anything).Return(nil).Once()
	err = max.Poll()
	r.Nil(err)

	tmp := []byte{0x0, 0xd1, 0x40, 0x00, 0xFF, 0xFF, 0x0, 0x0, 0x0}
	sensorMock.On("ReadWrite", []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}).Return(tmp, nil)

	triggerMock.cb()
	triggerMock.cb()
	<-time.After(1 * time.Millisecond)
	readings := max.GetReadings()
	r.Len(readings, 2)
	// One of error should be ErrToMuchTriggers, dunno which one
	// The other error should be empty, so...
	r.Equal(max31865.ErrTooMuchTriggers.Error(), readings[0].Error+readings[1].Error)

	// Here, if we call Readings, it should be empty
	r.Nil(max.GetReadings())

	// Proper close
	triggerMock.On("Close").Return(nil).Once()
	sensorMock.On("Close").Once()
	r.Nil(max.Close())
	// Second call should do... nothing
	r.Nil(max.Close())
}

func (s *SensorSuite) TestNew_Errors() {
	t := s.Require()
	{
		// no ReadWriteCloser interface
		max, err := max31865.NewSensor()
		t.Nil(max)
		t.NotNil(err)
		t.ErrorIs(err, max31865.ErrNoReadWriteCloser)
	}
	{
		// error on Opt
		e := max31865.WithSpidev("blah")(nil)
		max, err := max31865.NewSensor(max31865.WithSpidev("blah"))
		t.Nil(max)
		t.NotNil(err)
		t.ErrorContains(err, e.Error())
	}
	{
		// Error on initial config
		// Initial configReg call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
		// Configuration call
		err := io.ErrClosedPipe
		sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, err).Once()
		max, e := max31865.NewSensor(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0), max31865.WithReady(triggerMock))
		t.Nil(max)
		t.ErrorIs(e, err)
	}

}
func (s *SensorTriggerMock) Open(callback func()) error {
	called := s.Called()
	s.cb = callback
	return called.Error(0)
}

func (s *SensorTriggerMock) Close() {
	_ = s.Called()
}

func (s *SensorTransferMock) Close() error {
	args := s.Called()
	return args.Error(0)
}

func (s *SensorTransferMock) ReadWrite(write []byte) (read []byte, err error) {
	args := s.Called(write)
	return args.Get(0).([]byte), args.Error(1)
}
