/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20_test

import (
	"errors"
	"github.com/a-clap/iot/internal/embedded/ds18b20"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

type DSOnewireMock struct {
	mock.Mock
}
type DSFileMock struct {
	mock.Mock
}

type DSTestSuite struct {
	suite.Suite
	onewire  *DSOnewireMock
	file     []*DSFileMock
	mock     *SensorHandlerMock
	pollData *PollDataMock
}

func TestDS8B20Run(t *testing.T) {
	suite.Run(t, new(DSTestSuite))
}

func (t *DSTestSuite) SetupTest() {
	t.onewire = new(DSOnewireMock)
	t.file = make([]*DSFileMock, 0)
}

func (t *DSTestSuite) TearDownTest() {
	t.onewire.AssertExpectations(t.T())
	for _, f := range t.file {
		f.AssertExpectations(t.T())
	}
}
func (t *DSTestSuite) TearDownAllSuite() {
	t.onewire = nil
	for i := range t.file {
		t.file[i] = nil
	}
}

func (t *DSTestSuite) TestIDs() {
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
			dirEntry: []string{"28-051693848dff", "w1_master_name", "28-05169397aeff", "w1_master_pointer", "driver", "w1_master_pullup", "power", "w1_master_remove", "subsystem", "w1_master_search",
				"therm_bulk_read", "w1_master_slave_count", "uevent", "w1_master_slaves", "w1_master_add", "w1_master_timeout", "w1_master_attempts", "w1_master_timeout_us", "w1_master_max_slave_count"},
			ids: []string{"28-051693848dff", "28-05169397aeff"},
		},
	}
	for _, arg := range args {
		t.onewire.On("Path").Return(arg.path).Once()
		t.onewire.On("ReadDir", arg.path).Return(arg.dirEntry, arg.err).Once()

		h, err := ds18b20.NewBus(
			ds18b20.WithInterface(t.onewire),
		)

		t.NotNil(h, arg.name)
		t.Nil(err, arg.name)

		ids, err := h.IDs()
		t.EqualValues(arg.ids, ids, arg.name)

		if arg.err != nil {
			t.ErrorContains(err, arg.err.Error(), arg.name)
		} else {
			t.Nil(err, arg.name)
		}
	}
}
func (t *DSTestSuite) TestResolution_Set() {
	id := "28-05169397aeff"
	w1path := "TestResolution_Set"
	dirEntry := []string{"28-05169397aeff"}

	t.file = make([]*DSFileMock, 2)
	for i := range t.file {
		t.file[i] = new(DSFileMock)
	}

	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()

	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()

	data := t.file[0].TestData()
	data["writeBuf"] = []byte("567")
	call := t.file[0].On("Read", mock.Anything).Return(3, nil).Once()
	readCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
	t.file[0].On("Close").Return(nil).Once().NotBefore(readCall)
	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)
	sensor, _ := bus.NewSensor(id)

	resFile := t.file[1]
	args := []struct {
		writeBuf []byte
		res      models.DSResolution
	}{
		{
			writeBuf: []byte(strconv.FormatInt(int64(models.Resolution9BIT), 10)),
			res:      models.Resolution9BIT,
		},
		{
			writeBuf: []byte(strconv.FormatInt(int64(models.Resolution10BIT), 10)),
			res:      models.Resolution10BIT,
		}, {
			writeBuf: []byte(strconv.FormatInt(int64(models.Resolution11BIT), 10)),
			res:      models.Resolution11BIT,
		},
		{
			writeBuf: []byte(strconv.FormatInt(int64(models.Resolution12BIT), 10)),
			res:      models.Resolution12BIT,
		},
	}
	for _, arg := range args {
		openCall := t.onewire.On("Open", path.Join(w1path, id, "resolution")).Return(resFile, nil).Once()
		writeCall := resFile.On("Write", arg.writeBuf).Return(len(arg.writeBuf), nil).Once().NotBefore(openCall)
		resFile.On("Close").Return(nil).Once().NotBefore(writeCall)
		err := sensor.SetResolution(arg.res)
		t.Nil(err)
	}
}
func (t *DSTestSuite) TestResolution_Read() {
	id := "28-05169397aeff"
	w1path := "TestResolution_Read"
	dirEntry := []string{"28-05169397aeff"}

	t.file = make([]*DSFileMock, 2)
	for i := range t.file {
		t.file[i] = new(DSFileMock)
	}

	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()

	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()

	data := t.file[0].TestData()
	data["writeBuf"] = []byte("567")
	call := t.file[0].On("Read", mock.Anything).Return(3, nil).Once()
	readCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
	t.file[0].On("Close").Return(nil).Once().NotBefore(readCall)
	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)
	sensor, _ := bus.NewSensor(id)

	resFile := t.file[1]
	args := []struct {
		buf      []byte
		expected models.DSResolution
	}{
		{
			buf:      []byte(strconv.FormatInt(int64(models.Resolution9BIT), 10)),
			expected: models.Resolution9BIT,
		},
		{
			buf:      []byte(strconv.FormatInt(int64(models.Resolution10BIT), 10)),
			expected: models.Resolution10BIT,
		}, {
			buf:      []byte(strconv.FormatInt(int64(models.Resolution11BIT), 10)),
			expected: models.Resolution11BIT,
		},
		{
			buf:      []byte(strconv.FormatInt(int64(models.Resolution12BIT), 10)),
			expected: models.Resolution12BIT,
		},
	}
	for _, arg := range args {
		openCall := t.onewire.On("Open", path.Join(w1path, id, "resolution")).Return(resFile, nil).Once()
		dataResolution := resFile.TestData()
		dataResolution["writeBuf"] = arg.buf
		resReadCall := resFile.On("Read", mock.Anything).Return(len(arg.buf), nil).Once().NotBefore(openCall)
		resReadEOFcall := resFile.On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(resReadCall)
		resFile.On("Close").Return(nil).Once().NotBefore(resReadEOFcall)
		res, err := sensor.Resolution()
		t.Nil(err)
		t.EqualValues(arg.expected, res)
	}

}
func (t *DSTestSuite) TestSensor_Poll() {
	id := "28-05169397aeff"
	w1path := "TestSensor_Poll"
	dirEntry := []string{"28-05169397aeff"}

	t.file = make([]*DSFileMock, 1)
	t.file[0] = new(DSFileMock)
	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()
	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()

	data := t.file[0].TestData()
	data["writeBuf"] = []byte("123")
	call := t.file[0].On("Read", mock.Anything).Return(3, nil).Once()
	readCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
	t.file[0].On("Close").Return(nil).Once().NotBefore(readCall)

	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)
	sensor, _ := bus.NewSensor(id)

	interval := 10 * time.Millisecond
	readings := make(chan ds18b20.Readings)

	args := []struct {
		buf      []byte
		expected string
	}{
		{
			buf:      []byte("1"),
			expected: "0.001",
		},
		{
			buf:      []byte("988654\r\n"),
			expected: "988.654",
		},
		{
			buf:      []byte("12355\r\n"),
			expected: "12.355",
		},
		{
			buf:      []byte("1230\r"),
			expected: "1.230",
		},
		{
			buf:      []byte("456\n"),
			expected: "0.456",
		},
		{
			buf:      []byte("38\n"),
			expected: "0.038",
		},
	}

	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Times(len(args))
	_ = sensor.Poll(readings, interval)

	for _, arg := range args {
		t.file[0].TestData()["writeBuf"] = arg.buf
		call := t.file[0].On("Read", mock.Anything).Return(len(arg.buf), nil).Once()
		readCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
		t.file[0].On("Close").Return(nil).Once().NotBefore(readCall)
		now := time.Now()
		select {
		case r := <-readings:
			t.EqualValues(id, r.ID())
			t.EqualValues(arg.expected, r.Temperature())
			t.Nil(r.Error())
			diff := r.Stamp().Sub(now)
			t.Less(interval, diff)
		case <-time.After(2 * interval):
			t.Fail("failed, waiting for readings too long")
		}
	}
	wait := make(chan struct{})

	go func() {
		t.Nil(sensor.Close())
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(2 * interval):
		t.Fail("should be done after this time")
	}

}

func (t *DSTestSuite) TestSensor_PollTwice() {
	id := "28-05169397aeff"
	w1path := "TestSensor_PollTwice"
	dirEntry := []string{"28-05169397aeff"}

	t.file = make([]*DSFileMock, 1)
	t.file[0] = new(DSFileMock)
	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()
	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()

	data := t.file[0].TestData()
	data["writeBuf"] = []byte("123")
	call := t.file[0].On("Read", mock.Anything).Return(3, nil).Once()
	readCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
	t.file[0].On("Close").Return(nil).Once().NotBefore(readCall)

	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)
	sensor, _ := bus.NewSensor(id)

	t.file[0].On("Read", mock.Anything).Return(0, io.EOF)

	// We don't want to test Poll, just error handling
	interval := 1 * time.Second
	readings := make(chan ds18b20.Readings)
	err := sensor.Poll(readings, interval)
	t.Nil(err)

	err = sensor.Poll(readings, interval)
	t.ErrorIs(err, ds18b20.ErrAlreadyPolling)

	wait := make(chan struct{})

	go func() {
		t.Nil(sensor.Close())
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(2 * interval):
		t.Fail("should be done after this time")
	}
	close(wait)
}

func (t *DSTestSuite) TestSensor_TemperatureConversions() {
	id := "28-05169397aeff"
	w1path := "TestSensor_TemperatureConversions"
	dirEntry := []string{"28-05169397aeff"}

	t.file = make([]*DSFileMock, 1)
	t.file[0] = new(DSFileMock)
	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()
	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()

	data := t.file[0].TestData()
	data["writeBuf"] = []byte("123")
	call := t.file[0].On("Read", mock.Anything).Return(3, nil).Once()
	readCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
	t.file[0].On("Close").Return(nil).Once().NotBefore(readCall)
	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)
	sensor, _ := bus.NewSensor(id)
	args := []struct {
		buf      []byte
		expected string
	}{
		{
			buf:      []byte("1"),
			expected: "0.001",
		},
		{
			buf:      []byte("988654\r\n"),
			expected: "988.654",
		},
		{
			buf:      []byte("12355\r\n"),
			expected: "12.355",
		},
		{
			buf:      []byte("1230\r"),
			expected: "1.230",
		},
		{
			buf:      []byte("456\n"),
			expected: "0.456",
		},
		{
			buf:      []byte("38\n"),
			expected: "0.038",
		},
	}
	for i, arg := range args {
		data := t.file[0].TestData()
		data["writeBuf"] = arg.buf

		openCall := t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()
		readCall := t.file[0].On("Read", mock.Anything).Return(len(arg.buf), nil).Once().NotBefore(openCall)
		readEOFCall := t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(readCall)
		t.file[0].On("Close").Return(nil).Once().NotBefore(readEOFCall)
		temp, err := sensor.Temperature()
		t.Nil(err)
		t.EqualValues(arg.expected, temp, i)

	}

}

func (t *DSTestSuite) TestNewSensor_Good() {
	id := "28-05169397aeff"
	w1path := "TestNewSensor_Good"
	dirEntry := []string{"28-05169397aeff"}

	t.file = make([]*DSFileMock, 1)
	t.file[0] = new(DSFileMock)
	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()
	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], nil).Once()

	data := t.file[0].TestData()
	data["writeBuf"] = []byte("123")

	call := t.file[0].On("Read", mock.Anything).Return(3, nil).Once()
	t.file[0].On("Read", mock.Anything).Return(0, io.EOF).Once().NotBefore(call)
	t.file[0].On("Close").Return(nil).Once()
	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)

	sensor, err := bus.NewSensor(id)
	t.NotNil(sensor)
	t.Nil(err)
	t.EqualValues(id, sensor.ID())
}

func (t *DSTestSuite) TestNewSensor_ErrorOnOpenTempeatureFile() {
	id := "28-05169397aeff"
	w1path := "TestNewSensor_ErrorOnOpenTempeatureFile"
	dirEntry := []string{"28-05169397aeff"}
	expectedErr := os.ErrNotExist

	t.file = make([]*DSFileMock, 1)
	t.file[0] = new(DSFileMock)
	t.onewire.On("Path").Return(w1path).Twice()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()
	t.onewire.On("Open", path.Join(w1path, id, "temperature")).Return(t.file[0], expectedErr).Once()
	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)

	sensor, err := bus.NewSensor(id)
	t.Nil(sensor)
	t.NotNil(err)
	t.ErrorIs(err, expectedErr)

}

func (t *DSTestSuite) TestNewSensor_WrongID() {
	id := "hello world"
	w1path := "some w1path"
	dirEntry := []string{"28-05169397aeff"}
	expectedErr := ds18b20.ErrNoSuchID

	t.onewire.On("Path").Return(w1path).Once()
	t.onewire.On("ReadDir", w1path).Return(dirEntry, nil).Once()

	bus, _ := ds18b20.NewBus(
		ds18b20.WithInterface(t.onewire),
	)

	sensor, err := bus.NewSensor(id)
	t.Nil(sensor)
	t.NotNil(err)
	t.ErrorIs(err, expectedErr)

}

func (d *DSOnewireMock) Path() string {
	args := d.Called()
	return args.String(0)
}

func (d *DSOnewireMock) ReadDir(dirname string) ([]string, error) {
	args := d.Called(dirname)
	return args.Get(0).([]string), args.Error(1)

}

func (d *DSOnewireMock) Open(name string) (ds18b20.File, error) {
	args := d.Called(name)
	return args.Get(0).(ds18b20.File), args.Error(1)
}

func (d *DSFileMock) Read(p []byte) (n int, err error) {
	args := d.Called(p)
	if maybeData, ok := d.TestData()["writeBuf"]; ok {
		copy(p, maybeData.([]byte))
	}
	return args.Int(0), args.Error(1)
}

func (d *DSFileMock) Write(p []byte) (n int, err error) {
	args := d.Called(p)
	return args.Int(0), args.Error(1)
}

func (d *DSFileMock) Close() error {
	return d.Called().Error(0)
}
