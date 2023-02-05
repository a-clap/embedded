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

func (t *GPIOTestSuite) gpios() []models.GPIO {
	gpios := make([]models.GPIO, len(t.mocks))
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
	cfg := models.GPIOConfig{
		ID:          "blah",
		Direction:   models.Input,
		ActiveLevel: models.High,
		Value:       true,
	}
	newCfg := models.GPIOConfig{
		ID:          "blah",
		Direction:   models.Input,
		ActiveLevel: models.Low,
		Value:       true,
	}
	m := new(GPIOMock)
	m.On("ID").Return(cfg.ID)
	m.On("Config").Return(newCfg, nil)
	m.On("SetConfig", newCfg).Return(nil)
	t.mocks = append(t.mocks, m)

	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(newCfg)

	t.req, _ = http.NewRequest(http.MethodPut, strings.Replace(embedded.RoutesConfigGPIO, ":hardware_id", newCfg.ID, 1), &body)
	t.req.Header.Add("Content-Type", "application/json")

	h, _ := embedded.New(embedded.WithGPIOs(t.gpios()))
	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusOK, t.resp.Code)
	t.JSONEq(toJSON(newCfg), string(b))

}
func (t *GPIOTestSuite) TestGPIO_RestAPI_GetGPIOs() {
	r := t.Require()
	args := []models.GPIOConfig{
		{
			ID:          "blah",
			Direction:   models.Input,
			ActiveLevel: models.High,
			Value:       true,
		},
		{
			ID:          "another",
			Direction:   models.Output,
			ActiveLevel: models.Low,
			Value:       false,
		},
	}
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.ID)
		m.On("Config").Return(arg, nil)
		t.mocks = append(t.mocks, m)
	}

	handler, _ := embedded.New(embedded.WithGPIOs(t.gpios()))
	r.NotNil(handler)

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetGPIOs, nil)
	handler.ServeHTTP(t.resp, t.req)

	r.Equal(http.StatusOK, t.resp.Code)

	b, _ := io.ReadAll(t.resp.Body)
	var bodyJson []models.GPIOConfig
	fromJSON(b, &bodyJson)
	r.ElementsMatch(args, bodyJson)
}

func (t *GPIOTestSuite) TestGPIO_SetConfig() {
	r := t.Require()
	args := []struct {
		cfg    models.GPIOConfig
		newCfg models.GPIOConfig
	}{
		{
			cfg: models.GPIOConfig{
				ID:          "blah",
				Direction:   models.Input,
				ActiveLevel: models.High,
				Value:       true,
			},
			newCfg: models.GPIOConfig{
				ID:          "blah",
				Direction:   models.Input,
				ActiveLevel: models.Low,
				Value:       true,
			},
		},
		{
			cfg: models.GPIOConfig{
				ID:          "another",
				Direction:   models.Output,
				ActiveLevel: models.Low,
				Value:       false,
			},
			newCfg: models.GPIOConfig{
				ID:          "another",
				Direction:   models.Input,
				ActiveLevel: models.High,
				Value:       true,
			},
		},
	}
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.cfg.ID)
		m.On("Config").Return(arg.cfg, nil)
		m.On("SetConfig", arg.newCfg).Return(nil)
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
		cfg models.GPIOConfig
	}{
		{
			cfg: models.GPIOConfig{
				ID:          "blah",
				Direction:   models.Input,
				ActiveLevel: models.High,
				Value:       true,
			},
		},
		{
			cfg: models.GPIOConfig{
				ID:          "another",
				Direction:   models.Output,
				ActiveLevel: models.Low,
				Value:       false,
			},
		},
	}
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.cfg.ID)
		m.On("Config").Return(arg.cfg, nil)
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

func (g *GPIOMock) Config() (models.GPIOConfig, error) {
	args := g.Called()
	return args.Get(0).(models.GPIOConfig), args.Error(1)
}

func (g *GPIOMock) SetConfig(cfg models.GPIOConfig) error {
	return g.Called(cfg).Error(0)
}
