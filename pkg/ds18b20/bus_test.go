/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20_test

import (
	"errors"
	"io"
	"path"
	"testing"

	"github.com/a-clap/embedded/pkg/ds18b20"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type OnewireMock struct {
	mock.Mock
}

type FileMock struct {
	mock.Mock
}

type BusSuite struct {
	suite.Suite
}

func TestDS8B20Run(t *testing.T) {
	suite.Run(t, new(BusSuite))
}

func (t *BusSuite) TestBus_IDs() {
	args := []struct {
		name   string
		path   string
		err    error
		slaves []byte
		ids    []string
	}{
		{
			name:   "handle interface error",
			path:   "expectedPath",
			err:    errors.New("interface error"),
			slaves: nil,
			ids:    nil,
		},
		{
			name:   "no devices on bus",
			path:   "new temperaturePath",
			err:    nil,
			slaves: nil,
			ids:    nil,
		},
		{
			name:   "single device on bus",
			path:   "another temperaturePath",
			err:    nil,
			slaves: []byte{50, 56, 45, 48, 53, 49, 54, 57, 51, 56, 52, 56, 100, 102, 102, 10},
			ids:    []string{"28-051693848dff"},
		},
		{
			name: "ignore other files",
			path: "another temperaturePath",
			err:  nil,
			slaves: []byte{50, 56, 45, 48, 53, 49, 54, 57, 51, 56, 52, 56, 100, 102, 102, 10, 50, 56, 45, 48, 53, 49,
				54, 57, 51, 57, 55, 97, 101, 102, 102, 10},
			ids: []string{"28-051693848dff", "28-05169397aeff"},
		},
	}
	r := t.Require()
	for _, arg := range args {
		onewire := new(OnewireMock)
		onewire.On("Path").Return(arg.path)
		onewire.On("ReadFile", path.Join(arg.path, "w1_master_slaves")).Return(arg.slaves, arg.err)

		h, err := ds18b20.NewBus(
			ds18b20.WithInterface(onewire),
		)

		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		ids, err := h.IDs()
		r.EqualValues(arg.ids, ids, arg.name)

		if arg.err != nil {
			r.ErrorContains(err, arg.err.Error(), arg.name)
		} else {
			r.Nil(err, arg.name)
		}
		onewire.AssertExpectations(t.T())
	}
}

func (t *BusSuite) TestBus_NewSensorGood() {
	id := "28-05169397aeff"
	resolutionBuf := []byte("11")
	w1path := "/sys/bus/w1/devices/w1_bus_master1"

	onewire := new(OnewireMock)
	onewire.On("Path").Return(w1path)
	onewire.On("ReadFile", path.Join(w1path, id, "resolution")).Return(resolutionBuf, nil).Once()

	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(onewire),
	)

	sensor, err := bus.NewSensor(id)

	t.NotNil(sensor)
	t.Nil(err)
	sid := sensor.ID()
	t.EqualValues(id, sid)

	onewire.AssertExpectations(t.T())
}

func (t *BusSuite) TestBus_NewSensorIDsError() {
	onewire := new(OnewireMock)

	r := t.Require()
	bus, err := ds18b20.NewBus(ds18b20.WithInterface(onewire))
	r.Nil(err)
	r.NotNil(bus)

	w1Path := "error is coming"
	internal := io.ErrShortBuffer
	onewire.On("Path").Return(w1Path)

	onewire.On("ReadFile", mock.Anything).Return([]byte{}, internal)
	ids, err := bus.NewSensor("hello")
	r.Nil(ids)
	r.NotNil(err)
	r.ErrorIs(err, internal)
	r.ErrorContains(err, w1Path)
	r.ErrorContains(err, "resolution")
	r.ErrorContains(err, "NewSensor")
}

func (t *BusSuite) TestBus_IDSError() {
	onewire := new(OnewireMock)

	r := t.Require()
	bus, err := ds18b20.NewBus(ds18b20.WithInterface(onewire))
	r.Nil(err)
	r.NotNil(bus)

	w1Path := "error is coming"
	internal := io.ErrShortBuffer
	onewire.On("Path").Return(w1Path)
	onewire.On("ReadFile", path.Join(w1Path, "w1_master_slaves")).Return([]byte{}, internal).Once()

	ids, err := bus.IDs()
	r.Nil(ids)
	r.NotNil(err)
	r.ErrorIs(err, internal)
	r.ErrorContains(err, w1Path)
	r.ErrorContains(err, "ReadFile")
}

func (t *BusSuite) TestBus_DiscoverSingleError() {
	onewire := new(OnewireMock)
	w1Path := "all good"

	ids := []string{"1", "2", "3"}
	errs := []error{nil, io.ErrNoProgress, nil}
	res := []byte("9")
	for i, id := range ids {
		onewire.On("ReadFile", path.Join(w1Path, id, "resolution")).Return(res, errs[i])
	}

	onewire.On("Path").Return(w1Path)
	onewire.On("ReadFile", path.Join(w1Path, "w1_master_slaves")).Return([]byte("1\n2\n3"), nil)
	r := t.Require()
	bus, err := ds18b20.NewBus(ds18b20.WithInterface(onewire))
	r.Nil(err)
	r.NotNil(bus)

	s, errs := bus.Discover()
	r.NotNil(errs)
	r.NotNil(s)

	r.Len(s, 2)
	r.Len(errs, 1)
	r.ErrorIs(errs[0], io.ErrNoProgress)

}

func (t *BusSuite) TestBus_DiscoverFine() {
	onewire := new(OnewireMock)
	w1Path := "all good"

	s_ids := []string{"1", "2", "3"}
	ids := []byte("1\n2\n3")
	res := []byte("9")
	for _, id := range s_ids {
		onewire.On("ReadFile", path.Join(w1Path, id, "resolution")).Return(res, nil)
	}

	onewire.On("Path").Return(w1Path)
	onewire.On("ReadFile", path.Join(w1Path, "w1_master_slaves")).Return(ids, nil)

	r := t.Require()
	bus, err := ds18b20.NewBus(ds18b20.WithInterface(onewire))
	r.Nil(err)
	r.NotNil(bus)

	s, errs := bus.Discover()
	r.Nil(errs)
	r.NotNil(s)
	r.Len(s, len(s_ids))

}

func (t *BusSuite) TestBus_DiscoverError() {
	onewire := new(OnewireMock)

	r := t.Require()
	bus, err := ds18b20.NewBus(ds18b20.WithInterface(onewire))
	r.Nil(err)
	r.NotNil(bus)

	w1Path := "error is coming"
	internal := io.ErrShortBuffer
	onewire.On("Path").Return(w1Path)
	onewire.On("ReadFile", mock.Anything).Return([]byte{}, internal).Once()

	sensors, errs := bus.Discover()
	r.Nil(sensors)
	r.NotNil(errs)
	r.Len(errs, 1)
	r.ErrorIs(errs[0], internal)
	r.ErrorContains(errs[0], w1Path)
	r.ErrorContains(errs[0], "ReadFile")
	r.ErrorContains(errs[0], "Discover")
}

func (t *BusSuite) TestBus_NoInterface() {
	r := t.Require()
	bus, err := ds18b20.NewBus()
	r.Nil(bus)
	r.NotNil(err)
	r.ErrorIs(err, ds18b20.ErrNoInterface)
}

func (d *OnewireMock) Path() string {
	args := d.Called()
	return args.String(0)
}

func (d *OnewireMock) WriteFile(name string, data []byte) error {
	return d.Called(name, data).Error(0)
}

func (d *OnewireMock) ReadFile(name string) ([]byte, error) {
	args := d.Called(name)
	return args.Get(0).([]byte), args.Error(1)
}
func (d *FileMock) WriteFile(name string, data []byte) error {
	return d.Called(name, data).Error(0)
}

func (d *FileMock) ReadFile(name string) ([]byte, error) {
	args := d.Called(name)
	return args.Get(0).([]byte), args.Error(1)
}
