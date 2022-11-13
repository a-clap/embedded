package max31865_test

import (
	"github.com/a-clap/iot/pkg/max31865"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"time"
)

type SensorSuite struct {
	suite.Suite
}

type SensorTransferMock struct {
	mock.Mock
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
	maxInitCall = []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	maxPORState = []byte{0x0, 0x0, 0x0, 0x0, 0xFF, 0xFF, 0x0, 0x0, 0x0}
)

func TestMaxSensor(t *testing.T) {
	suite.Run(t, new(SensorSuite))
}

func (s *SensorSuite) SetupTest() {
	sensorMock = new(SensorTransferMock)
}

func (s *SensorSuite) TestNew() {
	args := []struct {
		newArgs    []any
		call       []byte
		returnArgs []byte
	}{
		{
			newArgs:    nil,
			call:       []byte{0x80, 0xd1},
			returnArgs: []byte{0x00, 0x00},
		},
		{
			newArgs:    []any{max31865.TwoWire},
			call:       []byte{0x80, 0xc1},
			returnArgs: []byte{0x00, 0x00},
		},
		{
			newArgs:    []any{max31865.ThreeWire},
			call:       []byte{0x80, 0xd1},
			returnArgs: []byte{0x00, 0x00},
		},
		{
			newArgs:    []any{max31865.FourWire},
			call:       []byte{0x80, 0xc1},
			returnArgs: []byte{0x00, 0x00},
		},
	}
	for _, arg := range args {
		sensorMock = new(SensorTransferMock)
		// Initial config call, always constant
		sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil)
		// Configuration call
		sensorMock.On("ReadWrite", arg.call).Return(arg.returnArgs, nil)
		max, _ := max31865.New(sensorMock, arg.newArgs...)
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
		max, _ := max31865.New(sensorMock, max31865.RefRes(400.0))
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
	max, _ := max31865.New(sensorMock, max31865.RefRes(400.0))
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
	max, _ := max31865.New(sensorMock, max31865.RefRes(400.0))
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	stopCh := make(chan struct{})
	pollTime := time.Duration(-1)

	finCh, errCh, err := max.Poll(dataCh, stopCh, pollTime)
	s.Nil(finCh)
	s.Nil(errCh)
	s.NotNil(err)
	s.ErrorIs(err, max31865.ErrNoReadyInterface)
}

func (s *SensorSuite) TestPollTime() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	id := max31865.ID("max")
	max, _ := max31865.New(sensorMock, max31865.RefRes(400.0), id)
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	stopCh := make(chan struct{})
	pollTime := 5 * time.Millisecond

	finCh, errCh, err := max.Poll(dataCh, stopCh, pollTime)
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
			rid, tmp, stamp := r.Get()
			s.EqualValues(id, rid)
			val, _ := strconv.ParseFloat(tmp, 32)
			s.InDelta(expectedTmp[i], float32(val), 1)
			diff := stamp.Sub(now)
			s.InDelta(pollTime.Milliseconds(), diff.Milliseconds(), 1)
		case e := <-errCh:
			s.Fail("received error ", e)
		case <-time.After(2 * pollTime):
			s.Fail("failed, waiting for readings too long")
		}
	}

	stopCh <- struct{}{}
	for range dataCh {
	}

	select {
	case <-finCh:
	case <-time.After(2 * pollTime):
		s.Fail("should be done after this time")
	}

}

func (s *SensorSuite) TestPollTwice() {
	// Initial config call, always constant
	sensorMock.On("ReadWrite", maxInitCall).Return(maxPORState, nil).Once()
	// Configuration call
	sensorMock.On("ReadWrite", []byte{0x80, 0xd1}).Return([]byte{0x00, 0x00}, nil)
	max, _ := max31865.New(sensorMock, max31865.RefRes(400.0))
	s.NotNil(max)

	dataCh := make(chan max31865.Readings)
	stopCh := make(chan struct{})
	pollTime := 5 * time.Millisecond

	finCh, _, err := max.Poll(dataCh, stopCh, pollTime)
	s.Nil(err)
	_, _, err = max.Poll(dataCh, stopCh, pollTime)
	s.ErrorIs(err, max31865.ErrAlreadyPolling)
	stopCh <- struct{}{}
	for range dataCh {
	}
	<-finCh
}
