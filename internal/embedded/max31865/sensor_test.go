package max31865_test

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SensorSuite struct {
	suite.Suite
}

type SensorTransferMock struct {
	mock.Mock
}

type SensorTriggerMock struct {
	mock.Mock
	cb   func(any) error
	args any
}

func (s *SensorTransferMock) Close() error {
	args := s.Called()
	return args.Error(0)
}

func (s *SensorTransferMock) ReadWrite(write []byte) (read []byte, err error) {
	args := s.Called(write)
	return args.Get(0).([]byte), args.Error(1)
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
		// Initial config call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
		// Configuration call
		sensorMock.On("ReadWrite", arg.call).Return(arg.returnArgs, nil)
		arg.newArgs = append(arg.newArgs, max31865.WithReadWriteCloser(sensorMock))
		max, _ := max31865.New(arg.newArgs...)
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
		// Initial config call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
		// Configuration call
		sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
		max, _ := max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
		s.NotNil(max)

		sensorMock.On("ReadWrite", maxInitCall).Return(arg.returnArgs, nil).Once()
		tmp, err := max.Temperature()
		s.Equal(arg.err, err)
		s.InDelta(arg.tmp, tmp, 1)
	}
}

func (s *SensorSuite) TestTemperatureError() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	var max, _ = max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
	s.NotNil(max)

	// Return error (lsb of rtd set to 1)
	errArgs := []byte{0x0, 0xd1, 0x51, 0x55, 0xFF, 0xFF, 0x0, 0x0, 0x0}
	sensorMock.On("ReadWrite", maxInitCall).Return(errArgs, nil).Once()
	// max will try to reset error
	sensorMock.On("ReadWrite", []byte{0x80, 0xd3}).Return([]byte{0x00, 0xd1}, nil).Once()
	tmp, err := max.Temperature()
	s.ErrorIs(err, max31865.ErrRtd)
	s.InDelta(0.0, tmp, 1)
}

func (s *SensorSuite) TestPollErrors() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	max, _ := max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	pollTime := time.Duration(-1)

	err := max.Poll(dataCh, pollTime)
	s.NotNil(err)
	s.ErrorIs(err, max31865.ErrNoReadyInterface)
}

func (s *SensorSuite) TestPollTime() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	id := "max"
	max, _ := max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0), max31865.WithID(id))
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	pollTime := 5 * time.Millisecond

	err := max.Poll(dataCh, pollTime)
	s.Nil(err)

	expectedTmp := []float32{
		-200.0,
		-175.0,
		-50.0,
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

	for i := 0; i < 5; i++ {
		now := time.Now()
		select {
		case r := <-dataCh:
			s.Nil(r.Error())
			s.EqualValues(id, r.ID())
			s.InDelta(expectedTmp[i], r.Temperature(), 1)
			diff := r.Stamp().Sub(now)
			s.InDelta(pollTime.Milliseconds(), diff.Milliseconds(), 1)
		case <-time.After(2 * pollTime):
			s.Fail("failed, waiting for readings too long")
		}
	}
	sensorMock.On("Close").Return(nil)
	wait := make(chan struct{})

	go func() {
		s.Nil(max.Close())
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(2 * pollTime):
		s.Fail("should be done after this time")
	}
	close(wait)

}

func (s *SensorSuite) TestPollTwice() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	max, _ := max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0))
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	pollTime := 5 * time.Millisecond

	// Will call once at least
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	err := max.Poll(dataCh, pollTime)
	s.Nil(err)
	err = max.Poll(dataCh, pollTime)
	s.ErrorIs(err, max31865.ErrAlreadyPolling)

	sensorMock.On("Close").Return(nil)
	wait := make(chan struct{})

	go func() {
		s.Nil(max.Close())
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(2 * pollTime):
		s.Fail("should be done after this time")
	}
	close(wait)
}

func (s *SensorSuite) TestPollTriggerReturnsCorrectErrors() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil).Once()
	max, _ := max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0), max31865.WithReady(triggerMock))
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	triggerErr := errors.New("broken")
	triggerMock.On("Open", mock.Anything).Return(triggerErr).Once()
	err := max.Poll(dataCh, -1)
	s.NotNil(err)
	s.ErrorIs(err, triggerErr)

	{
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Twice()
		sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil).Once()
		max, _ := max31865.New(max31865.WithReadWriteCloser(sensorMock), max31865.WithRefRes(400.0), max31865.WithReady(triggerMock))

		triggerMock.On("Open", mock.Anything).Return(nil).Once()
		err = max.Poll(dataCh, -1)
		s.Nil(err)

		err = triggerMock.cb(triggerMock.args)
		s.Nil(err)

		err = triggerMock.cb(triggerMock.args)
		s.NotNil(err)
		s.ErrorIs(max31865.ErrTooMuchTriggers, err)

	}
}

func (s *SensorTriggerMock) Open(callback func(any) error, args any) error {
	called := s.Called()
	s.cb = callback
	s.args = args
	return called.Error(0)
}

func (s *SensorTriggerMock) Close() {
	_ = s.Called()
}
