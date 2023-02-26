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
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PTClientSuite struct {
	suite.Suite
}

func TestPTClient(t *testing.T) {
	suite.Run(t, new(PTClientSuite))
}

func (p *PTClientSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

func (p *PTClientSuite) Test_Temperatures() {
	t := p.Require()

	m := new(PTMock)
	onGet := []embedded.PTSensorConfig{{
		Enabled: false,
		SensorConfig: max31865.SensorConfig{
			ID:           "2",
			Correction:   0,
			ASyncPoll:    false,
			PollInterval: 0,
			Samples:      0,
		}}}

	m.On("Get").Return(onGet, nil)

	h, _ := distillation.New(distillation.WithPT(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	pt := distillation.NewPTClient(srv.URL, 1*time.Second)
	s, err := pt.Temperatures()
	t.Nil(err)
	t.NotNil(s)

}

func (p *PTClientSuite) Test_Configure() {
	t := p.Require()

	m := new(PTMock)
	onGet := []embedded.PTSensorConfig{{
		Enabled: false,
		SensorConfig: max31865.SensorConfig{
			ID:           "2",
			Correction:   0,
			ASyncPoll:    false,
			PollInterval: 0,
			Samples:      0,
		}}}

	m.On("Get").Return(onGet, nil)

	h, _ := distillation.New(distillation.WithPT(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	pt := distillation.NewPTClient(srv.URL, 1*time.Second)
	s, err := pt.GetSensors()
	t.Nil(err)
	t.NotNil(s)
	t.ElementsMatch([]distillation.PTConfig{{PTSensorConfig: onGet[0]}}, s)

	// Expected error - sensor doesn't exist
	_, err = pt.Configure(distillation.PTConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNoSuchID.Error())
	t.ErrorContains(err, distillation.RoutesConfigurePT)

	// Error on set now
	errSet := errors.New("hello world")
	m.On("Configure", mock.Anything).Return(embedded.PTSensorConfig{}, errSet).Once()
	_, err = pt.Configure(distillation.PTConfig{PTSensorConfig: onGet[0]})
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())

	// All good now
	onGet[0].Enabled = true
	m.On("Configure", onGet[0]).Return(onGet[0], nil).Once()
	cfg, err := pt.Configure(distillation.PTConfig{PTSensorConfig: onGet[0]})
	t.Nil(err)
	t.Equal(cfg, distillation.PTConfig{PTSensorConfig: onGet[0]})

}

func (p *PTClientSuite) Test_NotImplemented() {
	t := p.Require()
	h, _ := distillation.New()
	srv := httptest.NewServer(h)
	defer srv.Close()

	pt := distillation.NewPTClient(srv.URL, 1*time.Second)
	s, err := pt.GetSensors()
	t.Nil(s)
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetPT)

	_, err = pt.Configure(distillation.PTConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesConfigurePT)

	temps, err := pt.Temperatures()
	t.Nil(temps)
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetPTTemperatures)
}
