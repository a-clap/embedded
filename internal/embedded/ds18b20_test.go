/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a-clap/iot/internal/embedded"
	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DS18B20TestSuite struct {
	suite.Suite
	mock []*DS18B20SensorMock
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
	t.resp = httptest.NewRecorder()
}

func (t *DS18B20TestSuite) sensors() []embedded.DSSensor {
	sensors := make([]embedded.DSSensor, len(t.mock))
	for i, v := range t.mock {
		sensors[i] = v
	}
	return sensors
}

func (t *DS18B20TestSuite) TestRestAPI_DSConfig() {
	args := []struct {
		name     string
		old, new embedded.DSSensorConfig
	}{
		{
			name: "basic",
			old: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			},
			new: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   2,
					Resolution:   3,
					PollInterval: 4,
					Samples:      5,
				},
			},
		},
		{
			name: "enable dsSensor",
			old: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "3",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			},
			new: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: true,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "3",
					Correction:   2,
					Resolution:   3,
					PollInterval: 4,
					Samples:      5,
				},
			},
		},
	}

	t.mock = make([]*DS18B20SensorMock, len(args))
	for i, cfg := range args {
		m := new(DS18B20SensorMock)
		m.On("Name").Return(cfg.old.Bus, cfg.old.ID)
		initCall := m.On("GetConfig").Return(cfg.old.SensorConfig).Once()
		m.On("GetConfig").Return(cfg.new.SensorConfig).NotBefore(initCall).Once()
		m.On("Configure", cfg.new.SensorConfig).Return(nil).Once()
		if cfg.new.Enabled != cfg.old.Enabled {
			if cfg.new.Enabled {
				m.On("Poll", mock.Anything).Return(nil).Once()
			} else {
				m.On("Close").Return(nil).Once()
			}
		}
		t.mock[i] = m
	}
	h, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	r := t.Require()
	for _, arg := range args {
		var body bytes.Buffer
		r.Nil(json.NewEncoder(&body).Encode(arg.new))

		t.req, _ = http.NewRequest(http.MethodPut, embedded.RoutesConfigOnewireSensor, &body)
		t.req.Header.Add("Content-Type", "application/json")

		h.ServeHTTP(t.resp, t.req)
		b, _ := io.ReadAll(t.resp.Body)

		r.Equal(http.StatusOK, t.resp.Code)
		r.JSONEq(toJSON(arg.new), string(b))
	}
}

func (t *DS18B20TestSuite) TestRestAPI_GetTemperatures() {
	args := []struct {
		cfg         embedded.DSSensorConfig
		temperature embedded.DSTemperature
	}{
		{
			cfg: embedded.DSSensorConfig{
				Bus:     "456",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "ablah",
					Correction:   -123,
					Resolution:   ds18b20.Resolution11Bit,
					PollInterval: 13 * time.Microsecond,
					Samples:      15,
				},
			},
			temperature: embedded.DSTemperature{
				Bus: "456",
				Readings: []ds18b20.Readings{
					{
						ID:          "ablah",
						Temperature: 123,
						Average:     15,
						Stamp:       time.Unix(1, 1),
						Error:       "",
					},
					{
						ID:          "ablah",
						Temperature: 1,
						Average:     17,
						Stamp:       time.Unix(1, 1),
						Error:       io.EOF.Error(),
					},
				},
			},
		},
		{
			cfg: embedded.DSSensorConfig{
				Bus:     "676",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "zablah",
					Correction:   -123,
					Resolution:   ds18b20.Resolution11Bit,
					PollInterval: 13 * time.Microsecond,
					Samples:      15,
				},
			},
			temperature: embedded.DSTemperature{
				Bus: "676",
				Readings: []ds18b20.Readings{
					{
						ID:          "zablah",
						Temperature: 123,
						Average:     15,
						Stamp:       time.Unix(1, 1),
						Error:       "",
					},
					{
						ID:          "zablah",
						Temperature: 1,
						Average:     17,
						Stamp:       time.Unix(1, 1),
						Error:       errors.New("hello world").Error(),
					},
				},
			},
		},
	}

	t.mock = make([]*DS18B20SensorMock, len(args))
	var temps []embedded.DSTemperature
	for i, arg := range args {
		m := new(DS18B20SensorMock)
		m.On("GetConfig").Return(arg.cfg.SensorConfig)
		m.On("Name").Return(arg.cfg.Bus, arg.cfg.ID)
		m.On("GetReadings").Return(arg.temperature.Readings)
		temps = append(temps, arg.temperature)
		t.mock[i] = m
	}

	r := t.Require()
	h, _ := embedded.New(embedded.WithDS18B20(t.sensors()))

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetOnewireTemperatures, nil)
	h.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []embedded.DSTemperature
	fromJSON(b, &bodyJson)
	r.Equal(http.StatusOK, t.resp.Code)
	r.ElementsMatch(temps, bodyJson)

}
func (t *DS18B20TestSuite) TestRestAPI_GetSensors() {
	cfgs := []embedded.DSSensorConfig{
		{
			Bus:     "123",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "blah",
				Correction:   -13,
				Resolution:   ds18b20.Resolution12Bit,
				PollInterval: 123 * time.Microsecond,
				Samples:      5,
			},
		},
		{
			Bus:     "456",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "ablah",
				Correction:   -123,
				Resolution:   ds18b20.Resolution11Bit,
				PollInterval: 13 * time.Microsecond,
				Samples:      15,
			},
		},
	}

	t.mock = make([]*DS18B20SensorMock, len(cfgs))
	for i, cfg := range cfgs {
		m := new(DS18B20SensorMock)
		m.On("GetConfig").Return(cfg.SensorConfig)
		m.On("Name").Return(cfg.Bus, cfg.ID)
		t.mock[i] = m
	}

	h, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	r := t.Require()

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetOnewireSensors, nil)

	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []embedded.DSSensorConfig
	fromJSON(b, &bodyJson)
	r.Equal(http.StatusOK, t.resp.Code)
	r.ElementsMatch(cfgs, bodyJson)
}

func (t *DS18B20TestSuite) TestDSConfig_EnableDisable() {
	startCfg := embedded.DSSensorConfig{
		Bus:     "1",
		Enabled: false,
		SensorConfig: ds18b20.SensorConfig{
			ID:           "2",
			Correction:   1,
			Resolution:   2,
			PollInterval: 3,
			Samples:      4,
		},
	}

	enableCfg := startCfg
	enableCfg.Enabled = true

	m := new(DS18B20SensorMock)
	m.On("Name").Return(startCfg.Bus, startCfg.ID)
	initCall := m.On("GetConfig").Return(startCfg.SensorConfig).Once()
	configureCall := m.On("GetConfig").Return(enableCfg.SensorConfig).NotBefore(initCall).Once()
	m.On("GetConfig").Return(startCfg.SensorConfig).NotBefore(configureCall).Once()

	firstConfigure := m.On("Configure", enableCfg.SensorConfig).Return(nil).Once()
	m.On("Poll").Return(nil)
	m.On("Configure", startCfg.SensorConfig).Return(nil).Once().NotBefore(firstConfigure)
	m.On("Close").Return(nil)
	t.mock = nil
	t.mock = append(t.mock, m)
	mainHandler, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	ds := mainHandler.DS
	r := t.Require()

	_, err := ds.SetConfig(enableCfg)
	r.Nil(err)
	_, err = ds.SetConfig(startCfg)
	r.Nil(err)

}
func (t *DS18B20TestSuite) TestDSConfig() {
	args := []struct {
		name     string
		old, new embedded.DSSensorConfig
	}{
		{
			name: "basic",
			old: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			},
			new: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   2,
					Resolution:   3,
					PollInterval: 4,
					Samples:      5,
				},
			},
		},
		{
			name: "enable dsSensor",
			old: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "3",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			},
			new: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: true,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "3",
					Correction:   2,
					Resolution:   3,
					PollInterval: 4,
					Samples:      5,
				},
			},
		},
	}

	t.mock = make([]*DS18B20SensorMock, len(args))
	for i, cfg := range args {
		m := new(DS18B20SensorMock)
		m.On("Name").Return(cfg.old.Bus, cfg.old.ID)
		initCall := m.On("GetConfig").Return(cfg.old.SensorConfig).Once()
		m.On("GetConfig").Return(cfg.new.SensorConfig).NotBefore(initCall).Once()
		m.On("Configure", cfg.new.SensorConfig).Return(nil).Once()
		if cfg.new.Enabled != cfg.old.Enabled {
			if cfg.new.Enabled {
				m.On("Poll", mock.Anything).Return(nil).Once()
			} else {
				m.On("Close").Return(nil).Once()
			}
		}
		t.mock[i] = m
	}

	mainHandler, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	ds := mainHandler.DS
	r := t.Require()
	for _, arg := range args {
		newCfg, err := ds.SetConfig(arg.new)
		r.Nil(err, arg.name)
		r.EqualValues(arg.new, newCfg)
	}
	for _, m := range t.mock {
		m.AssertExpectations(t.T())
	}

}
func (t *DS18B20TestSuite) TestDS_GetSensors() {
	cfgs := []embedded.DSSensorConfig{
		{
			Bus:     "123",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "blah",
				Correction:   -13,
				Resolution:   ds18b20.Resolution12Bit,
				PollInterval: 123 * time.Microsecond,
				Samples:      5,
			},
		},
		{
			Bus:     "456",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "ablah",
				Correction:   -123,
				Resolution:   ds18b20.Resolution11Bit,
				PollInterval: 13 * time.Microsecond,
				Samples:      15,
			},
		},
	}

	t.mock = make([]*DS18B20SensorMock, len(cfgs))
	for i, cfg := range cfgs {
		m := new(DS18B20SensorMock)
		m.On("GetConfig").Return(cfg.SensorConfig)
		m.On("Name").Return(cfg.Bus, cfg.ID)
		t.mock[i] = m
	}

	mainHandler, _ := embedded.New(embedded.WithDS18B20(t.sensors()))
	ds := mainHandler.DS
	cfg := ds.GetSensors()
	t.NotNil(cfg)
	t.ElementsMatch(cfgs, cfg)
}

func (m *DS18B20SensorMock) Poll() (err error) {
	return m.Called().Error(0)
}

func (m *DS18B20SensorMock) Temperature() (actual, average float64, err error) {
	args := m.Called()
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

func (m *DS18B20SensorMock) Average() float64 {
	return m.Called().Get(0).(float64)
}

func (m *DS18B20SensorMock) Configure(config ds18b20.SensorConfig) error {
	return m.Called(config).Error(0)
}

func (m *DS18B20SensorMock) GetConfig() ds18b20.SensorConfig {
	return m.Called().Get(0).(ds18b20.SensorConfig)
}

func (m *DS18B20SensorMock) Close() error {
	return m.Called().Error(0)
}

func (m *DS18B20SensorMock) Name() (bus string, id string) {
	args := m.Called()
	return args.String(0), args.String(1)
}

func (m *DS18B20SensorMock) GetReadings() []ds18b20.Readings {
	return m.Called().Get(0).([]ds18b20.Readings)
}
