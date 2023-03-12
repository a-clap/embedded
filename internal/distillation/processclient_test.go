/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation_test

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a-clap/iot/internal/distillation"
	"github.com/a-clap/iot/internal/distillation/process"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProcessClientSuite struct {
	suite.Suite
}

func TestProcessClient(t *testing.T) {
	suite.Run(t, new(ProcessClientSuite))
}

func (p *ProcessClientSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

type ProcessHeaterMock struct {
	mock.Mock
}
type ProcessSensorMock struct {
	mock.Mock
}

func (s *ProcessSensorMock) ID() string {
	return s.Called().String(0)
}

func (s *ProcessSensorMock) Temperature() float64 {
	return s.Called().Get(0).(float64)
}

func (p *ProcessHeaterMock) ID() string {
	return p.Called().String(0)
}

func (p *ProcessHeaterMock) SetPower(pwr int) error {
	return p.Called(pwr).Error(0)
}

func (p *ProcessClientSuite) Test_Config() {
	t := p.Require()
	h, _ := distillation.New()

	sensorMock := new(ProcessSensorMock)
	sensorMock.On("ID").Return("s1")
	sensors := []process.Sensor{sensorMock}

	heaterMock := new(ProcessHeaterMock)
	heaterMock.On("ID").Return("h1")
	heaters := []process.Heater{heaterMock}

	h.Process.ConfigureHeaters(heaters)
	h.Process.ConfigureSensors(sensors)
	srv := httptest.NewServer(h)
	defer srv.Close()

	ps := distillation.NewProcessClient(srv.URL, 1*time.Second)

	// Get anything
	cfg, err := ps.GetPhaseConfig(0)
	t.Nil(err)
	t.NotNil(cfg)

	// Correct config
	cfg.Heaters = []process.HeaterPhaseConfig{
		{
			ID:    "h1",
			Power: 0,
		},
	}
	cfg.Next.SecondsToMove = 1

	newCfg, err := ps.ConfigurePhase(0, cfg)
	t.Nil(err)
	t.NotNil(newCfg)
	t.Equal(cfg, newCfg)

	// Ask for not existing phase
	newCfg, err = ps.ConfigurePhase(3, cfg)
	t.NotNil(err)
	t.ErrorContains(err, distillation.RoutesProcessConfigPhase)
	t.ErrorContains(err, process.ErrNoSuchPhase.Error())

}

func (p *ProcessClientSuite) Test_PhaseCount() {
	t := p.Require()
	h, _ := distillation.New()
	srv := httptest.NewServer(h)
	defer srv.Close()

	ps := distillation.NewProcessClient(srv.URL, 1*time.Second)
	s, err := ps.GetPhaseCount()
	t.Nil(err)
	t.NotNil(s)
	// Initial value
	t.EqualValues(3, s.PhaseNumber)

	// Good
	s.PhaseNumber = 5
	s, err = ps.ConfigurePhaseCount(s)
	t.Nil(err)
	t.EqualValues(5, s.PhaseNumber)

	// Wrong - phases can't be below or equal 0
	s.PhaseNumber = 0
	_, err = ps.ConfigurePhaseCount(s)
	t.NotNil(err)
	t.ErrorContains(err, process.ErrPhasesBelowZero.Error())
	t.ErrorContains(err, distillation.RoutesProcessPhases)

	// Wrong - phases can't be below or equal 0
	s.PhaseNumber = -1
	_, err = ps.ConfigurePhaseCount(s)
	t.NotNil(err)
	t.ErrorContains(err, process.ErrPhasesBelowZero.Error())
	t.ErrorContains(err, distillation.RoutesProcessPhases)

	// So there should be still 5 phases
	s, err = ps.GetPhaseCount()
	t.Nil(err)
	t.NotNil(s)
	t.EqualValues(5, s.PhaseNumber)
}
