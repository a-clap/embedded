/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package heater_test

import (
	"github.com/a-clap/iot/internal/embedded/heater"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type HeaterSuite struct {
	suite.Suite
	heating *HeatingMock
	ticker  *TickerMock
}

type HeatingMock struct {
	mock.Mock
}
type TickerMock struct {
	mock.Mock
}

func TestHeaterSuite(t *testing.T) {
	suite.Run(t, new(HeaterSuite))
}

func (t *HeaterSuite) SetupTest() {
	t.heating = new(HeatingMock)
	t.ticker = new(TickerMock)

}

func (t *HeaterSuite) TestHeater_Status() {
	ticker := make(<-chan time.Time)
	t.heating.On("Open").Return(nil)
	t.heating.On("Set", mock.Anything).Return(nil)

	t.ticker.On("Start", mock.Anything).Return(nil)
	t.ticker.On("Tick", mock.Anything).Return(ticker)
	t.ticker.On("Stop", mock.Anything)

	h, err := heater.New(heater.WithHeating(t.heating), heater.WithTicker(t.ticker))
	t.Nil(err)
	t.NotNil(h)

	t.Equal(false, h.Enabled())
	t.EqualValues(0, h.Power())

	t.Nil(h.SetPower(36))
	t.Equal(false, h.Enabled())
	t.EqualValues(36, h.Power())

	h.Enable(true)
	t.Equal(true, h.Enabled())
	t.EqualValues(36, h.Power())

	h.Enable(false)
	t.Equal(false, h.Enabled())
	t.EqualValues(36, h.Power())

}

func (t *HeaterSuite) TestHeater_Power() {
	t.heating.On("Open").Return(nil)

	h, _ := heater.New(heater.WithHeating(t.heating), heater.WithTicker(t.ticker))
	t.Nil(h.SetPower(0))
	t.Nil(h.SetPower(71))
	t.Nil(h.SetPower(100))
	err := h.SetPower(101)
	t.NotNil(err)
	t.ErrorIs(heater.ErrPowerOutOfRange, err)
}
func (t *HeaterSuite) TestHeater_Running() {
	ticker := make(chan time.Time)
	t.heating.On("Open").Return(nil)

	t.heating.On("Set", true).Return(nil).Once()

	t.ticker.On("Start", mock.Anything)
	t.ticker.On("Tick", mock.Anything).Return((<-chan time.Time)(ticker))
	t.ticker.On("Stop", mock.Anything)

	h, _ := heater.New(heater.WithHeating(t.heating), heater.WithTicker(t.ticker))

	t.Nil(h.SetPower(37))
	h.Enable(true)

	t.heating.On("Set", true).Return(nil).Times(37)
	t.heating.On("Set", false).Return(nil).Times(63)

	for i := 0; i < 100; i++ {
		ticker <- time.Now()
		<-time.After(1 * time.Millisecond)
	}

	go func() {
		t.heating.On("Set", true).Return(nil).Maybe()
		<-time.After(3 * time.Millisecond)
		ticker <- time.Now()
	}()

	h.Enable(false)
	if !t.heating.AssertExpectations(t.T()) {
		panic("not fulfilled")
	}
}
func (h *HeatingMock) Open() error {
	args := h.Called()
	return args.Error(0)
}

func (h *HeatingMock) Set(b bool) error {
	args := h.Called(b)
	return args.Error(0)
}

func (t *TickerMock) Start(d time.Duration) {
	t.Called(d)
}

func (t *TickerMock) Stop() {
	t.Called()
}

func (t *TickerMock) Tick() <-chan time.Time {
	args := t.Called()
	return args.Get(0).(<-chan time.Time)
}
