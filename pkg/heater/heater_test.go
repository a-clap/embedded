/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package heater_test

import (
	"io"
	"testing"
	"time"

	"github.com/a-clap/embedded/pkg/heater"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HeaterSuite struct {
	suite.Suite
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

func (t *HeaterSuite) TestHeater_Status() {
	heating := new(HeatingMock)
	ticker := new(TickerMock)
	tickerCh := make(<-chan time.Time)

	heating.On("Open").Return(nil)
	heating.On("Set", mock.Anything).Return(nil)

	ticker.On("Start", mock.Anything).Return(nil)
	ticker.On("Tick", mock.Anything).Return(tickerCh)
	ticker.On("Stop", mock.Anything)

	h, err := heater.New(heater.WithHeating(heating), heater.WithTicker(ticker))
	t.Nil(err)
	t.NotNil(h)

	t.Equal(false, h.Enabled())
	t.EqualValues(0, h.Power())

	t.Nil(h.SetPower(36))
	t.Equal(false, h.Enabled())
	t.EqualValues(36, h.Power())

	h.Enable(nil)
	t.Equal(true, h.Enabled())
	t.EqualValues(36, h.Power())

	h.Disable()
	t.Equal(false, h.Enabled())
	t.EqualValues(36, h.Power())

}

func (t *HeaterSuite) TestHeater_Power() {
	heating := new(HeatingMock)
	ticker := new(TickerMock)

	heating.On("Open").Return(nil)

	h, _ := heater.New(heater.WithHeating(heating), heater.WithTicker(ticker))
	t.Nil(h.SetPower(0))
	t.Nil(h.SetPower(71))
	t.Nil(h.SetPower(100))
	err := h.SetPower(101)
	t.NotNil(err)
	t.ErrorContains(err, heater.ErrPowerOutOfRange.Error())
}

func (t *HeaterSuite) TestHeater_ReceiveErrors() {
	heating := new(HeatingMock)
	ticker := new(TickerMock)

	tickerCh := make(chan time.Time)
	heating.On("Open").Return(nil)

	ticker.On("Start", mock.Anything)
	ticker.On("Tick", mock.Anything).Return((<-chan time.Time)(tickerCh))
	ticker.On("Stop", mock.Anything)

	h, _ := heater.New(heater.WithHeating(heating), heater.WithTicker(ticker))

	t.Nil(h.SetPower(100))

	errCh := make(chan error, 10)
	h.Enable(errCh)
	defer h.Disable()
	heating.On("Set", mock.Anything).Return(io.ErrClosedPipe)
	tickerCh <- time.Time{}
	select {
	case e := <-errCh:
		t.Require().ErrorIs(e, io.ErrClosedPipe)
	case <-time.After(5 * time.Millisecond):
		t.Fail("shouldn't be here")
	}

}

func (t *HeaterSuite) TestHeater_Running() {
	heating := new(HeatingMock)
	ticker := new(TickerMock)

	tickerCh := make(chan time.Time)
	heating.On("Open").Return(nil)
	heating.On("Set", true).Return(nil).Once()

	ticker.On("Start", mock.Anything)
	ticker.On("Tick", mock.Anything).Return((<-chan time.Time)(tickerCh))
	ticker.On("Stop", mock.Anything)

	h, _ := heater.New(heater.WithHeating(heating), heater.WithTicker(ticker))

	t.Nil(h.SetPower(37))
	h.Enable(nil)

	heating.On("Set", true).Return(nil).Times(36)
	heating.On("Set", false).Return(nil).Times(64)

	for i := 0; i < 100; i++ {
		tickerCh <- time.Now()
		<-time.After(1 * time.Millisecond)
	}

	go func() {
		heating.On("Set", true).Return(nil).Maybe()
		<-time.After(3 * time.Millisecond)
		tickerCh <- time.Now()
	}()

	h.Disable()
	if !heating.AssertExpectations(t.T()) {
		panic("not fulfilled")
	}
}

func (t *HeaterSuite) TestHeater_New() {
	r := t.Require()
	{
		// no heating interface
		ticker := new(TickerMock)
		h, err := heater.New(heater.WithTicker(ticker))
		r.Nil(h)
		r.ErrorIs(err, heater.ErrNoHeating)
	}
	{
		// no ticker interface
		heating := new(HeatingMock)
		h, err := heater.New(heater.WithHeating(heating))
		r.Nil(h)
		r.ErrorIs(err, heater.ErrNoTicker)
	}
	{
		// err on heating open interface
		heating := new(HeatingMock)
		ticker := new(TickerMock)

		heating.On("Open").Return(io.ErrClosedPipe)
		h, err := heater.New(
			heater.WithHeating(heating),
			heater.WithTicker(ticker))
		r.Nil(h)
		r.ErrorIs(err, io.ErrClosedPipe)
		r.ErrorContains(err, "Open")
		r.ErrorContains(err, "New")
	}
	{
		// err on heating open interface
		heating := new(HeatingMock)
		ticker := new(TickerMock)

		heating.On("Open").Return(io.ErrClosedPipe)
		h, err := heater.New(
			heater.WithHeating(heating),
			heater.WithTicker(ticker))
		r.Nil(h)
		r.ErrorIs(err, io.ErrClosedPipe)
		r.ErrorContains(err, "Open")
		r.ErrorContains(err, "New")
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
