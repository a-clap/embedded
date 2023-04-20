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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HeaterClientSuite struct {
	suite.Suite
}

func TestHeaterClient(t *testing.T) {
	suite.Run(t, new(HeaterClientSuite))
}

func (p *HeaterClientSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

func (p *HeaterClientSuite) Test_Configure() {
	t := p.Require()
	args := []embedded.HeaterConfig{
		{
			ID:      "firstHeater",
			Enabled: false,
			Power:   13,
		},
	}
	var mocks []*HeaterMock
	heaters := make(map[string]embedded.Heater, 0)
	for _, arg := range args {
		heater := new(HeaterMock)
		if arg.Enabled {
			heater.On("Enable", mock.Anything).Once()
		} else {
			heater.On("Disable").Once()
		}

		mocks = append(mocks, heater)
		heaters[arg.ID] = heater
	}
	h, _ := embedded.New(embedded.WithHeaters(heaters))
	srv := httptest.NewServer(h)
	defer srv.Close()

	hc := embedded.NewHeaterClient(srv.URL, 1*time.Second)

	// Heater doesn't exist
	// Expected error - heater doesn't exist
	_, err := hc.Configure(embedded.HeaterConfig{})
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNoSuchID.Error())
	t.ErrorContains(err, embedded.RoutesConfigHeater)

	// Error on set
	errSet := errors.New("hello world")
	args[0].Enabled = true
	mocks[0].On("SetPower", mock.Anything).Return(errSet).Once()
	_, err = hc.Configure(args[0])
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())
	t.ErrorContains(err, embedded.RoutesConfigHeater)

	// All good
	args[0].Power = 15
	mocks[0].On("SetPower", args[0].Power).Return(nil)

	if args[0].Enabled {
		mocks[0].On("Enable", mock.Anything).Once()
	} else {
		mocks[0].On("Disable").Once()
	}

	mocks[0].On("Power").Return(args[0].Power)
	mocks[0].On("Enabled").Return(args[0].Enabled)
	cfg, err := hc.Configure(args[0])
	t.Nil(err)
	t.EqualValues(args[0], cfg)
}

func (p *HeaterClientSuite) Test_NotImplemented() {
	t := p.Require()
	h, _ := embedded.New()
	srv := httptest.NewServer(h)
	defer srv.Close()

	hc := embedded.NewHeaterClient(srv.URL, 1*time.Second)

	_, err := hc.Get()
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesGetHeaters)

	_, err = hc.Configure(embedded.HeaterConfig{})
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesConfigHeater)
}
