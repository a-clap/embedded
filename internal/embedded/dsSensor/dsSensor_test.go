package dsSensor_test

import (
	"github.com/a-clap/iot/internal/embedded/dsSensor"
	"github.com/a-clap/iot/pkg/avg"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type DsSensorTestSuite struct {
	suite.Suite
	mock     *SensorHandlerMock
	pollData *PollDataMock
}

type SensorHandlerMock struct {
	mock.Mock
}

type PollDataMock struct {
	mock.Mock
}

func TestDSSensorTestSuite(t *testing.T) {
	suite.Run(t, new(DsSensorTestSuite))
}

func (t *DsSensorTestSuite) SetupTest() {
	t.mock = new(SensorHandlerMock)
}

func (t *DsSensorTestSuite) TestPoll() {
	t.mock = new(SensorHandlerMock)
	t.pollData = new(PollDataMock)
	id := "123"

	t.mock.On("ID").Return(id)
	t.mock.On("Resolution").Return(dsSensor.Resolution10BIT, nil)

	t.mock.On("Poll", mock.Anything, mock.Anything).Return(nil)
	t.mock.On("StopPoll").Return(nil)

	t.pollData.On("ID").Return(id)

	ds := dsSensor.New(t.mock)
	err := ds.Poll()
	t.Nil(err)

	// Add this point there should be any data
	stat := ds.Status()
	t.Equal(id, stat.ID)
	t.Equal(true, stat.Enabled)

	polledChannel := t.mock.TestData()["readings"].(chan dsSensor.PollData)

	temperatures := []float32{100.0, 200.0, 300.0, 400.0, 500.0}
	t.EqualValues(len(temperatures), dsSensor.DefaultSamples)

	for i, elem := range temperatures {
		now := time.Now()
		t.pollData.On("Temperature").Return(elem).Once()
		t.pollData.On("Stamp").Return(now).Once()

		polledChannel <- t.pollData
		// Just to force scheduler work
		<-time.After(1 * time.Millisecond)
		stat = ds.Status()
		t.Equal(id, stat.ID)
		t.Equal(true, stat.Enabled)
		t.Equal(now, stat.Stamp)

		expectedTmp := float32(0)
		for j := 0; j <= i; j++ {
			expectedTmp += temperatures[j]
		}
		expectedTmp /= float32(i + 1)

		t.InDelta(expectedTmp, stat.Temperature, 0.1)

	}

	err = ds.StopPoll()
	t.Nil(err)

}

func (t *DsSensorTestSuite) TestSetGetConfig() {
	args := []struct {
		newConfig dsSensor.Config
		err       error
	}{
		{
			newConfig: dsSensor.Config{
				ID:             "1",
				Enabled:        false,
				Resolution:     dsSensor.Resolution9BIT,
				PollTimeMillis: 120,
				Samples:        13,
			},
			err: nil,
		},
		{
			newConfig: dsSensor.Config{
				ID:             "2",
				Enabled:        false,
				Resolution:     dsSensor.Resolution11BIT,
				PollTimeMillis: 0,
				Samples:        1,
			},
			err: nil,
		},
		{
			newConfig: dsSensor.Config{
				ID:             "sasaaax",
				Enabled:        false,
				Resolution:     dsSensor.Resolution12BIT,
				PollTimeMillis: 163,
				Samples:        25,
			},
			err: nil,
		},
		{
			newConfig: dsSensor.Config{
				ID:             "heeeey",
				Enabled:        false,
				Resolution:     dsSensor.Resolution12BIT,
				PollTimeMillis: 163,
				Samples:        0,
			},
			err: avg.ErrSizeIsZero,
		},
	}
	for _, arg := range args {
		t.mock = new(SensorHandlerMock)
		t.mock.On("ID").Return(arg.newConfig.ID)
		t.mock.On("Resolution").Return(arg.newConfig.Resolution, nil)
		t.mock.On("SetPollTime", arg.newConfig.PollTimeMillis).Return(nil)

		ds := dsSensor.New(t.mock)
		t.NotNil(ds)

		err := ds.SetConfig(arg.newConfig)
		t.Equal(arg.err, err)
		if arg.err != nil {
			continue
		}

		c := ds.Config()
		t.Equal(arg.newConfig, c)

	}

}

func (t *DsSensorTestSuite) TestNew_VerifyConfig() {
	args := []struct {
		name string
		id   string
		res  dsSensor.Resolution
	}{
		{
			name: "10 bit reso",
			id:   "123",
			res:  dsSensor.Resolution10BIT,
		},
		{
			name: "12 bit reso",
			id:   "hello world 2",
			res:  dsSensor.Resolution12BIT,
		},
		{
			name: "11 bit reso",
			id:   "hello world 3",
			res:  dsSensor.Resolution11BIT,
		},
		{
			name: "9 bit reso",
			id:   "hello world 4",
			res:  dsSensor.Resolution9BIT,
		},
	}
	for _, arg := range args {
		t.mock = new(SensorHandlerMock)
		t.mock.On("ID").Return(arg.id)
		t.mock.On("Resolution").Return(arg.res, nil)

		ds := dsSensor.New(t.mock)
		t.NotNil(ds, arg.name)
		cfg := ds.Config()

		t.EqualValues(arg.id, cfg.ID, arg.name)
		t.EqualValues(arg.res, cfg.Resolution, arg.name)
		t.EqualValues(false, cfg.Enabled, arg.name)

	}
}

func (s *SensorHandlerMock) ID() string {
	return s.Called().String(0)
}

func (s *SensorHandlerMock) Resolution() (dsSensor.Resolution, error) {
	args := s.Called()
	return args.Get(0).(dsSensor.Resolution), args.Error(1)
}

func (s *SensorHandlerMock) SetResolution(resolution dsSensor.Resolution) error {
	return s.Called(resolution).Error(0)
}

func (s *SensorHandlerMock) PollTime() uint {
	return s.Called().Get(0).(uint)
}

func (s *SensorHandlerMock) SetPollTime(duration uint) error {
	return s.Called(duration).Error(0)
}

func (s *SensorHandlerMock) Poll(data chan dsSensor.PollData, timeMillis uint) error {
	s.TestData()["readings"] = data
	return s.Called(data, timeMillis).Error(0)
}

func (s *SensorHandlerMock) StopPoll() error {
	return s.Called().Error(0)
}

func (p *PollDataMock) ID() string {
	return p.Called().String(0)
}

func (p *PollDataMock) Temperature() float32 {
	return p.Called().Get(0).(float32)
}

func (p *PollDataMock) Stamp() time.Time {
	return p.Called().Get(0).(time.Time)
}

func (p *PollDataMock) Error() error {
	return p.Called().Error(0)
}
