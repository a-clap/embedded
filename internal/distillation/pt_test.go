/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/a-clap/iot/internal/distillation"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/a-clap/iot/internal/embedded/max31865"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PTTestSuite struct {
	suite.Suite
}

type PTMock struct {
	mock.Mock
}

func (p *PTMock) Get() ([]embedded.PTSensorConfig, error) {
	args := p.Called()
	return args.Get(0).([]embedded.PTSensorConfig), args.Error(1)
}

func (p *PTMock) Set(s embedded.PTSensorConfig) error {
	return p.Called(s).Error(0)
}

func (p *PTMock) Temperatures() ([]embedded.PTTemperature, error) {
	args := p.Called()
	return args.Get(0).([]embedded.PTTemperature), args.Error(1)
}

func TestPT(t *testing.T) {
	suite.Run(t, new(PTTestSuite))
}

func (t *PTTestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

func (t *PTTestSuite) TestGetSensors_Rest() {
	args := []struct {
		name     string
		onGet    []embedded.PTSensorConfig
		expected []distillation.PTConfig
	}{
		{
			name: "single element",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}},
			expected: []distillation.PTConfig{{
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "two sensors on one bus",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "2",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.PTConfig{{
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "2",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "multiple sensors on multiple bus",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "2",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "4",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "5",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "12",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.PTConfig{{
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "2",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "4",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "5",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "12",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		h, err := distillation.New(distillation.WithPT(m))
		r.NotNil(h)
		r.Nil(err)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, distillation.RoutesGetPT, nil)

		h.ServeHTTP(recorder, request)
		r.Equal(http.StatusOK, recorder.Code, arg.name)
		var retCfg []distillation.PTConfig

		r.Nil(json.NewDecoder(recorder.Body).Decode(&retCfg), arg.name)
	}
}

func (t *PTTestSuite) TestTemperature_REST() {
	args := []struct {
		name                 string
		onGet                []embedded.PTSensorConfig
		onTemperatures       []embedded.PTTemperature
		expectedTemperatures []distillation.PTTemperature
	}{
		{
			name: "return average",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}},
			onTemperatures: []embedded.PTTemperature{{
				Readings: []max31865.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     123.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}},
			expectedTemperatures: []distillation.PTTemperature{{
				ID:          "1",
				Temperature: 123.0,
			}}}, {
			name: "return last average",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}},
			onTemperatures: []embedded.PTTemperature{{
				Readings: []max31865.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     123.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Readings: []max31865.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     128.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Readings: []max31865.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     -200.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}},
			expectedTemperatures: []distillation.PTTemperature{{
				ID:          "1",
				Temperature: 123.0,
			}}}}

	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.New(distillation.WithPT(m))
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.PTHandler.Update(), 0, arg.name)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, distillation.RoutesGetPTTemperatures, nil)

		h.ServeHTTP(recorder, request)
		r.Equal(http.StatusOK, recorder.Code, arg.name)
		var retCfg []distillation.PTTemperature

		r.Nil(json.NewDecoder(recorder.Body).Decode(&retCfg), arg.name)
	}
}

func (t *PTTestSuite) TestConfigureSensor_REST() {
	args := []struct {
		name        string
		newConfig   distillation.PTConfig
		onGet       []embedded.PTSensorConfig
		onSetErr    error
		errContains string
	}{{
		name: "all good",
		newConfig: distillation.PTConfig{PTSensorConfig: embedded.PTSensorConfig{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   1,
				ASyncPoll:    false,
				PollInterval: 3,
				Samples:      4,
			}}},
		onGet: []embedded.PTSensorConfig{{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   0,
				ASyncPoll:    false,
				PollInterval: 0,
				Samples:      0,
			}}},
		onSetErr:    nil,
		errContains: "",
	}, {
		name: "error on set interface",
		newConfig: distillation.PTConfig{PTSensorConfig: embedded.PTSensorConfig{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   1,
				ASyncPoll:    false,
				PollInterval: 3,
				Samples:      4,
			}}},
		onGet: []embedded.PTSensorConfig{{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   0,
				ASyncPoll:    false,
				PollInterval: 0,
				Samples:      0,
			}}},
		onSetErr:    errors.New("hello"),
		errContains: "hello",
	}, {
		name: "wrong ID",
		newConfig: distillation.PTConfig{PTSensorConfig: embedded.PTSensorConfig{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   1,
				ASyncPoll:    false,
				PollInterval: 3,
				Samples:      4,
			}}},
		onGet: []embedded.PTSensorConfig{{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "2",
				Correction:   0,
				ASyncPoll:    false,
				PollInterval: 0,
				Samples:      0,
			}}},
		onSetErr:    nil,
		errContains: distillation.ErrNoSuchID.Error(),
	}}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Set", arg.newConfig.PTSensorConfig).Return(arg.onSetErr)
		h, err := distillation.New(distillation.WithPT(m))
		r.Nil(err, arg.name)

		recorder := httptest.NewRecorder()
		var body bytes.Buffer
		r.Nil(json.NewEncoder(&body).Encode(arg.newConfig))

		request, _ := http.NewRequest(http.MethodPut, distillation.RoutesConfigurePT, &body)
		request.Header.Add("Content-Type", "application/json")

		h.ServeHTTP(recorder, request)
		if arg.errContains != "" {
			err := distillation.PTError{}
			r.Nil(json.NewDecoder(recorder.Body).Decode(&err), arg.name)
			r.Equal(http.StatusInternalServerError, recorder.Code, arg.name+":"+err.Error())
			r.ErrorContains(&err, arg.errContains, arg.name)
			continue
		}

		r.Equal(http.StatusOK, recorder.Code, recorder.Body.String())
		retCfg := distillation.PTConfig{}
		r.Nil(json.NewDecoder(recorder.Body).Decode(&retCfg), arg.name)

		r.Equal(arg.newConfig, retCfg, arg.name)
	}
}

func (t *PTTestSuite) TestAfterHistory_StillCanReadData() {
	args := []struct {
		name            string
		onGet           []embedded.PTSensorConfig
		onTemperatures  []embedded.PTTemperature
		expectedHistory []embedded.PTTemperature
		ids             []string
		temps           []float32
	}{{
		name: "single element in history",
		onGet: []embedded.PTSensorConfig{{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   0,
				ASyncPoll:    false,
				PollInterval: 0,
				Samples:      0,
			}}},
		onTemperatures: []embedded.PTTemperature{{
			Readings: []max31865.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     123.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}},
		expectedHistory: []embedded.PTTemperature{},
		ids:             []string{"1"},
		temps:           []float32{123.0},
	}, {
		name: "return all  but last element in history",
		onGet: []embedded.PTSensorConfig{
			{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}},
		onTemperatures: []embedded.PTTemperature{{
			Readings: []max31865.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -100.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}, {
			Readings: []max31865.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -50.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}, {
			Readings: []max31865.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -150.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}},
		expectedHistory: []embedded.PTTemperature{{
			Readings: []max31865.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -100.0,
				Stamp:       time.Time{},
				Error:       "",
			}, {
				ID:          "1",
				Temperature: 1,
				Average:     -50.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}},
		ids:   []string{"1"},
		temps: []float32{-150.0},
	},
		{
			name: "return all  but last element in history - more and mixed data",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "2",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "3",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}},
			onTemperatures: []embedded.PTTemperature{{
				Readings: []max31865.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     -100.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "3",
					Temperature: 1,
					Average:     -100.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "1",
					Temperature: 1,
					Average:     -125.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "3",
					Temperature: 1,
					Average:     -12300.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Readings: []max31865.Readings{{
					ID:          "2",
					Temperature: 1,
					Average:     -50.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "2",
					Temperature: 1,
					Average:     -510.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "2",
					Temperature: 1,
					Average:     -520.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Readings: []max31865.Readings{
					{
						ID:          "3",
						Temperature: 1,
						Average:     -150.0,
						Stamp:       time.Time{},
						Error:       "",
					}}}},
			expectedHistory: []embedded.PTTemperature{{
				Readings: []max31865.Readings{
					{
						ID:          "1",
						Temperature: 1,
						Average:     -100.0,
						Stamp:       time.Time{},
						Error:       "",
					}}}, {
				Readings: []max31865.Readings{{
					ID:          "3",
					Temperature: 1,
					Average:     -100.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "3",
					Temperature: 1,
					Average:     -12300.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Readings: []max31865.Readings{{
					ID:          "2",
					Temperature: 1,
					Average:     -50.0,
					Stamp:       time.Time{},
					Error:       "",
				}, {
					ID:          "2",
					Temperature: 1,
					Average:     -510.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}},
			ids:   []string{"1", "2", "3"},
			temps: []float32{-125.0, -520.0, -150.0},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.NewPTHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.Update(), 0, arg.name)
		history := h.History()
		r.ElementsMatch(arg.expectedHistory, history, arg.name)
		// Second call should be empty without update
		r.Len(h.History(), 0, arg.name)

		// Now verify args
		r.Equal(len(arg.ids), len(arg.temps), arg.name)
		for i := range arg.ids {
			t, err := h.Temperature(arg.ids[i])
			r.Nil(err)
			r.InDelta(arg.temps[i], t, 0.01)
		}

	}
}
func (t *PTTestSuite) TestHistory() {
	args := []struct {
		name            string
		onGet           []embedded.PTSensorConfig
		onTemperatures  []embedded.PTTemperature
		expectedHistory []embedded.PTTemperature
	}{
		{
			name: "single element in history",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     123.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
			expectedHistory: []embedded.PTTemperature{},
		},
		{
			name: "return all  but last element in history",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -100.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -50.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -150.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
			expectedHistory: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -100.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "1",
							Temperature: 1,
							Average:     -50.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
		},
		{
			name: "return all  but last element in history - more and mixed data",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "2",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "3",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -100.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "3",
							Temperature: 1,
							Average:     -100.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "1",
							Temperature: 1,
							Average:     -125.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "3",
							Temperature: 1,
							Average:     -12300.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "2",
							Temperature: 1,
							Average:     -50.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "2",
							Temperature: 1,
							Average:     -510.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "2",
							Temperature: 1,
							Average:     -520.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "3",
							Temperature: 1,
							Average:     -150.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
			expectedHistory: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -100.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "3",
							Temperature: 1,
							Average:     -100.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "3",
							Temperature: 1,
							Average:     -12300.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "2",
							Temperature: 1,
							Average:     -50.0,
							Stamp:       time.Time{},
							Error:       "",
						},
						{
							ID:          "2",
							Temperature: 1,
							Average:     -510.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.NewPTHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.Update(), 0, arg.name)
		history := h.History()

		r.ElementsMatch(arg.expectedHistory, history, arg.name)
	}
}

func (t *PTTestSuite) TestTemperature() {
	args := []struct {
		name                string
		id                  string
		onGet               []embedded.PTSensorConfig
		onTemperatures      []embedded.PTTemperature
		expectedTemperature float32
	}{
		{
			name: "return average",
			id:   "1",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     123.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
			expectedTemperature: 123.0,
		},
		{
			name: "return last average",
			id:   "1",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     123.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     128.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
				{
					Readings: []max31865.Readings{
						{
							ID:          "1",
							Temperature: 1,
							Average:     -200.0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
			expectedTemperature: -200.0,
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.NewPTHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.Update(), 0, arg.name)

		temp, err := h.Temperature(arg.id)
		r.Nil(err, arg.name)
		r.InDelta(arg.expectedTemperature, temp, 0.01, arg.name)
	}
}

func (t *PTTestSuite) TestUpdate_Errors() {
	args := []struct {
		name              string
		onGet             []embedded.PTSensorConfig
		onTemperatures    []embedded.PTTemperature
		onTemperaturesErr error
		expectedErr       []error
	}{
		{
			name: "interface error",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures:    []embedded.PTTemperature{},
			onTemperaturesErr: errors.New("hello world"),
			expectedErr:       []error{errors.New("hello world")},
		},
		{
			name: "unexpected ID",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.PTTemperature{
				{
					Readings: []max31865.Readings{
						{
							ID:          "2",
							Temperature: 1,
							Average:     0,
							Stamp:       time.Time{},
							Error:       "",
						},
					},
				},
			},
			onTemperaturesErr: nil,
			expectedErr:       []error{distillation.ErrUnexpectedID},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, arg.onTemperaturesErr)

		h, err := distillation.NewPTHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		errs := h.Update()
		r.Len(arg.expectedErr, len(errs), arg.name)
		for i := range errs {
			r.ErrorContains(errs[i], arg.expectedErr[i].Error(), arg.name+strconv.FormatInt(int64(i), 10))
		}
	}
}

func (t *PTTestSuite) TestTemperatureErrors() {
	args := []struct {
		name        string
		onGet       []embedded.PTSensorConfig
		id          string
		expectedErr error
	}{
		{
			name: "wrong ID",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			id:          "2",
			expectedErr: distillation.ErrNoSuchID,
		},
		{
			name: "no temps",
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			id:          "1",
			expectedErr: distillation.ErrNoTemps,
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		h, err := distillation.NewPTHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		_, err = h.Temperature(arg.id)
		r.NotNil(err, arg.name)
		r.ErrorContains(err, arg.expectedErr.Error(), arg.name)
	}
}

func (t *PTTestSuite) TestConfigureSensor() {
	args := []struct {
		name        string
		newConfig   distillation.PTConfig
		onGet       []embedded.PTSensorConfig
		onSetErr    error
		errContains string
	}{
		{
			name: "all good",
			newConfig: distillation.PTConfig{PTSensorConfig: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   1,
					ASyncPoll:    false,
					PollInterval: 3,
					Samples:      4,
				},
			}},
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				}},
			onSetErr:    nil,
			errContains: "",
		},
		{
			name: "error on set interface",
			newConfig: distillation.PTConfig{PTSensorConfig: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   1,
					ASyncPoll:    false,
					PollInterval: 3,
					Samples:      4,
				},
			}},
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				}},
			onSetErr:    errors.New("hello"),
			errContains: "hello",
		},
		{
			name: "wrong ID",
			newConfig: distillation.PTConfig{PTSensorConfig: embedded.PTSensorConfig{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   1,
					ASyncPoll:    false,
					PollInterval: 3,
					Samples:      4,
				},
			}},
			onGet: []embedded.PTSensorConfig{
				{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "2",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					},
				}},
			onSetErr:    nil,
			errContains: distillation.ErrNoSuchID.Error(),
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Set", arg.newConfig.PTSensorConfig).Return(arg.onSetErr)
		ds, err := distillation.NewPTHandler(m)
		r.Nil(err, arg.name)

		err = ds.ConfigureSensor(arg.newConfig)
		if arg.errContains != "" {
			r.ErrorContains(err, arg.errContains, arg.name)
			continue
		}
		r.Nil(err, arg.name)
	}
}

func (t *PTTestSuite) TestGetSensors() {
	args := []struct {
		name     string
		onGet    []embedded.PTSensorConfig
		expected []distillation.PTConfig
	}{
		{
			name:     "empty slice",
			onGet:    nil,
			expected: nil,
		},
		{
			name: "single element",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}},
			expected: []distillation.PTConfig{{
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "two sensors",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    true,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "2",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.PTConfig{{
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    true,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "2",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "multiple sensors",
			onGet: []embedded.PTSensorConfig{{
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "1",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "2",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "4",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "5",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Enabled: false,
				SensorConfig: max31865.SensorConfig{
					ID:           "12",
					Correction:   0,
					ASyncPoll:    false,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.PTConfig{{
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "1",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "2",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "4",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "5",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				PTSensorConfig: embedded.PTSensorConfig{
					Enabled: false,
					SensorConfig: max31865.SensorConfig{
						ID:           "12",
						Correction:   0,
						ASyncPoll:    false,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(PTMock)
		m.On("Get").Return(arg.onGet, nil)
		h, err := distillation.NewPTHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.ElementsMatch(arg.expected, h.GetSensors(), arg.name)
	}
}

func (t *PTTestSuite) TestNew() {
	r := t.Require()
	{
		h, err := distillation.NewPTHandler(nil)
		r.Nil(h)
		r.NotNil(err)
		r.ErrorContains(err, distillation.ErrNoPTInterface.Error())
	}
	{
		m := new(PTMock)
		mockErr := errors.New("hello buddy")
		m.On("Get").Return([]embedded.PTSensorConfig{}, mockErr)
		h, err := distillation.NewPTHandler(m)
		r.Nil(h)
		r.NotNil(err)
		r.ErrorContains(err, mockErr.Error())
	}
	{
		m := new(PTMock)
		sensor := embedded.PTSensorConfig{
			Enabled: false,
			SensorConfig: max31865.SensorConfig{
				ID:           "1",
				Correction:   1,
				ASyncPoll:    false,
				PollInterval: 1,
				Samples:      1,
			},
		}

		m.On("Get").Return([]embedded.PTSensorConfig{sensor}, nil)
		h, err := distillation.NewPTHandler(m)
		r.NotNil(h)
		r.Nil(err)

		sensors := []distillation.PTConfig{{PTSensorConfig: sensor}}
		r.ElementsMatch(sensors, h.GetSensors())
	}
}
