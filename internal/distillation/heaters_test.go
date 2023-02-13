/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-clap/iot/internal/distillation"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slices"
)

type HeaterHandlerSuite struct {
	suite.Suite
	req  *http.Request
	resp *httptest.ResponseRecorder
}

type HeaterMock struct {
	mock.Mock
}

func TestHeaters(t *testing.T) {
	suite.Run(t, new(HeaterHandlerSuite))
}

func (t *HeaterHandlerSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
	t.resp = httptest.NewRecorder()
}

func (t *HeaterHandlerSuite) TestRestAPI_ConfigHeater() {
	args := []struct {
		heaters   []embedded.HeaterConfig
		newGlobal []distillation.HeaterConfigGlobal
		newConfig []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "1",
					Enabled: true,
				},
				{
					ID:      "2",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: true,
					Power:   77,
				},
				{
					ID:      "2",
					Enabled: false,
					Power:   91,
				},
			},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "4",
					Enabled: true,
				},
				{
					ID:      "5",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: false,
					Power:   100,
				},
				{
					ID:      "5",
					Enabled: true,
					Power:   0,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}

	r := t.Require()
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newConfig))
		r.Equal(len(arg.heaters), len(arg.newGlobal))

		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.New(distillation.WithHeaters(m))
		r.Nil(err)
		r.NotNil(h)

		for _, glob := range arg.newGlobal {
			r.Nil(h.HeatersHandler.ConfigureGlobal(glob))
		}

		for _, cfg := range arg.newConfig {
			t.resp = httptest.NewRecorder()
			var body bytes.Buffer
			r.Nil(json.NewEncoder(&body).Encode(cfg))

			t.req, _ = http.NewRequest(http.MethodPut, distillation.RoutesConfigEnabledHeater, &body)
			t.req.Header.Add("Content-Type", "application/json")

			// Should be called
			m.Mock.On("Set", cfg).Return(nil).Once()
			h.ServeHTTP(t.resp, t.req)
			r.Equal(http.StatusOK, t.resp.Code, t.resp.Body.String())

			retCfg := embedded.HeaterConfig{}
			r.Nil(json.NewDecoder(t.resp.Body).Decode(&retCfg))

			t.Equal(http.StatusOK, t.resp.Code)
			t.Equal(cfg, retCfg)
		}

	}
}

func (t *HeaterHandlerSuite) Test_RestAPI_ConfigureGlobal() {
	r := t.Require()
	args := []struct {
		heaters   []embedded.HeaterConfig
		newGlobal []distillation.HeaterConfigGlobal
		errs      []error
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "1",
					Enabled: true,
				},
				{
					ID:      "2",
					Enabled: true,
				},
			},
			errs: []error{nil, nil},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "3",
					Enabled: true,
				},
				{
					ID:      "5",
					Enabled: true,
				},
			},
			errs: []error{distillation.ErrNoSuchHeater, nil},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newGlobal))
		r.Equal(len(arg.heaters), len(arg.errs))
		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.New(distillation.WithHeaters(m))
		r.Nil(err)
		r.NotNil(h)

		for i, global := range arg.newGlobal {
			t.resp = httptest.NewRecorder()
			var body bytes.Buffer
			r.Nil(json.NewEncoder(&body).Encode(global))
			t.req, _ = http.NewRequest(http.MethodPut, distillation.RoutesConfigHeater, &body)
			t.req.Header.Add("Content-Type", "application/json")

			h.ServeHTTP(t.resp, t.req)
			if arg.errs[i] != nil {
				r.Equal(http.StatusInternalServerError, t.resp.Code)
				continue
			}
			r.Equal(http.StatusOK, t.resp.Code, t.resp.Body.String())

			retCfg := distillation.HeaterConfigGlobal{}
			r.Nil(json.NewDecoder(t.resp.Body).Decode(&retCfg))

			t.Equal(http.StatusOK, t.resp.Code)
			t.Equal(global, retCfg)
		}

	}

}
func (t *HeaterHandlerSuite) Test_RestAPI_GetEnabled() {
	args := []struct {
		heaters   []embedded.HeaterConfig
		newGlobal []distillation.HeaterConfigGlobal
		newConfig []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "1",
					Enabled: true,
				},
				{
					ID:      "2",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: true,
					Power:   77,
				},
				{
					ID:      "2",
					Enabled: false,
					Power:   91,
				},
			},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "4",
					Enabled: true,
				},
				{
					ID:      "5",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: false,
					Power:   100,
				},
				{
					ID:      "5",
					Enabled: true,
					Power:   0,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}

	r := t.Require()
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newConfig))
		r.Equal(len(arg.heaters), len(arg.newGlobal))

		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, _ := distillation.New(distillation.WithHeaters(m))
		heaters := h.HeatersHandler

		// Get list of enabled heaters
		expectedEnabled := make([]embedded.HeaterConfig, 0, 1)
		for i, glob := range arg.newGlobal {
			r.Nil(heaters.ConfigureGlobal(glob))
			if glob.Enabled {
				expectedEnabled = append(expectedEnabled, arg.newConfig[i])
			}
		}
		// Enable heaters
		for _, newcfg := range arg.newConfig {
			m.Mock.On("Set", newcfg).Return(nil).Once()
			r.Nil(heaters.Configure(newcfg))
		}

		t.req, _ = http.NewRequest(http.MethodGet, distillation.RoutesGetEnabledHeaters, nil)

		h.ServeHTTP(t.resp, t.req)
		r.Equal(http.StatusOK, t.resp.Code)

		var bodyJson []embedded.HeaterConfig
		if err := json.NewDecoder(t.resp.Body).Decode(&bodyJson); err != nil {
			panic(err)
		}
		r.Equal(len(expectedEnabled), len(bodyJson))
		r.ElementsMatch(expectedEnabled, bodyJson)
	}
}

func (t *HeaterHandlerSuite) TestRest_GetGlobalHeaters() {
	r := t.Require()

	args := []struct {
		heaters []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}

	for _, arg := range args {
		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.New(distillation.WithHeaters(m))
		r.Nil(err)

		t.req, _ = http.NewRequest(http.MethodGet, distillation.RoutesGetAllHeaters, nil)

		h.ServeHTTP(t.resp, t.req)
		r.Equal(http.StatusOK, t.resp.Code)

		var bodyJson []distillation.HeaterConfigGlobal
		if err := json.NewDecoder(t.resp.Body).Decode(&bodyJson); err != nil {
			panic(err)
		}
		r.Equal(len(arg.heaters), len(bodyJson))

		expected := make([]distillation.HeaterConfigGlobal, len(arg.heaters))
		for i, heater := range arg.heaters {
			expected[i] = distillation.HeaterConfigGlobal{
				ID:      heater.ID,
				Enabled: false,
			}
		}

		r.ElementsMatch(expected, bodyJson)

	}
}

func (t *HeaterHandlerSuite) Test_GetEnabled() {
	args := []struct {
		heaters   []embedded.HeaterConfig
		newGlobal []distillation.HeaterConfigGlobal
		newConfig []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "1",
					Enabled: true,
				},
				{
					ID:      "2",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: true,
					Power:   77,
				},
				{
					ID:      "2",
					Enabled: false,
					Power:   91,
				},
			},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "4",
					Enabled: true,
				},
				{
					ID:      "5",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: false,
					Power:   100,
				},
				{
					ID:      "5",
					Enabled: true,
					Power:   0,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}

	r := t.Require()
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newConfig))
		r.Equal(len(arg.heaters), len(arg.newGlobal))

		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, _ := distillation.NewHandlerHeaters(m)

		expectedEnabled := make([]embedded.HeaterConfig, 0, 1)
		for i, glob := range arg.newGlobal {
			r.Nil(h.ConfigureGlobal(glob))
			if glob.Enabled {
				expectedEnabled = append(expectedEnabled, arg.newConfig[i])
			}
		}
		for _, newcfg := range arg.newConfig {
			m.Mock.On("Set", newcfg).Return(nil).Once()
			r.Nil(h.Configure(newcfg))
		}

		enabled := h.Configs()
		r.Len(expectedEnabled, len(enabled))
		r.ElementsMatch(expectedEnabled, enabled)
	}
}

func (t *HeaterHandlerSuite) TestConfigureGlobal_DisableEnabledHeater() {
	heaters := []embedded.HeaterConfig{
		{
			ID:      "1",
			Enabled: true,
			Power:   13,
		},
		{
			ID:      "2",
			Enabled: true,
			Power:   99,
		}}

	globalConfig := distillation.HeaterConfigGlobal{
		ID:      "1",
		Enabled: true,
	}

	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}
	r := t.Require()
	m := new(HeaterMock)
	m.Mock.On("Get").Return(heaters, nil).Once()
	// On new handler should set config to initial state - disabled and power = 0
	for _, heater := range heaters {
		expectedConfig.ID = heater.ID
		m.Mock.On("Set", expectedConfig).Return(nil).Once()
	}

	h, err := distillation.NewHandlerHeaters(m)
	r.Nil(err)
	// Enabling heater in global
	r.Nil(h.ConfigureGlobal(globalConfig))

	cfg := embedded.HeaterConfig{
		ID:      heaters[0].ID,
		Enabled: true,
		Power:   13,
	}
	// Enabling heater
	m.On("Set", cfg).Return(nil).Once()
	r.Nil(h.Configure(cfg))

	// Verify
	configs := h.Configs()
	pos := slices.Index(configs, cfg)
	r.True(configs[pos].Enabled)

	// Now configure global should call set and disable heater
	globalConfig.Enabled = false
	cfg.Enabled = false
	m.On("Set", cfg).Return(nil).Once()
	r.Nil(h.ConfigureGlobal(globalConfig))
}

func (t *HeaterHandlerSuite) TestConfigure_EnableHeater() {
	args := []struct {
		heaters   []embedded.HeaterConfig
		newGlobal []distillation.HeaterConfigGlobal
		newConfig []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "1",
					Enabled: true,
				},
				{
					ID:      "2",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: true,
					Power:   77,
				},
				{
					ID:      "2",
					Enabled: false,
					Power:   91,
				},
			},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "4",
					Enabled: true,
				},
				{
					ID:      "5",
					Enabled: true,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: false,
					Power:   100,
				},
				{
					ID:      "5",
					Enabled: true,
					Power:   0,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}

	r := t.Require()
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newConfig))
		r.Equal(len(arg.heaters), len(arg.newGlobal))

		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.NewHandlerHeaters(m)
		r.Nil(err)
		r.NotNil(h)

		for _, glob := range arg.newGlobal {
			r.Nil(h.ConfigureGlobal(glob))
		}

		for _, cfg := range arg.newConfig {
			m.Mock.On("Set", cfg).Return(nil).Once()
			r.Nil(h.Configure(cfg))
		}
	}
}
func (t *HeaterHandlerSuite) TestConfigure_CantEnableHeater() {
	r := t.Require()
	args := []struct {
		heaters       []embedded.HeaterConfig
		newConfig     []embedded.HeaterConfig
		expectedWrite []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newConfig: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: true,
					Power:   17,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   55,
				},
			},
			expectedWrite: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   17,
				},
				{
					ID:      "2",
					Enabled: false,
					Power:   55,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newConfig))
		r.Equal(len(arg.heaters), len(arg.expectedWrite))
		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.NewHandlerHeaters(m)
		r.Nil(err)
		r.NotNil(h)
		for i := range arg.expectedWrite {
			m.Mock.On("Set", arg.expectedWrite[i]).Return(nil).Once()
			r.Nil(h.Configure(arg.newConfig[i]))
		}
	}

}

func (t *HeaterHandlerSuite) TestConfigureGlobal() {
	r := t.Require()
	args := []struct {
		heaters   []embedded.HeaterConfig
		newGlobal []distillation.HeaterConfigGlobal
		errs      []error
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "1",
					Enabled: true,
				},
				{
					ID:      "2",
					Enabled: true,
				},
			},
			errs: []error{nil, nil},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
			newGlobal: []distillation.HeaterConfigGlobal{
				{
					ID:      "3",
					Enabled: true,
				},
				{
					ID:      "5",
					Enabled: true,
				},
			},
			errs: []error{distillation.ErrNoSuchHeater, nil},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}
	for _, arg := range args {
		r.Equal(len(arg.heaters), len(arg.newGlobal))
		r.Equal(len(arg.heaters), len(arg.errs))
		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.NewHandlerHeaters(m)
		r.Nil(err)
		r.NotNil(h)

		for i, global := range arg.newGlobal {
			err = h.ConfigureGlobal(global)
			if arg.errs[i] != nil {
				r.ErrorIs(err, arg.errs[i])
				continue
			}
			r.Nil(err)
		}

	}

}

func (t *HeaterHandlerSuite) TestNew_GetGetGlobal() {
	r := t.Require()

	args := []struct {
		heaters []embedded.HeaterConfig
	}{
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "1",
					Enabled: false,
					Power:   13,
				},
				{
					ID:      "2",
					Enabled: true,
					Power:   99,
				},
			},
		},
		{
			heaters: []embedded.HeaterConfig{
				{
					ID:      "4",
					Enabled: true,
					Power:   16,
				},
				{
					ID:      "5",
					Enabled: false,
					Power:   0,
				},
			},
		},
	}
	expectedConfig := embedded.HeaterConfig{
		ID:      "",
		Enabled: false,
		Power:   0,
	}
	expectedGlobalConfig := distillation.HeaterConfigGlobal{
		ID:      "",
		Enabled: false,
	}
	for _, arg := range args {
		m := new(HeaterMock)
		m.Mock.On("Get").Return(arg.heaters, nil).Once()
		// On new handler should set config to initial state - disabled and power = 0
		for _, heater := range arg.heaters {
			expectedConfig.ID = heater.ID
			m.Mock.On("Set", expectedConfig).Return(nil).Once()
		}

		h, err := distillation.NewHandlerHeaters(m)
		r.Nil(err)

		heaters := h.Configs()
		// And now should return updated config
		for _, heater := range heaters {
			expectedConfig.ID = heater.ID
			r.Equal(expectedConfig, heater)
		}

		global := h.ConfigsGlobal()
		for _, heater := range global {
			expectedGlobalConfig.ID = heater.ID
			r.Equal(expectedGlobalConfig, heater)
		}

	}
}

func (m *HeaterMock) Get() ([]embedded.HeaterConfig, error) {
	args := m.Called()
	return args.Get(0).([]embedded.HeaterConfig), args.Error(1)
}

func (m *HeaterMock) Set(heater embedded.HeaterConfig) error {
	return m.Called(heater).Error(0)
}
