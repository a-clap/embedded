/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20_test

import (
	"errors"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/a-clap/iot/pkg/ds18b20"
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
		name     string
		path     string
		err      error
		dirEntry []string
		ids      []string
	}{
		{
			name:     "handle interface error",
			path:     "expectedPath",
			err:      errors.New("interface error"),
			dirEntry: nil,
			ids:      nil,
		},
		{
			name:     "no devices on bus",
			path:     "new temperaturePath",
			err:      nil,
			dirEntry: nil,
			ids:      nil,
		},
		{
			name:     "single device on bus",
			path:     "another temperaturePath",
			err:      nil,
			dirEntry: []string{"28-05169397aeff"},
			ids:      []string{"28-05169397aeff"},
		},
		{
			name: "ignore other files",
			path: "another temperaturePath",
			err:  nil,
			dirEntry: []string{"28-051693848dff", "w1_master_name", "28-05169397aeff", "w1_master_pointer", "driver", "w1_master_pullup", "power",
				"w1_master_remove", "subsystem", "w1_master_search",
				"therm_bulk_read", "w1_master_slave_count", "uevent", "w1_master_slaves", "w1_master_add", "w1_master_timeout", "w1_master_attempts",
				"w1_master_timeout_us", "w1_master_max_slave_count"},
			ids: []string{"28-051693848dff", "28-05169397aeff"},
		},
	}
	r := t.Require()
	for _, arg := range args {
		onewire := new(OnewireMock)
		onewire.On("Path").Return(arg.path)
		onewire.On("ReadDir", arg.path).Return(arg.dirEntry, arg.err)

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
	w1path := "/sys/bus/w1/devices/w1_bus_master1"
	dirEntry := []string{"28-05169397aeff"}

	files := make([]*FileMock, 1)
	files[0] = new(FileMock)
	onewire := new(OnewireMock)
	onewire.On("Path").Return(w1path)
	onewire.On("ReadDir", w1path).Return(dirEntry, nil)
	onewire.On("Open", path.Join(w1path, id, "resolution")).Return(files[0], nil).Once()
	files[0].On("Close").Return(nil)

	resolutionBuf := []byte("11")
	call := files[0].On("Read", mock.Anything).Return(len(resolutionBuf), nil).Once().Run(func(args mock.Arguments) {
		buf := args.Get(0).([]byte)
		copy(buf, resolutionBuf)
	})

	files[0].On("Read", mock.Anything).Return(0, io.EOF).NotBefore(call)

	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(onewire),
	)

	sensor, err := bus.NewSensor(id)

	t.NotNil(sensor)
	t.Nil(err)
	fullID := sensor.ID()
	s := strings.Split(fullID, ":")
	t.Len(s, 2)
	t.EqualValues("w1_bus_master1", s[0])
	t.EqualValues(id, s[1])

	onewire.AssertExpectations(t.T())
}

func (t *BusSuite) TestBus_NewSensorWrongID() {
	id := "hello world"
	w1path := "some w1path"
	dirEntry := []string{"28-05169397aeff"}
	expectedErr := ds18b20.ErrNoSuchID

	onewire := new(OnewireMock)
	onewire.On("Path").Return(w1path)
	onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()

	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(onewire),
	)

	sensor, err := bus.NewSensor(id)
	t.Nil(sensor)
	t.NotNil(err)
	t.ErrorContains(err, expectedErr.Error())

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
	onewire.On("ReadDir", w1Path).Return([]string{}, internal).Once()

	ids, err := bus.NewSensor("hello")
	r.Nil(ids)
	r.NotNil(err)
	r.ErrorIs(err, internal)
	r.ErrorContains(err, w1Path)
	r.ErrorContains(err, "ReadDir")
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
	onewire.On("ReadDir", w1Path).Return([]string{}, internal).Once()

	ids, err := bus.IDs()
	r.Nil(ids)
	r.NotNil(err)
	r.ErrorIs(err, internal)
	r.ErrorContains(err, w1Path)
	r.ErrorContains(err, "ReadDir")
}

func (t *BusSuite) TestBus_DiscoverSingleError() {
	onewire := new(OnewireMock)
	w1Path := "all good"

	ids := []string{"1", "2", "3"}
	errs := []error{io.EOF, io.ErrNoProgress, io.EOF}
	res := []byte("9")
	for i, id := range ids {
		f := new(FileMock)
		f.On("Read", mock.Anything).Return(len(res), errs[i]).Run(
			func(args mock.Arguments) {
				buf := args.Get(0).([]byte)
				copy(buf, res)
			})
		f.On("Read")
		f.On("Close").Return(nil)
		onewire.On("Open", path.Join(w1Path, id, "resolution")).Return(f, nil)
	}

	onewire.On("Path").Return(w1Path)
	onewire.On("ReadDir", w1Path).Return(ids, nil)

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

	ids := []string{"1", "2", "3"}
	res := []byte("9")
	for _, id := range ids {
		f := new(FileMock)
		f.On("Read", mock.Anything).Return(len(res), io.EOF).Run(
			func(args mock.Arguments) {
				buf := args.Get(0).([]byte)
				copy(buf, res)
			})
		f.On("Read")
		f.On("Close").Return(nil)
		onewire.On("Open", path.Join(w1Path, id, "resolution")).Return(f, nil)
	}

	onewire.On("Path").Return(w1Path)
	onewire.On("ReadDir", w1Path).Return(ids, nil)

	r := t.Require()
	bus, err := ds18b20.NewBus(ds18b20.WithInterface(onewire))
	r.Nil(err)
	r.NotNil(bus)

	s, errs := bus.Discover()
	r.Nil(errs)
	r.NotNil(s)
	r.Len(s, len(ids))

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
	onewire.On("ReadDir", w1Path).Return([]string{}, internal).Once()

	sensors, errs := bus.Discover()
	r.Nil(sensors)
	r.NotNil(errs)
	r.Len(errs, 1)
	r.ErrorIs(errs[0], internal)
	r.ErrorContains(errs[0], w1Path)
	r.ErrorContains(errs[0], "ReadDir")
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

func (d *OnewireMock) ReadDir(dirname string) ([]string, error) {
	args := d.Called(dirname)
	return args.Get(0).([]string), args.Error(1)

}

func (d *OnewireMock) Open(name string) (ds18b20.File, error) {
	args := d.Called(name)
	return args.Get(0).(ds18b20.File), args.Error(1)
}

func (d *FileMock) Read(p []byte) (n int, err error) {
	args := d.Called(p)
	return args.Int(0), args.Error(1)
}

func (d *FileMock) Write(p []byte) (n int, err error) {
	args := d.Called(p)
	return args.Int(0), args.Error(1)
}

func (d *FileMock) Close() error {
	return d.Called().Error(0)
}