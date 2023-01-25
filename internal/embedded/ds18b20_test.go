package embedded_test

import (
	"github.com/a-clap/iot/internal/embedded"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

type DS18B20TestSuite struct {
	suite.Suite
	mock map[string][]*DS18B20SensorMock
}

type DS18B20SensorMock struct {
	mock.Mock
}

func TestDS18B20TestSuite(t *testing.T) {
	suite.Run(t, new(DS18B20TestSuite))
}

func (t *DS18B20TestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
	t.mock = make(map[string][]*DS18B20SensorMock)
}

func (t *DS18B20TestSuite) TearDownTest() {
	for _, v := range t.mock {
		for _, m := range v {
			m.AssertExpectations(t.T())
		}
	}
}

// ?
func (t *DS18B20TestSuite) sensors() map[embedded.OnewireBusName][]embedded.DSSensorHandler {
	sensors := make(map[embedded.OnewireBusName][]embedded.DSSensorHandler)
	for k, v := range t.mock {
		part := make([]embedded.DSSensorHandler, len(v))
		for i, s := range v {
			part[i] = s
		}
		sensors[embedded.OnewireBusName(k)] = part
	}
	return sensors
}

func (t *DS18B20TestSuite) TestDSConfig() {
	retSensors := []embedded.OnewireSensors{
		{
			Bus: "first",
			DSConfig: []embedded.DSConfig{
				{
					ID:      "first_1",
					Enabled: false,
					BusConfig: embedded.BusConfig{
						Resolution:     embedded.DS18B20Resolution_11BIT,
						PollTimeMillis: 375,
						Samples:        1,
					},
				},
				{
					ID:      "first_2",
					Enabled: false,
					BusConfig: embedded.BusConfig{
						Resolution:     embedded.DS18B20Resolution_9BIT,
						PollTimeMillis: 94,
						Samples:        1,
					},
				},
			},
		},
	}
	for _, bus := range retSensors {
		mocks := make([]*DS18B20SensorMock, len(bus.DSConfig))
		for i, cfg := range bus.DSConfig {
			mocks[i] = new(DS18B20SensorMock)
			mocks[i].On("ID").Return(cfg.ID).Once()
			mocks[i].On("Resolution").Return(cfg.Resolution, nil).Once()

		}
		t.mock[string(bus.Bus)] = mocks
	}

	mainHandler, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	ds := mainHandler.DS

	toCfg := embedded.DSConfig{
		ID:      "first_1",
		Enabled: true,
		BusConfig: embedded.BusConfig{
			Resolution:     embedded.DS18B20Resolution_10BIT,
			PollTimeMillis: 100,
			Samples:        10,
		},
	}

	t.mock["first"][0].On("SetResolution", embedded.DS18B20Resolution_10BIT).Return(nil).Once()
	t.mock["first"][0].On("SetPollTime", uint(100)).Return(nil).Once()
	t.mock["first"][0].On("Poll", mock.Anything, mock.Anything).Return(nil).Once()
	t.mock["first"][0].On("StopPoll").Return(nil).Once()

	_, err := ds.ConfigSensor(toCfg)
	t.Nil(err)
	ds.Close()

}

func (t *DS18B20TestSuite) TestDSStatus() {
	expected := []embedded.OnewireSensors{
		{
			Bus: "first",
			DSConfig: []embedded.DSConfig{
				{
					ID:      "first_1",
					Enabled: false,
					BusConfig: embedded.BusConfig{
						Resolution:     embedded.DS18B20Resolution_11BIT,
						PollTimeMillis: 375,
						Samples:        1,
					},
				},
				{
					ID:      "first_2",
					Enabled: false,
					BusConfig: embedded.BusConfig{
						Resolution:     embedded.DS18B20Resolution_9BIT,
						PollTimeMillis: 94,
						Samples:        1,
					},
				},
			},
		},
		{
			Bus: "second",
			DSConfig: []embedded.DSConfig{
				{
					ID:      "second_1",
					Enabled: false,
					BusConfig: embedded.BusConfig{
						Resolution:     embedded.DS18B20Resolution_12BIT,
						PollTimeMillis: 750,
						Samples:        1,
					},
				},
				{
					ID:      "second_2",
					Enabled: false,
					BusConfig: embedded.BusConfig{
						Resolution:     embedded.DS18B20Resolution_10BIT,
						PollTimeMillis: 188,
						Samples:        1,
					},
				},
			},
		},
	}

	for _, bus := range expected {
		mocks := make([]*DS18B20SensorMock, len(bus.DSConfig))
		for i, cfg := range bus.DSConfig {
			mocks[i] = new(DS18B20SensorMock)
			mocks[i].On("ID").Return(cfg.ID).Twice()
			mocks[i].On("Resolution").Return(cfg.Resolution, nil).Once()
		}
		t.mock[string(bus.Bus)] = mocks
	}

	mainHandler, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	ds := mainHandler.DS
	cfg, err := ds.Status()
	t.NotNil(cfg)
	t.Nil(err)
	t.ElementsMatch(expected, cfg)
}

func (m *DS18B20SensorMock) ID() string {
	return m.Called().String(0)
}

func (m *DS18B20SensorMock) Resolution() (embedded.DS18B20Resolution, error) {
	args := m.Called()
	return args.Get(0).(embedded.DS18B20Resolution), args.Error(1)
}

func (m *DS18B20SensorMock) SetResolution(resolution embedded.DS18B20Resolution) error {
	return m.Called(resolution).Error(0)
}
func (m *DS18B20SensorMock) PollTime() uint {
	return m.Called().Get(0).(uint)
}

func (m *DS18B20SensorMock) SetPollTime(duration uint) error {
	return m.Called(duration).Error(0)
}

func (m *DS18B20SensorMock) Poll(data chan embedded.TemperatureReadings, t uint) error {
	return m.Called(data, t).Error(0)
}

func (m *DS18B20SensorMock) StopPoll() error {
	return m.Called().Error(0)
}
