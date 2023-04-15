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
	
	"github.com/a-clap/iot/pkg/embedded"
	"github.com/a-clap/iot/pkg/max31865"
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
	
	args := []struct {
		cfg      embedded.PTSensorConfig
		readings []max31865.Readings
	}{
		{
			cfg: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "heyo",
					Correction:   1.0,
					ASyncPoll:    true,
					PollInterval: 100,
					Samples:      13,
				},
			},
			readings: []max31865.Readings{
				{
					ID:          "heyo",
					Temperature: 13,
					Average:     1,
					Stamp:       time.Time{},
					Error:       "",
				},
			},
		}, {
			cfg: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "heyo 2",
					Correction:   12.0,
					ASyncPoll:    false,
					PollInterval: 101,
					Samples:      15,
				},
			},
			readings: []max31865.Readings{
				{
					ID:          "heyo 2",
					Temperature: 14,
					Average:     51,
					Stamp:       time.Time{},
					Error:       "",
				},
			},
		},
	}
	
	var sensors []embedded.PTSensor
	var readings []embedded.PTTemperature
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.cfg.ID)
		m.On("GetConfig").Return(elem.cfg.SensorConfig)
		m.On("GetReadings").Return(elem.readings)
		m.On("Configure", mock.Anything).Return(nil)
		m.On("Poll").Return(nil)
		
		readings = append(readings, embedded.PTTemperature{Readings: elem.readings})
		sensors = append(sensors, m)
	}
	
	h, _ := embedded.New(embedded.WithPT(sensors))
	// enable sensor, so we can get temperatures
	for _, elem := range args {
		cfg, err := h.PT.GetConfig(elem.cfg.ID)
		t.Nil(err)
		cfg.Enabled = true
		_, err = h.PT.SetConfig(cfg)
		t.Nil(err)
	}
	
	srv := httptest.NewServer(h)
	defer srv.Close()
	
	pt := embedded.NewPTClient(srv.URL, 1*time.Second)
	s, err := pt.Temperatures()
	t.Nil(err)
	t.NotNil(s)
	t.ElementsMatch(readings, s)
	
}

func (p *PTClientSuite) Test_Configure() {
	t := p.Require()
	
	args := []struct {
		cfg embedded.PTSensorConfig
	}{
		{
			cfg: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "heyo",
					Correction:   1.0,
					ASyncPoll:    true,
					PollInterval: 100,
					Samples:      13,
				},
			},
		}, {
			cfg: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "heyo 2",
					Correction:   12.0,
					ASyncPoll:    false,
					PollInterval: 101,
					Samples:      15,
				},
			},
		},
	}
	
	var sensors []embedded.PTSensor
	var cfgs []embedded.PTSensorConfig
	var mocks []*PTMock
	for _, elem := range args {
		m := new(PTMock)
		m.On("ID").Return(elem.cfg.ID)
		m.On("GetConfig").Return(elem.cfg.SensorConfig).Once()
		cfgs = append(cfgs, elem.cfg)
		mocks = append(mocks, m)
		sensors = append(sensors, m)
	}
	
	h, _ := embedded.New(embedded.WithPT(sensors))
	srv := httptest.NewServer(h)
	defer srv.Close()
	
	pt := embedded.NewPTClient(srv.URL, 1*time.Second)
	s, err := pt.Get()
	t.Nil(err)
	t.NotNil(s)
	t.ElementsMatch(cfgs, s)
	
	// Expected error - sensor doesn't exist
	_, err = pt.Configure(embedded.PTSensorConfig{})
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNoSuchID.Error())
	t.ErrorContains(err, embedded.RoutesConfigPT100Sensor)
	
	// Error on set now
	errSet := errors.New("hello world")
	cfgs[0].Samples = 15
	mocks[0].On("Configure", mock.Anything).Return(errSet).Once()
	_, err = pt.Configure(cfgs[0])
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())
	//
	// All good now
	cfgs[0].Samples = 23
	mocks[0].On("Configure", mock.Anything).Return(nil)
	mocks[0].On("GetConfig").Return(cfgs[0].SensorConfig).Once()
	cfg, err := pt.Configure(cfgs[0])
	t.Nil(err)
	t.Equal(cfgs[0], cfg)
	
}

func (p *PTClientSuite) Test_NotImplemented() {
	t := p.Require()
	h, _ := embedded.New()
	srv := httptest.NewServer(h)
	defer srv.Close()
	
	pt := embedded.NewPTClient(srv.URL, 1*time.Second)
	
	s, err := pt.Get()
	t.Nil(s)
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesGetPT100Sensors)
	
	_, err = pt.Configure(embedded.PTSensorConfig{})
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesConfigPT100Sensor)
	
	temps, err := pt.Temperatures()
	t.Nil(temps)
	t.NotNil(err)
	t.ErrorContains(err, embedded.ErrNotImplemented.Error())
	t.ErrorContains(err, embedded.RoutesGetPT100Temperatures)
}
