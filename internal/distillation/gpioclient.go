/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"time"

	"github.com/a-clap/iot/internal/restclient"
)

type GPIOClient struct {
	addr    string
	timeout time.Duration
}

func NewGPIOClient(addr string, timeout time.Duration) *GPIOClient {
	return &GPIOClient{addr: addr, timeout: timeout}
}

func (h *GPIOClient) Get() ([]GPIOConfig, error) {
	return restclient.Get[[]GPIOConfig, *Error](h.addr+RoutesGetGPIO, h.timeout)
}

func (h *GPIOClient) Configure(setConfig GPIOConfig) (GPIOConfig, error) {
	return restclient.Put[GPIOConfig, *Error](h.addr+RoutesConfigureGPIO, h.timeout, setConfig)
}
