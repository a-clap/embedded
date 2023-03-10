/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type NextTestSuite struct {
	suite.Suite
}

type ClockMock struct {
	mock.Mock
}

type SensorMock struct {
	mock.Mock
}

func TestProcess_Next(t *testing.T) {
	suite.Run(t, new(NextTestSuite))
}

func (n *NextTestSuite) TestByTemperature() {
	t := n.Require()

	retTime := int64(0)

	sn := new(SensorMock)

	// First call during constructor
	byTemp := newByTemperature(retTime, sn, 30.0, 10)
	t.NotNil(byTemp)

	// Return temperature not over threshold
	// Unix shouldn't be called
	sn.On("Temperature").Return(11.0).Once()
	t.False(byTemp.next(retTime))

	// Temperature over threshold, Unix should be called twice - once for reset, second time to check
	sn.On("Temperature").Return(30.1).Once()
	t.False(byTemp.next(retTime))

	// Temperature still over threshold
	sn.On("Temperature").Return(30.1).Once()
	// But time is not satisfied
	// Unix() should be called once - byTime is not resetted
	retTime = 9
	t.False(byTemp.next(retTime))

	// Temperature under threshold, Unix() not called
	sn.On("Temperature").Return(29.1).Once()
	t.False(byTemp.next(retTime))

	// Happy path - temperature over threshold, and time get elapsed
	sn.On("Temperature").Return(30.1).Twice()
	retTime = 100
	t.False(byTemp.next(retTime))
	retTime = 110
	t.True(byTemp.next(retTime))
}
func (n *NextTestSuite) TestByTime() {
	t := n.Require()

	retTime := int64(0)

	// First call during constructor
	byTm := newByTime(retTime, 100)

	// Call for each 'next'
	retTime = 99
	t.False(byTm.next(retTime))
	t.False(byTm.next(retTime))
	t.False(byTm.next(retTime))

	// Now should be expired
	retTime = 100
	t.True(byTm.next(retTime))

	// On reset Unix() should be called once again
	byTm.reset(retTime)

	// Now retTime was 100
	// To satisfy next, we must return 200
	retTime = 199
	t.False(byTm.next(retTime))

	// And here it goes
	retTime = 200
	t.True(byTm.next(retTime))
}

func (c *ClockMock) Unix() int64 {
	return c.Called().Get(0).(int64)
}

func (s *SensorMock) ID() string {
	return s.Called().Get(0).(string)
}

func (s *SensorMock) Temperature() float64 {
	return s.Called().Get(0).(float64)
}
