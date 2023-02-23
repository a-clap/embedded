/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"time"

	"github.com/a-clap/iot/internal/restclient"
)

type HeaterClient struct {
	addr    string
	timeout time.Duration
}

func NewHeaterClient(addr string, timeout time.Duration) *HeaterClient {
	return &HeaterClient{addr: addr, timeout: timeout}
}

func (h *HeaterClient) GetEnabled() ([]HeaterConfig, error) {
	return restclient.Get[[]HeaterConfig, *Error](h.addr+RoutesGetEnabledHeaters, h.timeout)
}

func (h *HeaterClient) GetAll() ([]HeaterConfigGlobal, error) {
	return restclient.Get[[]HeaterConfigGlobal, *Error](h.addr+RoutesGetAllHeaters, h.timeout)
}

func (h *HeaterClient) Enable(setConfig HeaterConfigGlobal) (HeaterConfigGlobal, error) {
	return restclient.Put[HeaterConfigGlobal, *Error](h.addr+RoutesEnableHeater, h.timeout, setConfig)
}

func (h *HeaterClient) Configure(setConfig HeaterConfig) (HeaterConfig, error) {
	return restclient.Put[HeaterConfig, *Error](h.addr+RoutesConfigureHeater, h.timeout, setConfig)
}
