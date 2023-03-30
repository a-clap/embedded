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
	"github.com/a-clap/iot/pkg/ds18b20"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DSTestSuite struct {
	suite.Suite
}

type DSMock struct {
	mock.Mock
}

func TestDS(t *testing.T) {
	suite.Run(t, new(DSTestSuite))
}

func (t *DSTestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
}

func (t *DSTestSuite) TestGetSensors_Rest() {
	args := []struct {
		name     string
		onGet    []embedded.DSSensorConfig
		expected []distillation.DSConfig
	}{
		{
			name: "single element",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}},
			expected: []distillation.DSConfig{{
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "two sensors on one bus",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.DSConfig{{
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "2",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "multiple sensors on multiple bus",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "3",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "4",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "3",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "5",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "5",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "12",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.DSConfig{{
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "2",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "3",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "4",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "3",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "5",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "5",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "12",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		h, err := distillation.New(distillation.WithDS(m))
		r.NotNil(h)
		r.Nil(err)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, distillation.RoutesGetDS, nil)

		h.ServeHTTP(recorder, request)
		r.Equal(http.StatusOK, recorder.Code, arg.name)
		var retCfg []distillation.DSConfig

		r.Nil(json.NewDecoder(recorder.Body).Decode(&retCfg), arg.name)
	}
}

func (t *DSTestSuite) TestTemperature_REST() {
	args := []struct {
		name                 string
		onGet                []embedded.DSSensorConfig
		onTemperatures       []embedded.DSTemperature
		expectedTemperatures []distillation.DSTemperature
	}{
		{
			name: "return average",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}},
			onTemperatures: []embedded.DSTemperature{{
				Bus: "1",
				Readings: []ds18b20.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     123.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}},
			expectedTemperatures: []distillation.DSTemperature{{
				ID:          "1",
				Temperature: 123.0,
			}}}, {
			name: "return last average",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}},
			onTemperatures: []embedded.DSTemperature{{
				Bus: "1",
				Readings: []ds18b20.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     123.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Bus: "1",
				Readings: []ds18b20.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     128.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
				Bus: "1",
				Readings: []ds18b20.Readings{{
					ID:          "1",
					Temperature: 1,
					Average:     -200.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}},
			expectedTemperatures: []distillation.DSTemperature{{
				ID:          "1",
				Temperature: 123.0,
			}}}}

	r := t.Require()
	for _, arg := range args {
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.New(distillation.WithDS(m))
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.DSHandler.Update(), 0, arg.name)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, distillation.RoutesGetDSTemperatures, nil)

		h.ServeHTTP(recorder, request)
		r.Equal(http.StatusOK, recorder.Code, arg.name)
		var retCfg []distillation.DSTemperature

		r.Nil(json.NewDecoder(recorder.Body).Decode(&retCfg), arg.name)
	}
}

func (t *DSTestSuite) TestConfigureSensor_REST() {
	args := []struct {
		name        string
		newConfig   distillation.DSConfig
		onGet       []embedded.DSSensorConfig
		onSetErr    error
		errContains string
	}{{
		name: "all good",
		newConfig: distillation.DSConfig{DSSensorConfig: embedded.DSSensorConfig{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   1,
				Resolution:   2,
				PollInterval: 3,
				Samples:      4,
			}}},
		onGet: []embedded.DSSensorConfig{{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}},
		onSetErr:    nil,
		errContains: "",
	}, {
		name: "error on set interface",
		newConfig: distillation.DSConfig{DSSensorConfig: embedded.DSSensorConfig{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   1,
				Resolution:   2,
				PollInterval: 3,
				Samples:      4,
			}}},
		onGet: []embedded.DSSensorConfig{{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}},
		onSetErr:    errors.New("hello"),
		errContains: "hello",
	}, {
		name: "wrong ID",
		newConfig: distillation.DSConfig{DSSensorConfig: embedded.DSSensorConfig{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   1,
				Resolution:   2,
				PollInterval: 3,
				Samples:      4,
			}}},
		onGet: []embedded.DSSensorConfig{{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "2",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}},
		onSetErr:    nil,
		errContains: distillation.ErrNoSuchID.Error(),
	}}
	r := t.Require()
	for _, arg := range args {
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Configure", arg.newConfig.DSSensorConfig).Return(arg.newConfig.DSSensorConfig, arg.onSetErr)
		h, err := distillation.New(distillation.WithDS(m))
		r.Nil(err, arg.name)

		recorder := httptest.NewRecorder()
		var body bytes.Buffer
		r.Nil(json.NewEncoder(&body).Encode(arg.newConfig))

		request, _ := http.NewRequest(http.MethodPut, distillation.RoutesConfigureDS, &body)
		request.Header.Add("Content-Type", "application/json")

		h.ServeHTTP(recorder, request)
		if arg.errContains != "" {
			err := distillation.Error{}
			r.Nil(json.NewDecoder(recorder.Body).Decode(&err), arg.name)
			r.Equal(http.StatusInternalServerError, recorder.Code, arg.name+":"+err.Detail)
			r.Contains(err.Detail, arg.errContains, arg.name)
			continue
		}

		r.Equal(http.StatusOK, recorder.Code, recorder.Body.String())
		retCfg := distillation.DSConfig{}
		r.Nil(json.NewDecoder(recorder.Body).Decode(&retCfg), arg.name)

		r.Equal(arg.newConfig, retCfg, arg.name)
	}
}

func (t *DSTestSuite) TestAfterHistory_StillCanReadData() {
	args := []struct {
		name            string
		onGet           []embedded.DSSensorConfig
		onTemperatures  []embedded.DSTemperature
		expectedHistory []embedded.DSTemperature
		ids             []string
		temps           []float32
	}{{
		name: "single element in history",
		onGet: []embedded.DSSensorConfig{{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}},
		onTemperatures: []embedded.DSTemperature{{
			Bus: "1",
			Readings: []ds18b20.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     123.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}},
		expectedHistory: []embedded.DSTemperature{},
		ids:             []string{"1"},
		temps:           []float32{123.0},
	}, {
		name: "return all  but last element in history",
		onGet: []embedded.DSSensorConfig{
			{Bus: "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}},
		onTemperatures: []embedded.DSTemperature{{
			Bus: "1",
			Readings: []ds18b20.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -100.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}, {
			Bus: "1",
			Readings: []ds18b20.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -50.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}, {
			Bus: "1",
			Readings: []ds18b20.Readings{{
				ID:          "1",
				Temperature: 1,
				Average:     -150.0,
				Stamp:       time.Time{},
				Error:       "",
			}}}},
		expectedHistory: []embedded.DSTemperature{{
			Bus: "1",
			Readings: []ds18b20.Readings{{
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
	}, {
		name: "return all  but last element in history - more and mixed data",
		onGet: []embedded.DSSensorConfig{{
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}, {
			Bus:     "2",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "2",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}, {
			Bus:     "1",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "3",
				Correction:   0,
				Resolution:   0,
				PollInterval: 0,
				Samples:      0,
			}}},
		onTemperatures: []embedded.DSTemperature{{
			Bus: "1",
			Readings: []ds18b20.Readings{{
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
			Bus: "2",
			Readings: []ds18b20.Readings{{
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
			Bus: "1",
			Readings: []ds18b20.Readings{
				{
					ID:          "3",
					Temperature: 1,
					Average:     -150.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}},
		expectedHistory: []embedded.DSTemperature{{
			Bus: "1",
			Readings: []ds18b20.Readings{
				{
					ID:          "1",
					Temperature: 1,
					Average:     -100.0,
					Stamp:       time.Time{},
					Error:       "",
				}}}, {
			Bus: "1",
			Readings: []ds18b20.Readings{{
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
			Bus: "2",
			Readings: []ds18b20.Readings{{
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
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.NewDSHandler(m)
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
func (t *DSTestSuite) TestHistory() {
	args := []struct {
		name            string
		onGet           []embedded.DSSensorConfig
		onTemperatures  []embedded.DSTemperature
		expectedHistory []embedded.DSTemperature
	}{
		{
			name: "single element in history",
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
			expectedHistory: []embedded.DSTemperature{},
		},
		{
			name: "return all  but last element in history",
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "1",
					Readings: []ds18b20.Readings{
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
			expectedHistory: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
				{
					Bus:     "2",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "2",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "3",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "2",
					Readings: []ds18b20.Readings{
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
					Bus: "1",
					Readings: []ds18b20.Readings{
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
			expectedHistory: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "2",
					Readings: []ds18b20.Readings{
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
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.NewDSHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.Update(), 0, arg.name)
		history := h.History()

		r.ElementsMatch(arg.expectedHistory, history, arg.name)
	}
}

func (t *DSTestSuite) TestTemperature() {
	args := []struct {
		name                string
		id                  string
		onGet               []embedded.DSSensorConfig
		onTemperatures      []embedded.DSTemperature
		expectedTemperature float32
	}{
		{
			name: "return average",
			id:   "1",
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "1",
					Readings: []ds18b20.Readings{
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
					Bus: "1",
					Readings: []ds18b20.Readings{
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
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, nil)

		h, err := distillation.NewDSHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		r.Len(h.Update(), 0, arg.name)

		temp, err := h.Temperature(arg.id)
		r.Nil(err, arg.name)
		r.InDelta(arg.expectedTemperature, temp, 0.01, arg.name)
	}
}

func (t *DSTestSuite) TestUpdate_Errors() {
	args := []struct {
		name              string
		onGet             []embedded.DSSensorConfig
		onTemperatures    []embedded.DSTemperature
		onTemperaturesErr error
		expectedErr       []error
	}{
		{
			name: "interface error",
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures:    []embedded.DSTemperature{},
			onTemperaturesErr: errors.New("hello world"),
			expectedErr:       []error{errors.New("hello world")},
		},
		{
			name: "unexpected ID",
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				},
			},
			onTemperatures: []embedded.DSTemperature{
				{
					Bus: "1",
					Readings: []ds18b20.Readings{
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
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Temperatures").Return(arg.onTemperatures, arg.onTemperaturesErr)

		h, err := distillation.NewDSHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		errs := h.Update()
		r.Len(arg.expectedErr, len(errs), arg.name)
		for i := range errs {
			r.ErrorContains(errs[i], arg.expectedErr[i].Error(), arg.name+strconv.FormatInt(int64(i), 10))
		}
	}
}

func (t *DSTestSuite) TestTemperatureErrors() {
	args := []struct {
		name        string
		onGet       []embedded.DSSensorConfig
		id          string
		expectedErr error
	}{
		{
			name: "wrong ID",
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
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
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
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
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		h, err := distillation.NewDSHandler(m)
		r.NotNil(h, arg.name)
		r.Nil(err, arg.name)

		_, err = h.Temperature(arg.id)
		r.NotNil(err, arg.name)
		r.ErrorContains(err, arg.expectedErr.Error(), arg.name)
	}
}

func (t *DSTestSuite) TestConfigureSensor() {
	args := []struct {
		name        string
		newConfig   distillation.DSConfig
		onGet       []embedded.DSSensorConfig
		onSetErr    error
		errContains string
	}{
		{
			name: "all good",
			newConfig: distillation.DSConfig{DSSensorConfig: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			}},
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				}},
			onSetErr:    nil,
			errContains: "",
		},
		{
			name: "error on set interface",
			newConfig: distillation.DSConfig{DSSensorConfig: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			}},
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					},
				}},
			onSetErr:    errors.New("hello"),
			errContains: "hello",
		},
		{
			name: "wrong ID",
			newConfig: distillation.DSConfig{DSSensorConfig: embedded.DSSensorConfig{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   1,
					Resolution:   2,
					PollInterval: 3,
					Samples:      4,
				},
			}},
			onGet: []embedded.DSSensorConfig{
				{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "2",
						Correction:   0,
						Resolution:   0,
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
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		m.On("Configure", arg.newConfig.DSSensorConfig).Return(arg.newConfig.DSSensorConfig, arg.onSetErr)
		ds, err := distillation.NewDSHandler(m)
		r.Nil(err, arg.name)

		cfg, err := ds.ConfigureSensor(arg.newConfig)
		if arg.errContains != "" {
			r.ErrorContains(err, arg.errContains, arg.name)
			continue
		}
		r.EqualValues(arg.newConfig, cfg)
		r.Nil(err, arg.name)
	}
}

func (t *DSTestSuite) TestGetSensors() {
	args := []struct {
		name     string
		onGet    []embedded.DSSensorConfig
		expected []distillation.DSConfig
	}{
		{
			name:     "empty slice",
			onGet:    nil,
			expected: nil,
		},
		{
			name: "single element",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}},
			expected: []distillation.DSConfig{{
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "two sensors on one bus",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.DSConfig{{
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "2",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
		{
			name: "multiple sensors on multiple bus",
			onGet: []embedded.DSSensorConfig{{
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "1",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "1",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "2",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "3",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "4",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "3",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "5",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}}, {
				Bus:     "5",
				Enabled: false,
				SensorConfig: ds18b20.SensorConfig{
					ID:           "12",
					Correction:   0,
					Resolution:   0,
					PollInterval: 0,
					Samples:      0,
				}},
			},
			expected: []distillation.DSConfig{{
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "1",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "1",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "2",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "3",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "4",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "3",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "5",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}, {
				DSSensorConfig: embedded.DSSensorConfig{
					Bus:     "5",
					Enabled: false,
					SensorConfig: ds18b20.SensorConfig{
						ID:           "12",
						Correction:   0,
						Resolution:   0,
						PollInterval: 0,
						Samples:      0,
					}}}},
		},
	}
	r := t.Require()
	for _, arg := range args {
		m := new(DSMock)
		m.On("Get").Return(arg.onGet, nil)
		h, err := distillation.NewDSHandler(m)
		r.NotNil(h)
		r.Nil(err)

		r.ElementsMatch(arg.expected, h.GetSensors())
	}
}

func (t *DSTestSuite) TestNew() {
	r := t.Require()
	{
		h, err := distillation.NewDSHandler(nil)
		r.Nil(h)
		r.NotNil(err)
		r.ErrorContains(err, distillation.ErrNoDSInterface.Error())
	}
	{
		m := new(DSMock)
		mockErr := errors.New("hello buddy")
		m.On("Get").Return([]embedded.DSSensorConfig{}, mockErr)
		h, err := distillation.NewDSHandler(m)
		r.Nil(h)
		r.NotNil(err)
		r.ErrorContains(err, mockErr.Error())
	}
	{
		m := new(DSMock)
		sensor := embedded.DSSensorConfig{
			Bus:     "blah",
			Enabled: false,
			SensorConfig: ds18b20.SensorConfig{
				ID:           "1",
				Correction:   1,
				Resolution:   1,
				PollInterval: 1,
				Samples:      1,
			},
		}

		m.On("Get").Return([]embedded.DSSensorConfig{sensor}, nil)
		h, err := distillation.NewDSHandler(m)
		r.NotNil(h)
		r.Nil(err)

		sensors := []distillation.DSConfig{{DSSensorConfig: sensor}}
		r.ElementsMatch(sensors, h.GetSensors())
	}
}

func (m *DSMock) Get() ([]embedded.DSSensorConfig, error) {
	args := m.Called()
	return args.Get(0).([]embedded.DSSensorConfig), args.Error(1)
}

func (m *DSMock) Configure(s embedded.DSSensorConfig) (embedded.DSSensorConfig, error) {
	args := m.Called(s)
	return args.Get(0).(embedded.DSSensorConfig), args.Error(1)
}

func (m *DSMock) Temperatures() ([]embedded.DSTemperature, error) {
	args := m.Called()
	return args.Get(0).([]embedded.DSTemperature), args.Error(1)
}
