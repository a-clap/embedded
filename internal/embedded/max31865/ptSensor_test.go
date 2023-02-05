package max31865_test

import (
	"github.com/a-clap/iot/internal/embedded/avg"
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type PTSensorTestSuite struct {
	suite.Suite
	mock     *PTHandlerMock
	pollData *PollDataMock
}

type PTHandlerMock struct {
	mock.Mock
}

type PollDataMock struct {
	mock.Mock
}

func Test_PTSensor(t *testing.T) {
	suite.Run(t, new(PTSensorTestSuite))
}

func (t *PTSensorTestSuite) SetupTest() {
	t.mock = new(PTHandlerMock)
}

func (t *PTSensorTestSuite) TestPoll() {
	t.mock = new(PTHandlerMock)
	t.pollData = new(PollDataMock)
	id := "123"

	t.mock.On("ID").Return(id)
	t.mock.On("Resolution").Return(models.Resolution10BIT, nil)
	t.mock.On("Poll", mock.Anything, mock.Anything).Return(nil)
	t.mock.On("Close").Return(nil)

	t.pollData.On("ID").Return(id)

	pt, err := max31865.NewPTSensor(t.mock)
	t.Nil(err)
	err = pt.Poll()
	t.Nil(err)
	<-time.After(1 * time.Millisecond)
	// Add this point there should be any data
	stat := pt.Temperature()
	t.Equal(id, stat.ID)
	t.Equal(true, stat.Enabled)

	polledChannel := t.mock.TestData()["readings"].(chan max31865.Readings)

	temperatures := []float32{100.0, 200.0, 300.0, 400.0, 500.0}
	t.EqualValues(len(temperatures), models.DefaultSamples)

	for i, elem := range temperatures {
		now := time.Now()
		t.pollData.On("Temperature").Return(elem).Once()
		t.pollData.On("Stamp").Return(now).Once()

		polledChannel <- t.pollData
		// Just to force scheduler work
		<-time.After(1 * time.Millisecond)
		stat = pt.Temperature()
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

	err = pt.StopPoll()
	t.Nil(err)

}

func (t *PTSensorTestSuite) TestSetGetConfig() {
	args := []struct {
		newConfig models.PTConfig
		err       error
	}{
		{
			newConfig: models.PTConfig{
				ID:      "1",
				Enabled: false,
				Samples: 13,
			},
			err: nil,
		},
		{
			newConfig: models.PTConfig{
				ID:      "2",
				Enabled: false,
				Samples: 1,
			},
			err: nil,
		},
		{
			newConfig: models.PTConfig{
				ID:      "sasaaax",
				Enabled: false,
				Samples: 25,
			},
			err: nil,
		},
		{
			newConfig: models.PTConfig{
				ID:      "heeeey",
				Enabled: false,
				Samples: 0,
			},
			err: avg.ErrSizeIsZero,
		},
	}
	for _, arg := range args {
		t.mock = new(PTHandlerMock)
		t.mock.On("ID").Return(arg.newConfig.ID)

		pt, err := max31865.NewPTSensor(t.mock)
		t.NotNil(pt)
		t.Nil(err)

		err = pt.SetConfig(arg.newConfig)
		t.Equal(arg.err, err)
		if arg.err != nil {
			continue
		}

		c := pt.Config()
		t.Equal(arg.newConfig, c)

	}
}

func (m *PTHandlerMock) Poll(data chan max31865.Readings, pollTime time.Duration) (err error) {
	m.TestData()["readings"] = data
	return m.Called(data, pollTime).Error(0)
}

func (m *PTHandlerMock) Temperature() (float32, error) {
	args := m.Called()
	return args.Get(0).(float32), args.Error(1)
}

func (m *PTHandlerMock) Close() error {
	return m.Called().Error(0)
}

func (m *PTHandlerMock) ID() string {
	return m.Called().String(0)
}

func (m *PollDataMock) ID() string {
	return m.Called().String(0)
}

func (m *PollDataMock) Temperature() float32 {
	return m.Called().Get(0).(float32)
}

func (m *PollDataMock) Stamp() time.Time {
	return m.Called().Get(0).(time.Time)
}

func (m *PollDataMock) Error() error {
	return m.Called().Error(0)
}
