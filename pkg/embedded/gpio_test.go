/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/a-clap/iot/pkg/embedded"
	"github.com/a-clap/iot/pkg/embedded/gpio"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GPIOTestSuite struct {
	suite.Suite
	mocks []*GPIOMock
	req   *http.Request
	resp  *httptest.ResponseRecorder
}

type GPIOMock struct {
	mock.Mock
}

func (t *GPIOTestSuite) gpios() []embedded.GPIO {
	gpios := make([]embedded.GPIO, len(t.mocks))
	for i, gpio := range t.mocks {
		gpios[i] = gpio
	}
	return gpios
}

func TestGPIOTestSuite(t *testing.T) {
	suite.Run(t, new(GPIOTestSuite))
}

func (t *GPIOTestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
	t.resp = httptest.NewRecorder()
}
func (t *GPIOTestSuite) TestGPIO_RestAPI_ConfigGPIO() {
	cfg := embedded.GPIOConfig{
		Config: gpio.Config{
			ID:          "blah",
			Direction:   gpio.DirInput,
			ActiveLevel: gpio.High,
			Value:       true,
		},
	}
	newCfg := embedded.GPIOConfig{
		Config: gpio.Config{
			ID:          "blah",
			Direction:   gpio.DirInput,
			ActiveLevel: gpio.Low,
			Value:       true,
		},
	}
	m := new(GPIOMock)
	m.On("ID").Return(cfg.ID)
	m.On("GetConfig").Return(newCfg.Config, nil)
	m.On("Configure", newCfg.Config).Return(nil)
	t.mocks = append(t.mocks, m)
	
	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(newCfg)
	
	t.req, _ = http.NewRequest(http.MethodPut, embedded.RoutesConfigGPIO, &body)
	t.req.Header.Add("Content-Type", "application/json")
	
	h, _ := embedded.New(embedded.WithGPIOs(t.gpios()))
	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)
	
	t.Equal(http.StatusOK, t.resp.Code)
	t.JSONEq(toJSON(newCfg), string(b))
	
}
func (t *GPIOTestSuite) TestGPIO_RestAPI_GetGPIOs() {
	r := t.Require()
	args := []embedded.GPIOConfig{
		{
			Config: gpio.Config{
				ID:          "blah",
				Direction:   gpio.DirInput,
				ActiveLevel: gpio.High,
				Value:       true,
			},
		},
		{
			Config: gpio.Config{
				ID:          "another",
				Direction:   gpio.DirOutput,
				ActiveLevel: gpio.Low,
				Value:       false,
			},
		},
	}
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.ID)
		m.On("GetConfig").Return(arg.Config, nil)
		t.mocks = append(t.mocks, m)
	}
	
	handler, _ := embedded.New(embedded.WithGPIOs(t.gpios()))
	r.NotNil(handler)
	
	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetGPIOs, nil)
	handler.ServeHTTP(t.resp, t.req)
	
	r.Equal(http.StatusOK, t.resp.Code)
	
	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []embedded.GPIOConfig
	fromJSON(b, &bodyJson)
	r.ElementsMatch(args, bodyJson)
}

func (t *GPIOTestSuite) TestGPIO_SetConfig() {
	r := t.Require()
	args := []struct {
		cfg    embedded.GPIOConfig
		newCfg embedded.GPIOConfig
	}{
		{
			cfg: embedded.GPIOConfig{
				Config: gpio.Config{
					ID:          "blah",
					Direction:   gpio.DirInput,
					ActiveLevel: gpio.High,
					Value:       true,
				},
			},
			newCfg: embedded.GPIOConfig{
				Config: gpio.Config{
					ID:          "blah",
					Direction:   gpio.DirInput,
					ActiveLevel: gpio.Low,
					Value:       true,
				},
			},
		},
		{
			cfg: embedded.GPIOConfig{
				Config: gpio.Config{
					ID:          "another",
					Direction:   gpio.DirOutput,
					ActiveLevel: gpio.Low,
					Value:       false,
				},
			},
			newCfg: embedded.GPIOConfig{
				Config: gpio.Config{
					ID:          "another",
					Direction:   gpio.DirInput,
					ActiveLevel: gpio.High,
					Value:       true,
				},
			},
		},
	}
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.cfg.ID)
		m.On("GetConfig").Return(arg.cfg.Config, nil)
		m.On("Configure", arg.newCfg.Config).Return(nil)
		t.mocks = append(t.mocks, m)
	}
	
	handler, _ := embedded.New(embedded.WithGPIOs(t.gpios()))
	r.NotNil(handler)
	
	gpio := handler.GPIO
	for i, arg := range args {
		cfg, _ := gpio.GetConfig(arg.cfg.ID)
		r.EqualValues(arg.cfg, cfg, i)
		r.Nil(gpio.SetConfig(arg.newCfg))
	}
}

func (t *GPIOTestSuite) TestGPIO_GetConfig() {
	r := t.Require()
	args := []struct {
		cfg embedded.GPIOConfig
	}{
		{
			cfg: embedded.GPIOConfig{
				Config: gpio.Config{
					ID:          "blah",
					Direction:   gpio.DirInput,
					ActiveLevel: gpio.High,
					Value:       true,
				},
			},
		},
		{
			cfg: embedded.GPIOConfig{
				Config: gpio.Config{
					ID:          "another",
					Direction:   gpio.DirOutput,
					ActiveLevel: gpio.Low,
					Value:       false,
				},
			},
		},
	}
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.cfg.ID)
		m.On("GetConfig").Return(arg.cfg.Config, nil)
		t.mocks = append(t.mocks, m)
	}
	
	handler, _ := embedded.New(embedded.WithGPIOs(t.gpios()))
	r.NotNil(handler)
	
	gpio := handler.GPIO
	for i, arg := range args {
		cfg, _ := gpio.GetConfig(arg.cfg.ID)
		r.EqualValues(arg.cfg, cfg, i)
	}
}

func (g *GPIOMock) ID() string {
	return g.Called().String(0)
}

func (g *GPIOMock) GetConfig() (gpio.Config, error) {
	args := g.Called()
	return args.Get(0).(gpio.Config), args.Error(1)
}

func (g *GPIOMock) Configure(cfg gpio.Config) error {
	return g.Called(cfg).Error(0)
}

func (g *GPIOMock) Get() (bool, error) {
	args := g.Called()
	return args.Bool(0), args.Error(1)
}
