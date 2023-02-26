/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation_test

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a-clap/iot/internal/distillation"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/a-clap/iot/internal/embedded/gpio"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GPIOClientSuite struct {
	suite.Suite
}

func TestGPIOClient(t *testing.T) {
	suite.Run(t, new(GPIOClientSuite))
}

func (p *GPIOClientSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

func (p *GPIOClientSuite) Test_Configure() {
	t := p.Require()
	m := new(GPIOMock)
	onGet := []embedded.GPIOConfig{
		{
			Config: gpio.Config{
				ID:          "1",
				Direction:   gpio.DirInput,
				ActiveLevel: gpio.Low,
				Value:       false,
			},
		},
		{
			Config: gpio.Config{
				ID:          "2",
				Direction:   gpio.DirOutput,
				ActiveLevel: gpio.High,
				Value:       true,
			},
		},
	}
	m.On("Get").Return(onGet, nil)
	h, _ := distillation.New(distillation.WithGPIO(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	gp := distillation.NewGPIOClient(srv.URL, 1*time.Second)

	// GPIO doesn't exist
	// Expected error - gpio doesn't exist
	_, err := gp.Configure(distillation.GPIOConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNoSuchID.Error())
	t.ErrorContains(err, distillation.RoutesConfigureGPIO)

	// Error on set
	errSet := errors.New("hello world")
	m.On("Configure", mock.Anything).Return(onGet[0], errSet).Once()
	_, err = gp.Configure(distillation.GPIOConfig{GPIOConfig: onGet[0]})
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())

	// All good
	m.On("Configure", mock.Anything).Return(onGet[0], nil).Once()
	cfg, err := gp.Configure(distillation.GPIOConfig{GPIOConfig: onGet[0]})
	t.Nil(err)
	t.EqualValues(onGet[0], cfg.GPIOConfig)
}

func (p *GPIOClientSuite) Test_NotImplemented() {
	t := p.Require()
	h, _ := distillation.New()
	srv := httptest.NewServer(h)
	defer srv.Close()

	hc := distillation.NewGPIOClient(srv.URL, 1*time.Second)

	_, err := hc.Get()
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetGPIO)

	_, err = hc.Configure(distillation.GPIOConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesConfigureGPIO)
}
