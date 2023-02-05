package embedded_test

import (
	"bytes"
	"encoding/json"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type PTTestSuite struct {
	suite.Suite
	mock []*PTSensorMock
	req  *http.Request
	resp *httptest.ResponseRecorder
}

type PTSensorMock struct {
	mock.Mock
}

func (t *PTTestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
	t.mock = make([]*PTSensorMock, 1)
	t.resp = httptest.NewRecorder()
}

func TestPTTestSuite(t *testing.T) {
	suite.Run(t, new(PTTestSuite))
}

func (t *PTTestSuite) pts() []models.PTSensor {
	s := make([]models.PTSensor, len(t.mock))
	for i, m := range t.mock {
		s[i] = m
	}
	return s
}

func (t *PTTestSuite) TestPTRestAPI_ConfigSensor() {
	args := []struct {
		old, new models.PTConfig
	}{
		{
			old: models.PTConfig{
				ID:      "adin",
				Enabled: false,
				Samples: 1,
			},
			new: models.PTConfig{
				ID:      "adin",
				Enabled: true,
				Samples: 15,
			},
		},
		{
			old: models.PTConfig{
				ID:      "2",
				Enabled: true,
				Samples: 17,
			},
			new: models.PTConfig{
				ID:      "2",
				Enabled: false,
				Samples: 15,
			},
		},
	}
	t.mock = make([]*PTSensorMock, 0, len(args))
	for _, elem := range args {
		m := new(PTSensorMock)
		m.On("ID").Return(elem.old.ID)
		m.On("Config").Return(elem.new)
		m.On("SetConfig", elem.new).Return(nil)
		t.mock = append(t.mock, m)
	}
	handler, _ := embedded.New(embedded.WithPT(t.pts()))

	for i, elem := range args {
		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(elem.new)

		t.req, _ = http.NewRequest(http.MethodPut, strings.Replace(embedded.RoutesConfigPT100Sensor, ":hardware_id", elem.new.ID, 1), &body)
		t.req.Header.Add("Content-Type", "application/json")

		handler.ServeHTTP(t.resp, t.req)
		b, _ := io.ReadAll(t.resp.Body)
		var bodyJson models.PTConfig
		fromJSON(b, &bodyJson)
		t.Equal(http.StatusOK, t.resp.Code, i)
		t.EqualValues(elem.new, bodyJson, i)
	}
}

func (t *PTTestSuite) TestPTRestAPI_GetTemperatures() {
	args := []struct {
		cfg  models.PTConfig
		stat models.Temperature
	}{
		{
			cfg: models.PTConfig{
				ID:      "1",
				Enabled: true,
				Samples: 123,
			},
			stat: models.Temperature{
				ID:          "1",
				Enabled:     true,
				Temperature: 123.45,
				Stamp:       time.Unix(1, 1),
			},
		},
		{
			cfg: models.PTConfig{
				ID:      "2",
				Enabled: false,
				Samples: 12,
			},
			stat: models.Temperature{
				ID:          "2",
				Enabled:     false,
				Temperature: 11.1,
				Stamp:       time.Unix(1, 2),
			},
		},
	}
	t.mock = make([]*PTSensorMock, 0, len(args))
	for _, elem := range args {
		m := new(PTSensorMock)
		m.On("ID").Return(elem.cfg.ID)
		m.On("Config").Return(elem.cfg)
		m.On("Temperature").Return(elem.stat)
		t.mock = append(t.mock, m)
	}

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetPT100Temperatures, nil)
	h, _ := embedded.New(embedded.WithPT(t.pts()))
	h.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []models.Temperature
	fromJSON(b, &bodyJson)
	t.Equal(http.StatusOK, t.resp.Code)

	expected := make([]models.Temperature, 0, len(args))
	for _, stat := range args {
		expected = append(expected, stat.stat)
	}

	t.ElementsMatch(expected, bodyJson)

}

func (t *PTTestSuite) TestPTRestAPI_GetSensors() {
	args := []models.PTConfig{
		{
			ID:      "heyo",
			Enabled: true,
			Samples: 13,
		},
		{
			ID:      "heyo 2",
			Enabled: false,
			Samples: 17,
		},
	}
	t.mock = make([]*PTSensorMock, 0, len(args))
	for _, elem := range args {
		m := new(PTSensorMock)
		m.On("ID").Return(elem.ID)
		m.On("Config").Return(elem)
		t.mock = append(t.mock, m)
	}

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetPT100Sensors, nil)

	handler, _ := embedded.New(embedded.WithPT(t.pts()))

	handler.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []models.PTConfig
	fromJSON(b, &bodyJson)
	t.Equal(http.StatusOK, t.resp.Code)
	t.ElementsMatch(args, bodyJson)
}

func (t *PTTestSuite) TestPT_SetConfig() {
	args := []struct {
		old, new models.PTConfig
	}{
		{
			old: models.PTConfig{
				ID:      "adin",
				Enabled: false,
				Samples: 1,
			},
			new: models.PTConfig{
				ID:      "adin",
				Enabled: true,
				Samples: 15,
			},
		},
		{
			old: models.PTConfig{
				ID:      "2",
				Enabled: true,
				Samples: 17,
			},
			new: models.PTConfig{
				ID:      "2",
				Enabled: false,
				Samples: 15,
			},
		},
	}
	t.mock = make([]*PTSensorMock, 0, len(args))
	for _, elem := range args {
		m := new(PTSensorMock)
		m.On("ID").Return(elem.old.ID)
		m.On("Config").Return(elem.new)
		m.On("SetConfig", elem.new).Return(nil)

		t.mock = append(t.mock, m)
	}

	handler, _ := embedded.New(embedded.WithPT(t.pts()))
	pt := handler.PT

	for i, arg := range args {
		cfg, err := pt.SetConfig(arg.new)
		t.Nil(err, i)
		t.EqualValues(arg.new, cfg, i)
	}

}

func (t *PTTestSuite) TestPT_GetSensors() {
	args := []models.PTConfig{
		{
			ID:      "heyo",
			Enabled: true,
			Samples: 13,
		},
		{
			ID:      "heyo 2",
			Enabled: false,
			Samples: 17,
		},
	}
	t.mock = make([]*PTSensorMock, 0, len(args))
	for _, elem := range args {
		m := new(PTSensorMock)
		m.On("ID").Return(elem.ID)
		m.On("Config").Return(elem)
		t.mock = append(t.mock, m)
	}

	handler, _ := embedded.New(embedded.WithPT(t.pts()))
	pt := handler.PT
	s := pt.GetSensors()
	t.ElementsMatch(args, s)
}

func (t *PTTestSuite) TestPT_GetConfig() {
	args := []models.PTConfig{
		{
			ID:      "heyo",
			Enabled: true,
			Samples: 13,
		},
		{
			ID:      "heyo 2",
			Enabled: false,
			Samples: 17,
		},
	}
	t.mock = make([]*PTSensorMock, 0, len(args))
	for _, elem := range args {
		m := new(PTSensorMock)
		m.On("ID").Return(elem.ID)
		m.On("Config").Return(elem)
		t.mock = append(t.mock, m)
	}

	handler, _ := embedded.New(embedded.WithPT(t.pts()))
	pt := handler.PT

	s, err := pt.GetConfig(args[0].ID)
	t.Nil(err)
	t.EqualValues(args[0], s)

	s, err = pt.GetConfig("not exist")
	t.NotNil(err)
	t.ErrorIs(embedded.ErrNoSuchSensor, err)
}

func (p *PTSensorMock) ID() string {
	args := p.Called()
	return args.String(0)
}

func (p *PTSensorMock) Temperature() models.Temperature {
	args := p.Called()
	return args.Get(0).(models.Temperature)
}

func (p *PTSensorMock) Poll() error {
	args := p.Called()
	return args.Error(0)
}

func (p *PTSensorMock) StopPoll() error {
	args := p.Called()
	return args.Error(0)
}

func (p *PTSensorMock) Config() models.PTConfig {
	args := p.Called()
	return args.Get(0).(models.PTConfig)
}

func (p *PTSensorMock) SetConfig(cfg models.PTConfig) (err error) {
	args := p.Called(cfg)
	return args.Error(0)
}
