package embedded_test

import (
	"github.com/a-clap/iot/internal/embedded"
	"github.com/a-clap/iot/internal/embedded/models"
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

func (t *DS18B20TestSuite) sensors() map[embedded.OnewireBusName][]models.DSSensor {
	sensors := make(map[embedded.OnewireBusName][]models.DSSensor)
	for k, v := range t.mock {
		part := make([]models.DSSensor, len(v))
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
			DSConfig: []models.DSConfig{
				{
					ID:             "first_1",
					Enabled:        false,
					Resolution:     models.Resolution11BIT,
					PollTimeMillis: 375,
					Samples:        1,
				},
				{
					ID:             "first_2",
					Enabled:        false,
					Resolution:     models.Resolution9BIT,
					PollTimeMillis: 94,
					Samples:        1,
				},
			},
		},
	}
	for _, bus := range retSensors {
		mocks := make([]*DS18B20SensorMock, len(bus.DSConfig))
		for i, cfg := range bus.DSConfig {
			mocks[i] = new(DS18B20SensorMock)
			mocks[i].On("Config").Return(cfg)
			mocks[i].On("StopPoll").Return(nil)
		}
		t.mock[string(bus.Bus)] = mocks
	}

	mainHandler, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	ds := mainHandler.DS

	toCfg := models.DSConfig{
		ID:             "first_1",
		Enabled:        true,
		Resolution:     models.Resolution10BIT,
		PollTimeMillis: 100,
		Samples:        10,
	}

	t.mock["first"][0].On("SetConfig", toCfg).Return(nil).Once()

	_, err := ds.ConfigSensor(toCfg)
	t.Nil(err)
	ds.Close()

}

func (t *DS18B20TestSuite) TestDSStatus() {
	expected := []embedded.OnewireSensors{
		{
			Bus: "first",
			DSConfig: []models.DSConfig{
				{
					ID:             "first_1",
					Enabled:        false,
					Resolution:     models.Resolution11BIT,
					PollTimeMillis: 375,
					Samples:        5,
				},
				{
					ID:             "first_2",
					Enabled:        false,
					Resolution:     models.Resolution9BIT,
					PollTimeMillis: 94,
					Samples:        5,
				},
			},
		},
		{
			Bus: "second",
			DSConfig: []models.DSConfig{
				{
					ID:             "second_1",
					Enabled:        false,
					Resolution:     models.Resolution12BIT,
					PollTimeMillis: 750,
					Samples:        5,
				},
				{
					ID:             "second_2",
					Enabled:        false,
					Resolution:     models.Resolution10BIT,
					PollTimeMillis: 188,
					Samples:        5,
				},
			},
		},
	}

	for _, bus := range expected {
		mocks := make([]*DS18B20SensorMock, len(bus.DSConfig))
		for i, cfg := range bus.DSConfig {
			mocks[i] = new(DS18B20SensorMock)
			mocks[i].On("Config").Return(cfg)
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

func (d *DS18B20SensorMock) Status() models.DSStatus {
	return d.Called().Get(0).(models.DSStatus)
}

func (d *DS18B20SensorMock) Poll() error {
	return d.Called().Error(0)
}

func (d *DS18B20SensorMock) StopPoll() error {
	return d.Called().Error(0)
}

func (d *DS18B20SensorMock) Config() models.DSConfig {
	return d.Called().Get(0).(models.DSConfig)
}

func (d *DS18B20SensorMock) SetConfig(cfg models.DSConfig) (err error) {
	return d.Called(cfg).Error(0)
}
