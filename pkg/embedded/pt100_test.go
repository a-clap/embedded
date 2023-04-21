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
	
	"github.com/a-clap/embedded/pkg/embedded"
	"github.com/a-clap/embedded/pkg/max31865"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PTTestSuite struct {
	suite.Suite
	mock []*PTMock
	req  *http.Request
	resp *httptest.ResponseRecorder
}

// PTMock implements embedded.PTSensor
type PTMock struct {
	mock.Mock
}

func (t *PTTestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
	t.mock = make([]*PTMock, 1)
	t.resp = httptest.NewRecorder()
}

func TestPTTestSuite(t *testing.T) {
	suite.Run(t, new(PTTestSuite))
}

func (t *PTTestSuite) pts() []embedded.PTSensor {
	s := make([]embedded.PTSensor, len(t.mock))
	for i, m := range t.mock {
		s[i] = m
	}
	return s
}

func (t *PTTestSuite) TestPTRestAPI_ConfigSensor() {
	args := []struct {
		old, new embedded.PTSensorConfig
	}{
		{
			old: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "blah",
					Correction:   13.0,
					ASyncPoll:    false,
					PollInterval: 100,
					Samples:      1,
				},
			},
			new: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "blah",
					Correction:   14.0,
					ASyncPoll:    false,
					PollInterval: 101,
					Samples:      15,
				},
			},
		},
		{
			old: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "another",
					Correction:   -14.0,
					ASyncPoll:    false,
					PollInterval: 100,
					Samples:      1,
				},
			},
			new: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "another",
					Correction:   14.0,
					ASyncPoll:    true,
					PollInterval: 101,
					Samples:      15,
				},
			},
		},
	}
	t.mock = make([]*PTMock, 0, len(args))
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.old.ID)
		m.On("GetConfig").Return(elem.new.SensorConfig)
		m.On("Configure", elem.new.SensorConfig).Return(nil)
		
		t.mock = append(t.mock, m)
	}
	
	handler, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	
	for i, elem := range args {
		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(elem.new)
		
		t.req, _ = http.NewRequest(http.MethodPut, embedded.RoutesConfigPT100Sensor, &body)
		t.req.Header.Add("Content-Type", "application/json")
		
		handler.Router.ServeHTTP(t.resp, t.req)
		b, _ := io.ReadAll(t.resp.Body)
		var bodyJson embedded.PTSensorConfig
		fromJSON(b, &bodyJson)
		t.Equal(http.StatusOK, t.resp.Code, i)
		t.EqualValues(elem.new, bodyJson, i)
	}
}

func (t *PTTestSuite) TestPTRestAPI_GetTemperatures() {
	args := []struct {
		cfg  embedded.PTSensorConfig
		stat embedded.PTTemperature
	}{
		{
			cfg: embedded.PTSensorConfig{
				Enabled: true,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   10.0,
					ASyncPoll:    true,
					PollInterval: 110,
					Samples:      15,
				},
			},
			
			stat: embedded.PTTemperature{
				Readings: []max31865.Readings{
					{
						ID:          "1",
						Temperature: 12,
						Average:     13,
						Stamp:       time.Unix(1, 1),
						Error:       "nil",
					},
					{
						ID:          "1",
						Temperature: 12.3,
						Average:     13.0,
						Stamp:       time.Unix(1, 15),
						Error:       errors.New("helloworld").Error(),
					},
				},
			},
		},
	}
	t.mock = make([]*PTMock, 0, len(args))
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.cfg.ID)
		m.On("GetConfig").Return(elem.cfg.SensorConfig)
		m.On("GetReadings").Return(elem.stat.Readings)
		m.On("Configure", mock.Anything).Return(nil)
		m.On("Poll").Return(nil)
		
		t.mock = append(t.mock, m)
	}
	
	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetPT100Temperatures, nil)
	h, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	// enable sensor, so we can get temperatures
	for _, elem := range args {
		cfg, err := h.PT.GetConfig(elem.cfg.ID)
		t.Nil(err)
		cfg.Enabled = true
		_, err = h.PT.SetConfig(cfg)
		t.Nil(err)
	}
	
	h.Router.ServeHTTP(t.resp, t.req)
	
	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []embedded.PTTemperature
	fromJSON(b, &bodyJson)
	t.Equal(http.StatusOK, t.resp.Code)
	
	expected := make([]embedded.PTTemperature, 0, len(args))
	for _, stat := range args {
		expected = append(expected, stat.stat)
	}
	
	t.ElementsMatch(expected, bodyJson)
	
}
func (t *PTTestSuite) TestPTRestAPI_GetSensors() {
	args := []embedded.PTSensorConfig{
		{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "heyo",
				Correction:   1.0,
				ASyncPoll:    true,
				PollInterval: 100,
				Samples:      13,
			},
		},
		{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "heyo 2",
				Correction:   12.0,
				ASyncPoll:    false,
				PollInterval: 101,
				Samples:      15,
			},
		},
	}
	t.mock = make([]*PTMock, 0, len(args))
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.ID)
		m.On("GetConfig").Return(elem.SensorConfig)
		t.mock = append(t.mock, m)
	}
	
	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetPT100Sensors, nil)
	
	handler, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	
	handler.Router.ServeHTTP(t.resp, t.req)
	
	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []embedded.PTSensorConfig
	fromJSON(b, &bodyJson)
	t.Equal(http.StatusOK, t.resp.Code)
	t.ElementsMatch(args, bodyJson)
}

func (t *PTTestSuite) TestPT_SetConfig_EnableDisable() {
	disabled := embedded.PTSensorConfig{
		Enabled: false,
		SensorConfig: max31865.SensorConfig{
			ID:           "blah",
			Correction:   13.0,
			ASyncPoll:    false,
			PollInterval: 100,
			Samples:      1,
		},
	}
	
	enabled := embedded.PTSensorConfig{
		Enabled: true,
		SensorConfig: max31865.SensorConfig{
			ID:           "blah",
			Correction:   13.0,
			ASyncPoll:    false,
			PollInterval: 100,
			Samples:      1,
		},
	}
	
	t.mock = make([]*PTMock, 0, 1)
	m := new(PTMock)
	m.On("ID").Return(disabled.ID)
	m.On("GetConfig").Return(disabled.SensorConfig)
	m.On("Configure", enabled.SensorConfig).Return(nil)
	t.mock = append(t.mock, m)
	
	handler, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	pt := handler.PT
	
	// Poll should be called
	m.On("Poll").Return(nil)
	cfg, err := pt.SetConfig(enabled)
	t.Nil(err)
	t.EqualValues(enabled, cfg)
	
	// Close should be called
	m.On("Close").Return(nil)
	cfg, err = pt.SetConfig(disabled)
	t.Nil(err)
	t.EqualValues(disabled, cfg)
	
}
func (t *PTTestSuite) TestPT_SetConfig() {
	args := []struct {
		old, new embedded.PTSensorConfig
	}{
		{
			old: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "blah",
					Correction:   13.0,
					ASyncPoll:    false,
					PollInterval: 100,
					Samples:      1,
				},
			},
			new: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "blah",
					Correction:   14.0,
					ASyncPoll:    false,
					PollInterval: 101,
					Samples:      15,
				},
			},
		},
		{
			old: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "another",
					Correction:   -14.0,
					ASyncPoll:    false,
					PollInterval: 100,
					Samples:      1,
				},
			},
			new: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "another",
					Correction:   14.0,
					ASyncPoll:    true,
					PollInterval: 101,
					Samples:      15,
				},
			},
		},
	}
	t.mock = make([]*PTMock, 0, len(args))
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.old.ID)
		m.On("GetConfig").Return(elem.new.SensorConfig)
		m.On("Configure", elem.new.SensorConfig).Return(nil)
		
		t.mock = append(t.mock, m)
	}
	
	handler, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	pt := handler.PT
	
	for i, arg := range args {
		cfg, err := pt.SetConfig(arg.new)
		t.Nil(err, i)
		t.EqualValues(arg.new, cfg, i)
	}
	
}
func (t *PTTestSuite) TestPT_GetSensors() {
	args := []embedded.PTSensorConfig{
		{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "heyo",
				Correction:   1.0,
				ASyncPoll:    true,
				PollInterval: 100,
				Samples:      13,
			},
		},
		{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "heyo 2",
				Correction:   12.0,
				ASyncPoll:    false,
				PollInterval: 101,
				Samples:      15,
			},
		},
	}
	t.mock = make([]*PTMock, 0, len(args))
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.ID)
		m.On("GetConfig").Return(elem.SensorConfig)
		t.mock = append(t.mock, m)
	}
	
	handler, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	pt := handler.PT
	s := pt.GetSensors()
	t.ElementsMatch(args, s)
}

func (t *PTTestSuite) TestPT_GetConfig() {
	args := []embedded.PTSensorConfig{
		{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "heyo",
				Correction:   1.0,
				ASyncPoll:    true,
				PollInterval: 100,
				Samples:      13,
			},
		},
		{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "heyo 2",
				Correction:   12.0,
				ASyncPoll:    false,
				PollInterval: 101,
				Samples:      15,
			},
		},
	}
	t.mock = make([]*PTMock, 0, len(args))
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.ID)
		m.On("GetConfig").Return(elem.SensorConfig)
		t.mock = append(t.mock, m)
	}
	
	handler, _ := embedded.NewRest(embedded.WithPT(t.pts()))
	pt := handler.PT
	for _, elem := range args {
		s, err := pt.GetConfig(elem.ID)
		t.Nil(err)
		t.EqualValues(elem, s)
	}
	
	_, err := pt.GetConfig("not exist")
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNoSuchID.Error())
}

func (p *PTMock) ID() string {
	args := p.Called()
	return args.String(0)
}

func (p *PTMock) Poll() (err error) {
	args := p.Called()
	return args.Error(0)
}

func (p *PTMock) Configure(config max31865.SensorConfig) error {
	args := p.Called(config)
	return args.Error(0)
}

func (p *PTMock) GetConfig() max31865.SensorConfig {
	args := p.Called()
	return args.Get(0).(max31865.SensorConfig)
}

func (p *PTMock) Average() float64 {
	args := p.Called()
	return args.Get(0).(float64)
}

func (p *PTMock) Temperature() (actual float64, average float64, err error) {
	args := p.Called()
	return args.Get(0).(float64), args.Get(1).(float64), args.Error(2)
}

func (p *PTMock) GetReadings() []max31865.Readings {
	args := p.Called()
	return args.Get(0).([]max31865.Readings)
}

func (p *PTMock) Close() error {
	args := p.Called()
	return args.Error(0)
}
