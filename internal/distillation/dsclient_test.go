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
	"github.com/a-clap/iot/internal/embedded/ds18b20"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DSClientSuite struct {
	suite.Suite
}

func TestDSClient(t *testing.T) {
	suite.Run(t, new(DSClientSuite))
}

func (d *DSClientSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

func (d *DSClientSuite) Test_Temperatures() {
	t := d.Require()

	m := new(DSMock)
	onGet := []embedded.DSSensorConfig{{
		Enabled: false,
		SensorConfig: ds18b20.SensorConfig{
			ID:           "2",
			Correction:   0,
			Resolution:   13,
			PollInterval: 0,
			Samples:      0,
		}}}

	m.On("Get").Return(onGet, nil)

	h, _ := distillation.New(distillation.WithDS(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	ds := distillation.NewDSClient(srv.URL, 1*time.Second)
	s, err := ds.Temperatures()
	t.Nil(err)
	t.NotNil(s)

}

func (d *DSClientSuite) Test_Configure() {
	t := d.Require()

	m := new(DSMock)
	onGet := []embedded.DSSensorConfig{{
		Enabled: false,
		SensorConfig: ds18b20.SensorConfig{
			ID:           "2",
			Correction:   1,
			Resolution:   2,
			PollInterval: 3,
			Samples:      4,
		}}}

	m.On("Get").Return(onGet, nil)

	h, _ := distillation.New(distillation.WithDS(m))
	srv := httptest.NewServer(h)
	defer srv.Close()

	ds := distillation.NewDSClient(srv.URL, 1*time.Second)
	s, err := ds.GetSensors()
	t.Nil(err)
	t.NotNil(s)
	t.ElementsMatch([]distillation.DSConfig{{DSSensorConfig: onGet[0]}}, s)

	// Expected error - sensor doesn't exist
	_, err = ds.Configure(distillation.DSConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNoSuchID.Error())
	t.ErrorContains(err, distillation.RoutesConfigureDS)

	// Error on set now
	errSet := errors.New("hello world")
	m.On("Configure", mock.Anything).Return(embedded.DSSensorConfig{}, errSet).Once()
	_, err = ds.Configure(distillation.DSConfig{DSSensorConfig: onGet[0]})
	t.NotNil(err)
	t.ErrorContains(err, errSet.Error())

	// All good now
	onGet[0].Enabled = true
	m.On("Configure", onGet[0]).Return(onGet[0], nil).Once()
	cfg, err := ds.Configure(distillation.DSConfig{DSSensorConfig: onGet[0]})
	t.Nil(err)
	t.Equal(cfg, distillation.DSConfig{DSSensorConfig: onGet[0]})

}

func (d *DSClientSuite) Test_NotImplemented() {
	t := d.Require()
	h, _ := distillation.New()
	srv := httptest.NewServer(h)
	defer srv.Close()

	ds := distillation.NewDSClient(srv.URL, 1*time.Second)
	s, err := ds.GetSensors()
	t.Nil(s)
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetDS)

	_, err = ds.Configure(distillation.DSConfig{})
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesConfigureDS)

	temps, err := ds.Temperatures()
	t.Nil(temps)
	t.NotNil(err)
	t.ErrorContains(err, distillation.ErrNotImplemented.Error())
	t.ErrorContains(err, distillation.RoutesGetDSTemperatures)
}
