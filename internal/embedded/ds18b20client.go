/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"time"

	"github.com/a-clap/iot/internal/restclient"
)

type DS18B20Client struct {
	addr    string
	timeout time.Duration
}

func NewDS18B20Client(addr string, timeout time.Duration) *DS18B20Client {
	return &DS18B20Client{addr: addr, timeout: timeout}
}

func (p *DS18B20Client) Get() ([]DSSensorConfig, error) {
	return restclient.Get[[]DSSensorConfig, *Error](p.addr+RoutesGetOnewireSensors, p.timeout)
}

func (p *DS18B20Client) Configure(setConfig DSSensorConfig) (DSSensorConfig, error) {
	return restclient.Put[DSSensorConfig, *Error](p.addr+RoutesConfigOnewireSensor, p.timeout, setConfig)
}

func (p *DS18B20Client) Temperatures() ([]DSTemperature, error) {
	return restclient.Get[[]DSTemperature, *Error](p.addr+RoutesGetOnewireTemperatures, p.timeout)
}
