/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"time"
	
	"github.com/a-clap/iot/pkg/restclient"
)

type GPIOClient struct {
	addr    string
	timeout time.Duration
}

func NewGPIOClient(addr string, timeout time.Duration) *GPIOClient {
	return &GPIOClient{addr: addr, timeout: timeout}
}

func (p *GPIOClient) Get() ([]GPIOConfig, error) {
	return restclient.Get[[]GPIOConfig, *Error](p.addr+RoutesGetGPIOs, p.timeout)
}

func (p *GPIOClient) Configure(setConfig GPIOConfig) (GPIOConfig, error) {
	return restclient.Put[GPIOConfig, *Error](p.addr+RoutesConfigGPIO, p.timeout, setConfig)
}
