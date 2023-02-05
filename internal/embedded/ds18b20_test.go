/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

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

type DS18B20TestSuite struct {
	suite.Suite
	mock map[string][]*DS18B20SensorMock
	req  *http.Request
	resp *httptest.ResponseRecorder
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
	t.resp = httptest.NewRecorder()
}

func (t *DS18B20TestSuite) sensors() map[models.OnewireBusName][]models.DSSensor {
	sensors := make(map[models.OnewireBusName][]models.DSSensor)
	for k, v := range t.mock {
		part := make([]models.DSSensor, len(v))
		for i, s := range v {
			part[i] = s
		}
		sensors[models.OnewireBusName(k)] = part
	}
	return sensors
}

func (t *DS18B20TestSuite) TestRestAPI_ConfigSensor() {
	retSensors := []models.OnewireSensors{
		{
			Bus: "first",
			DSConfig: []models.DSConfig{
				{
					ID:             "first",
					Enabled:        false,
					Resolution:     models.Resolution11BIT,
					PollTimeMillis: 375,
					Samples:        1,
				},
				{
					ID:             "second",
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
			mocks[i].On("Config").Return(cfg).Once()
			mocks[i].On("SetConfig", mock.Anything).Return(nil)
			mocks[i].On("StopPoll").Return(nil)

		}
		t.mock[string(bus.Bus)] = mocks
	}
	newCfg := models.DSConfig{
		ID:             "first",
		Enabled:        true,
		Resolution:     models.Resolution12BIT,
		PollTimeMillis: 400,
		Samples:        123,
	}

	t.mock["first"][0].On("Config").Return(newCfg)

	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(newCfg)

	t.req, _ = http.NewRequest(http.MethodPut, strings.Replace(embedded.RoutesConfigOnewireSensor, ":hardware_id", newCfg.ID, 1), &body)
	t.req.Header.Add("Content-Type", "application/json")

	h, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusOK, t.resp.Code)
	t.JSONEq(toJSON(newCfg), string(b))
}

func (t *DS18B20TestSuite) TestRestAPI_GetTemperatures() {
	args := []struct {
		bus    models.OnewireSensors
		status []models.Temperature
	}{
		{
			bus: models.OnewireSensors{
				Bus: "my awesome bus",
				DSConfig: []models.DSConfig{
					{
						ID:             "awesome sensor",
						Enabled:        true,
						Resolution:     models.Resolution12BIT,
						PollTimeMillis: 375,
						Samples:        123,
					},
					{
						ID:             "not so good, but still here it is",
						Enabled:        false,
						Resolution:     models.Resolution9BIT,
						PollTimeMillis: 94,
						Samples:        1,
					},
				},
			},
			status: []models.Temperature{
				{
					ID:          "awesome sensor readings",
					Enabled:     true,
					Temperature: 123,
					Stamp:       time.Unix(1, 1),
				},
				{
					ID:          "not so good, but still here it is readings",
					Enabled:     false,
					Temperature: -123.0,
					Stamp:       time.Unix(5, 123),
				},
			},
		},
	}

	for _, arg := range args {
		mocks := make([]*DS18B20SensorMock, len(arg.bus.DSConfig))
		if len(arg.bus.DSConfig) != len(arg.status) {
			panic("must be equal")
		}

		for i, cfg := range arg.bus.DSConfig {
			mocks[i] = new(DS18B20SensorMock)
			mocks[i].On("Config").Return(cfg)
			mocks[i].On("StopPoll").Return(nil)
		}

		for i, cfg := range arg.status {
			mocks[i].On("Temperature").Return(cfg)
		}

		t.mock[string(arg.bus.Bus)] = mocks
	}
	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetOnewireTemperatures, nil)

	h, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	h.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []models.Temperature
	fromJSON(b, &bodyJson)
	t.Equal(http.StatusOK, t.resp.Code)
	t.ElementsMatch(args[0].status, bodyJson)
}

func (t *DS18B20TestSuite) TestRestAPI_GetSensors() {
	args := []models.OnewireSensors{
		{
			Bus: "my awesome bus",
			DSConfig: []models.DSConfig{
				{
					ID:             "awesome sensor",
					Enabled:        true,
					Resolution:     models.Resolution12BIT,
					PollTimeMillis: 375,
					Samples:        123,
				},
				{
					ID:             "not so good, but still here it is",
					Enabled:        false,
					Resolution:     models.Resolution9BIT,
					PollTimeMillis: 94,
					Samples:        1,
				},
			},
		},
	}
	for _, bus := range args {
		mocks := make([]*DS18B20SensorMock, len(bus.DSConfig))
		for i, cfg := range bus.DSConfig {
			mocks[i] = new(DS18B20SensorMock)
			mocks[i].On("Config").Return(cfg)
			mocks[i].On("StopPoll").Return(nil)
		}
		t.mock[string(bus.Bus)] = mocks
	}
	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetOnewireSensors, nil)

	h, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	h.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []models.OnewireSensors
	fromJSON(b, &bodyJson)
	t.Equal(http.StatusOK, t.resp.Code)
	t.ElementsMatch(args, bodyJson)
}

func (t *DS18B20TestSuite) TestDSConfig() {
	retSensors := []models.OnewireSensors{
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

	_, err := ds.SetConfig(toCfg)
	t.Nil(err)
	ds.Close()

}

func (t *DS18B20TestSuite) TestDS_GetSensors() {
	expected := []models.OnewireSensors{
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
	cfg := ds.GetSensors()
	t.NotNil(cfg)
	t.ElementsMatch(expected, cfg)
}

func (d *DS18B20SensorMock) Temperature() models.Temperature {
	return d.Called().Get(0).(models.Temperature)
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
