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
	m := new(HeaterMock)
	onGet := []embedded.HeaterConfig{
		{
			ID:      "1",
			Enabled: false,
			Power:   13,
		},
	}
	m.On("Get").Return(onGet, nil)
	m.On("Configure", mock.Anything).Return(onGet[0], nil).Once()
	h, _ := distillation.New(distillation.WithHeaters(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	hc := distillation.NewHeaterClient(srv.URL, 1*time.Second)

	// Heater doesn't exist
	// Expected error - heater doesn't exist
	_, err := hc.Configure(distillation.HeaterConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNoSuchID.Error())
	t.ErrorContains(err, distillation.RoutesConfigureHeater)

	// Error on set
	errSet := errors.New("hello world")
	m.On("Configure", mock.Anything).Return(onGet[0], errSet).Once()
	_, err = hc.Configure(distillation.HeaterConfig{HeaterConfig: onGet[0]})
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())

	// All good
	onGet[0].Power = 17
	m.On("Configure", mock.Anything).Return(onGet[0], nil).Once()
	cfg, err := hc.Configure(distillation.HeaterConfig{HeaterConfig: onGet[0]})
	t.Nil(err)
	t.EqualValues(onGet[0], cfg.HeaterConfig)
}
func (p *HeaterClientSuite) Test_Enable() {
	t := p.Require()

	m := new(HeaterMock)
	onGet := []embedded.HeaterConfig{
		{
			ID:      "1",
			Enabled: false,
			Power:   13,
		},
	}

	m.On("Get").Return(onGet, nil)
	m.On("Configure", mock.Anything).Return(onGet[0], nil).Once()
	h, _ := distillation.New(distillation.WithHeaters(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	hc := distillation.NewHeaterClient(srv.URL, 1*time.Second)
	s, err := hc.GetAll()
	t.Nil(err)
	t.NotNil(s)
	t.ElementsMatch([]distillation.HeaterConfigGlobal{{ID: onGet[0].ID, Enabled: onGet[0].Enabled}}, s)

	// Expected error - heater doesn't exist
	_, err = hc.Enable(distillation.HeaterConfigGlobal{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNoSuchID.Error())
	t.ErrorContains(err, distillation.RoutesEnableHeater)

	// All good now
	cfg, err := hc.Enable(distillation.HeaterConfigGlobal{ID: onGet[0].ID, Enabled: true})
	t.Nil(err)
	t.Equal(distillation.HeaterConfigGlobal{ID: onGet[0].ID, Enabled: true}, cfg)

}

func (p *HeaterClientSuite) Test_NotImplemented() {
	t := p.Require()
	h, _ := distillation.New()
	srv := httptest.NewServer(h)
	defer srv.Close()

	hc := distillation.NewHeaterClient(srv.URL, 1*time.Second)

	_, err := hc.GetEnabled()
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetEnabledHeaters)

	_, err = hc.GetAll()
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetAllHeaters)

	_, err = hc.Enable(distillation.HeaterConfigGlobal{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesEnableHeater)

	_, err = hc.Configure(distillation.HeaterConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesConfigureHeater)
}
