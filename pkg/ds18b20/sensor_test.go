/*
* Copyright (c) 2023 a-clap. All rights reserved.
* Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20_test

import (
	"errors"
	"io"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SensorSuite struct {
	suite.Suite
}

func TestSensorSuite(t *testing.T) {
	suite.Run(t, new(SensorSuite))
}

func (t *SensorSuite) TestSensor_OpenFileError() {
	filer := new(FileMock)
	filer.On("ReadFile", mock.Anything).Return([]byte{}, io.ErrNoProgress)

	r := t.Require()
	s, err := ds18b20.NewSensor(filer, "1", "base")
	r.Nil(s)
	r.NotNil(err)
	r.ErrorIs(err, io.ErrNoProgress)
	r.ErrorContains(err, "1")
	r.ErrorContains(err, "resolution")
	r.ErrorContains(err, "base")
}

func (t *SensorSuite) TestSensor_ResolutionParseIntError() {
	resolution := []byte("magic")
	filer := new(FileMock)
	filer.On("ReadFile", mock.Anything).Return(resolution, nil)
	r := t.Require()
	s, err := ds18b20.NewSensor(filer, "1", "base")
	r.Nil(s)
	r.NotNil(err)
	r.ErrorContains(err, "1")
	r.ErrorContains(err, "resolution")
	r.ErrorContains(err, "base")
	r.ErrorContains(err, "magic")
}

func (t *SensorSuite) TestSensor_SetResolutionError() {
	resolution := []byte("9")
	filer := new(FileMock)
	filer.On("ReadFile", mock.Anything).Return(resolution, nil).Once()
	r := t.Require()
	s, err := ds18b20.NewSensor(filer, "1", "base")
	r.NotNil(s)
	r.Nil(err)

	cfg := ds18b20.SensorConfig{
		Name:         "blah",
		ID:           "1",
		Correction:   1,
		Resolution:   ds18b20.Resolution12Bit,
		PollInterval: 0,
		Samples:      0,
	}

	filer.On("WriteFile", mock.Anything, mock.Anything).Return(io.ErrShortBuffer).Once()
	e := s.Configure(cfg)
	r.ErrorIs(e, io.ErrShortBuffer)
	r.ErrorContains(e, "base")
	r.ErrorContains(e, "1")
	r.ErrorContains(e, "base/1/resolution")
}

func (t *SensorSuite) TestSensor_NewSensor_ReadResolution() {
	args := []struct {
		name        string
		id, path    string
		expectedErr error
		resolution  []byte
		expectedRes ds18b20.Resolution
	}{
		{
			name:        "all good - 9 bit",
			id:          "hello",
			path:        "path",
			expectedErr: nil,
			resolution:  []byte("9"),
			expectedRes: ds18b20.Resolution9Bit,
		},
		{
			name:        "all good  - 10 bit",
			id:          "hello 123",
			path:        "path 1",
			expectedErr: nil,
			resolution:  []byte("10"),
			expectedRes: ds18b20.Resolution10Bit,
		},
		{
			name:        "all good - 11 bit",
			id:          "hello 123",
			path:        "path 1",
			expectedErr: nil,
			resolution:  []byte("11"),
			expectedRes: ds18b20.Resolution11Bit,
		},
		{
			name:        "all good - 12 bit",
			id:          "hello 123",
			path:        "path 1",
			expectedErr: nil,
			resolution:  []byte("12"),
			expectedRes: ds18b20.Resolution12Bit,
		},
		{
			name:        "all good - with \r\n",
			id:          "hello 123",
			path:        "path 1",
			expectedErr: nil,
			resolution:  []byte("12\r\n"),
			expectedRes: ds18b20.Resolution12Bit,
		},
	}
	for _, arg := range args {
		r := t.Require()
		resolutionPath := path.Join(arg.path, arg.id, "resolution")

		fileMock := new(FileMock)
		fileMock.On("ReadFile", resolutionPath).Return(arg.resolution, nil)

		s, err := ds18b20.NewSensor(fileMock, arg.id, arg.path)
		if arg.expectedErr != nil {
			r.NotNil(err, arg.name)
			r.ErrorContains(err, arg.expectedErr.Error(), arg.name)
			continue
		}
		res := s.GetConfig().Resolution
		r.Equal(arg.expectedRes, res, arg.name)
	}
}

func (t *SensorSuite) TestSensor_TemperatureConversions() {
	args := []struct {
		name           string
		expectedErr    error
		expected       float32
		temperatureBuf []byte
	}{
		{
			name:           "0.001",
			expected:       0.001,
			temperatureBuf: []byte("1"),
		},
		{
			name:           "988.654",
			expected:       988.654,
			temperatureBuf: []byte("988654\r\n"),
		},
		{
			name:           "12.355",
			expected:       12.355,
			temperatureBuf: []byte("12355\r\n"),
		},
		{
			name:           "1.230",
			expected:       1.230,
			temperatureBuf: []byte("1230\r"),
		},
		{
			name:           "0.456",
			expected:       0.456,
			temperatureBuf: []byte("456\n"),
		},
		{
			name:           "0.038",
			expected:       0.038,
			temperatureBuf: []byte("38\n"),
		},
		{
			name:           "not a float",
			expected:       0.038,
			expectedErr:    &strconv.NumError{},
			temperatureBuf: []byte("fff\n"),
		},
	}
	resolution := []byte("11")
	for _, arg := range args {
		r := t.Require()
		file := new(FileMock)
		file.On("ReadFile", path.Join("", "", "resolution")).Return(resolution, nil)
		file.On("ReadFile", path.Join("", "", "temperature")).Return(arg.temperatureBuf, nil)

		s, err := ds18b20.NewSensor(file, "", "")
		r.Nil(err, arg.name)
		v, _, err := s.Temperature()

		if arg.expectedErr != nil {
			r.NotNil(err, arg.name)
			continue
		}
		r.Nil(err, arg.name)
		r.InDelta(arg.expected, v, 0.0001)
	}
}

func (t *SensorSuite) TestSensor_InitConfig() {
	args := []struct {
		name           string
		new            ds18b20.SensorConfig
		expectedResBuf []byte
		resWriteErr    error
		expectedErr    error
	}{
		{
			name: "all good - 9 bit",
			new: ds18b20.SensorConfig{
				Name:         "blah",
				Correction:   13,
				Resolution:   ds18b20.Resolution9Bit,
				PollInterval: 123,
				Samples:      15,
			},
			expectedResBuf: []byte(strconv.FormatInt(int64(ds18b20.Resolution9Bit), 10) + "\r\n"),
			expectedErr:    nil,
		},
		{
			name: "all good - 10 bit",
			new: ds18b20.SensorConfig{
				Name:         "",
				Correction:   13,
				Resolution:   ds18b20.Resolution10Bit,
				PollInterval: 123,
				Samples:      15,
			},
			expectedResBuf: []byte(strconv.FormatInt(int64(ds18b20.Resolution10Bit), 10) + "\r\n"),
			expectedErr:    nil,
		},
		{
			name: "all good - 11 bit",
			new: ds18b20.SensorConfig{
				Name:         "helooooooooo",
				Correction:   13,
				Resolution:   ds18b20.Resolution11Bit,
				PollInterval: 123,
				Samples:      15,
			},
			expectedResBuf: []byte(strconv.FormatInt(int64(ds18b20.Resolution11Bit), 10) + "\r\n"),
			expectedErr:    nil,
		},
		{
			name: "all good - 12 bit",
			new: ds18b20.SensorConfig{
				Name:         "another",
				Correction:   13,
				Resolution:   ds18b20.Resolution12Bit,
				PollInterval: 123,
				Samples:      15,
			},
			expectedResBuf: []byte(strconv.FormatInt(int64(ds18b20.Resolution12Bit), 10) + "\r\n"),
			expectedErr:    nil,
		},
		{
			name: "err on resolution write",
			new: ds18b20.SensorConfig{
				Name:         "1",
				Correction:   13,
				Resolution:   ds18b20.Resolution12Bit,
				PollInterval: 123,
				Samples:      15,
			},
			expectedResBuf: []byte(strconv.FormatInt(int64(ds18b20.Resolution12Bit), 10) + "\r\n"),
			resWriteErr:    errors.New("write error"),
			expectedErr:    errors.New("write error"),
		},
	}
	resolution := []byte("11")
	for _, arg := range args {
		r := t.Require()

		filer := new(FileMock)
		filer.On("ReadFile", path.Join("", "", "resolution")).Return(resolution, nil)
		filer.On("WriteFile", path.Join("", "", "resolution"), arg.expectedResBuf).Return(arg.resWriteErr)

		s, err := ds18b20.NewSensor(filer, "", "")
		r.Nil(err, arg.name)

		err = s.Configure(arg.new)
		if arg.expectedErr != nil {
			r.NotNil(err, arg.name)
			r.ErrorContains(err, arg.expectedErr.Error(), arg.name)
			continue
		}
		r.Nil(err)
		r.Equal(arg.new, s.GetConfig())
		r.Equal(arg.new.Name, s.Name())

	}
}

func (t *SensorSuite) TestSensor_PollTwice() {

	resolution := []byte("11")
	temperature := []byte("100")
	filer := new(FileMock)
	filer.On("ReadFile", mock.Anything).Return(resolution, nil)
	filer.On("ReadFile", mock.Anything).Return(temperature, nil)

	sensor, _ := ds18b20.NewSensor(filer, "", "")
	cfg := ds18b20.SensorConfig{
		Correction:   0,
		Resolution:   ds18b20.Resolution11Bit,
		PollInterval: 15 * time.Millisecond,
		Samples:      3,
	}

	r := t.Require()
	r.Nil(sensor.Configure(cfg))

	// We don't want to test Poll, just error handling
	sensor.Poll()
	sensor.Poll()

	wait := make(chan struct{})

	go func() {
		sensor.Close()
		wait <- struct{}{}
	}()

	select {
	case <-wait:
	case <-time.After(100 * time.Millisecond):
		t.Fail("should be done after this time")
	}
	close(wait)
}

func (t *SensorSuite) TestSensor_Poll() {
	resolution := []byte("11")
	filer := new(FileMock)
	filer.On("ReadFile", mock.Anything).Return(resolution, nil)
	id := "blah"

	sensor, _ := ds18b20.NewSensor(filer, id, "")
	cfg := ds18b20.SensorConfig{
		Correction:   0,
		Resolution:   ds18b20.Resolution11Bit,
		PollInterval: 10 * time.Millisecond,
		Samples:      3,
	}

	r := t.Require()
	r.Nil(sensor.Configure(cfg))

	temperatures := []struct {
		expTmp float32
		expErr error
		tmp    string
		err    error
	}{
		{
			expTmp: 0.123,
			tmp:    "123",
			err:    io.EOF,
			expErr: nil,
		},
		{
			expTmp: 0,
			tmp:    "-123",
			err:    errors.New("error please"),
			expErr: errors.New("error please"),
		},
		{
			expTmp: 0.856,
			tmp:    "856",
			err:    io.EOF,
			expErr: nil,
		},
		{
			expTmp: 0,
			tmp:    "123343",
			err:    errors.New("another one"),
			expErr: errors.New("another one"),
		},
		{
			expTmp: 67.012,
			tmp:    "67012",
			err:    io.EOF,
			expErr: nil,
		},
	}

	for _, temp := range temperatures {
		filer.On("ReadFile", path.Join("", id, "temperature")).Return([]byte(temp.tmp), temp.err).Once()
	}

	sensor.Poll()
	data := make([]ds18b20.Readings, 0, len(temperatures))
	<-time.After(cfg.PollInterval * time.Duration(len(temperatures)+1))
	data = append(data, sensor.GetReadings()...)

	sensor.Close()
	r.Equal(len(data), len(temperatures))

	for i := range data {
		if i == 0 {
			continue
		}
		diff := data[i].Stamp.Sub(data[i-1].Stamp)
		r.InDelta(cfg.PollInterval, diff, float64(3*time.Millisecond))
	}

}
