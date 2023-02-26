/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation_test

import (
	"testing"

	"github.com/a-clap/iot/internal/embedded"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GPIOTestSuite struct {
	suite.Suite
}

type GPIOMock struct {
	mock.Mock
}

func (g *GPIOMock) Get() ([]embedded.GPIOConfig, error) {
	args := g.Called()
	return args.Get(0).([]embedded.GPIOConfig), args.Error(1)
}

func (g *GPIOMock) Configure(c embedded.GPIOConfig) (embedded.GPIOConfig, error) {
	args := g.Called(c)
	return args.Get(0).(embedded.GPIOConfig), args.Error(1)
}

func TestGPIOSuite(t *testing.T) {
	suite.Run(t, new(GPIOTestSuite))
}
