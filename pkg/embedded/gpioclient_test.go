/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded_test

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"
	
	"github.com/a-clap/embedded/pkg/embedded"
	"github.com/a-clap/embedded/pkg/gpio"
	"github.com/gin-gonic/gin"
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
	var mocks []*GPIOMock
	gpios := make([]embedded.GPIO, 0)
	for _, arg := range args {
		m := new(GPIOMock)
		m.On("ID").Return(arg.ID)
		m.On("GetConfig").Return(arg.Config, nil)
		gpios = append(gpios, m)
		mocks = append(mocks, m)
	}
	h, _ := embedded.New(embedded.WithGPIOs(gpios))
	srv := httptest.NewServer(h)
	defer srv.Close()
	
	hc := embedded.NewGPIOClient(srv.URL, 1*time.Second)
	
	// Heater doesn't exist
	// Expected error - heater doesn't exist
	_, err := hc.Configure(embedded.GPIOConfig{})
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNoSuchID.Error())
	t.ErrorContains(err, embedded.RoutesConfigGPIO)
	
	// Error on set
	errSet := errors.New("hello world")
	args[0].ActiveLevel = gpio.High
	mocks[0].On("Configure", args[0].Config).Return(errSet).Once()
	_, err = hc.Configure(args[0])
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())
	t.ErrorContains(err, embedded.RoutesConfigGPIO)
	
	// All good
	mocks[0].On("Configure", args[0].Config).Return(nil).Once()
	cfg, err := hc.Configure(args[0])
	t.Nil(err)
	t.EqualValues(args[0], cfg)
}

func (p *GPIOClientSuite) Test_NotImplemented() {
	t := p.Require()
	h, _ := embedded.New()
	srv := httptest.NewServer(h)
	defer srv.Close()
	
	hc := embedded.NewGPIOClient(srv.URL, 1*time.Second)
	
	_, err := hc.Get()
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesGetGPIOs)
	
	_, err = hc.Configure(embedded.GPIOConfig{})
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesConfigGPIO)
}
