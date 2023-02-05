/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20_test

import (
	"github.com/a-clap/iot/internal/embedded/avg"
	"github.com/a-clap/iot/internal/embedded/ds18b20"
	"github.com/a-clap/iot/internal/embedded/models"
	"github.com/stretchr/testify/mock"
	"strconv"
	"time"
)

type SensorHandlerMock struct {
	mock.Mock
}

type PollDataMock struct {
	mock.Mock
}

func (t *DSTestSuite) TestPoll() {
	t.mock = new(SensorHandlerMock)
	t.pollData = new(PollDataMock)
	id := "123"

	t.mock.On("ID").Return(id)
	t.mock.On("Resolution").Return(models.Resolution10BIT, nil)
	t.mock.On("Poll", mock.Anything, mock.Anything).Return(nil)
	t.mock.On("Close").Return(nil)

	t.pollData.On("ID").Return(id)

	ds := ds18b20.NewDSSensor(t.mock)
	err := ds.Poll()
	t.Nil(err)

	// Add this point there should be any data
	stat := ds.Temperature()
	t.Equal(id, stat.ID)
	t.Equal(true, stat.Enabled)

	polledChannel := t.mock.TestData()["readings"].(chan ds18b20.Readings)

	temperatures := []string{"100.0", "200.0", "300.0", "400.0", "500.0"}
	t.EqualValues(len(temperatures), models.DefaultSamples)

	for i, elem := range temperatures {
		now := time.Now()
		t.pollData.On("Temperature").Return(elem).Once()
		t.pollData.On("Stamp").Return(now).Once()

		polledChannel <- t.pollData
		// Just to force scheduler work
		<-time.After(1 * time.Millisecond)
		stat = ds.Temperature()
		t.Equal(id, stat.ID)
		t.Equal(true, stat.Enabled)
		t.Equal(now, stat.Stamp)

		expectedTmp := float32(0)
		for j := 0; j <= i; j++ {
			f, err := strconv.ParseFloat(temperatures[j], 32)
			t.Nil(err)
			expectedTmp += float32(f)
		}
		expectedTmp /= float32(i + 1)

		t.InDelta(expectedTmp, stat.Temperature, 0.1)

	}

	err = ds.StopPoll()
	t.Nil(err)

}

func (t *DSTestSuite) TestSetGetConfig() {
	args := []struct {
		newConfig models.DSConfig
		err       error
	}{
		{
			newConfig: models.DSConfig{
				ID:             "1",
				Enabled:        false,
				Resolution:     models.Resolution9BIT,
				PollTimeMillis: 120,
				Samples:        13,
			},
			err: nil,
		},
		{
			newConfig: models.DSConfig{
				ID:             "2",
				Enabled:        false,
				Resolution:     models.Resolution11BIT,
				PollTimeMillis: 0,
				Samples:        1,
			},
			err: nil,
		},
		{
			newConfig: models.DSConfig{
				ID:             "sasaaax",
				Enabled:        false,
				Resolution:     models.Resolution12BIT,
				PollTimeMillis: 163,
				Samples:        25,
			},
			err: nil,
		},
		{
			newConfig: models.DSConfig{
				ID:             "heeeey",
				Enabled:        false,
				Resolution:     models.Resolution12BIT,
				PollTimeMillis: 163,
				Samples:        0,
			},
			err: avg.ErrSizeIsZero,
		},
	}
	for _, arg := range args {
		t.mock = new(SensorHandlerMock)
		t.mock.On("ID").Return(arg.newConfig.ID)
		t.mock.On("Resolution").Return(arg.newConfig.Resolution, nil)

		ds := ds18b20.NewDSSensor(t.mock)
		t.NotNil(ds)

		err := ds.SetConfig(arg.newConfig)
		t.Equal(arg.err, err)
		if arg.err != nil {
			continue
		}

		c := ds.Config()
		t.Equal(arg.newConfig, c)

	}
}

func (t *DSTestSuite) TestNew_VerifyConfig() {
	args := []struct {
		name string
		id   string
		res  models.DSResolution
	}{
		{
			name: "10 bit reso",
			id:   "123",
			res:  models.Resolution10BIT,
		},
		{
			name: "12 bit reso",
			id:   "hello world 2",
			res:  models.Resolution12BIT,
		},
		{
			name: "11 bit reso",
			id:   "hello world 3",
			res:  models.Resolution11BIT,
		},
		{
			name: "9 bit reso",
			id:   "hello world 4",
			res:  models.Resolution9BIT,
		},
	}
	for _, arg := range args {
		t.mock = new(SensorHandlerMock)
		t.mock.On("ID").Return(arg.id)
		t.mock.On("Resolution").Return(arg.res, nil)

		ds := ds18b20.NewDSSensor(t.mock)
		t.NotNil(ds, arg.name)
		cfg := ds.Config()

		t.EqualValues(arg.id, cfg.ID, arg.name)
		t.EqualValues(arg.res, cfg.Resolution, arg.name)
		t.EqualValues(false, cfg.Enabled, arg.name)

	}
}

func (s *SensorHandlerMock) ID() string {
	return s.Called().String(0)
}

func (s *SensorHandlerMock) Resolution() (models.DSResolution, error) {
	args := s.Called()
	return args.Get(0).(models.DSResolution), args.Error(1)
}

func (s *SensorHandlerMock) SetResolution(resolution models.DSResolution) error {
	return s.Called(resolution).Error(0)
}

func (s *SensorHandlerMock) PollTime() uint {
	return s.Called().Get(0).(uint)
}

func (s *SensorHandlerMock) Poll(data chan ds18b20.Readings, pollTime time.Duration) error {
	s.TestData()["readings"] = data
	return s.Called(data, pollTime).Error(0)
}

func (s *SensorHandlerMock) Close() error {
	return s.Called().Error(0)
}

func (p *PollDataMock) ID() string {
	return p.Called().String(0)
}

func (p *PollDataMock) Temperature() string {
	return p.Called().String(0)
}

func (p *PollDataMock) Stamp() time.Time {
	return p.Called().Get(0).(time.Time)
}

func (p *PollDataMock) Error() error {
	return p.Called().Error(0)
}
