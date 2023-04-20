/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"time"

	"github.com/a-clap/embedded/pkg/restclient"
)

type PTClient struct {
	addr    string
	timeout time.Duration
}

func NewPTClient(addr string, timeout time.Duration) *PTClient {
	return &PTClient{addr: addr, timeout: timeout}
}

func (p *PTClient) Get() ([]PTSensorConfig, error) {
	return restclient.Get[[]PTSensorConfig, *Error](p.addr+RoutesGetPT100Sensors, p.timeout)
}

func (p *PTClient) Configure(setConfig PTSensorConfig) (PTSensorConfig, error) {
	return restclient.Put[PTSensorConfig, *Error](p.addr+RoutesConfigPT100Sensor, p.timeout, setConfig)
}

func (p *PTClient) Temperatures() ([]PTTemperature, error) {
	return restclient.Get[[]PTTemperature, *Error](p.addr+RoutesGetPT100Temperatures, p.timeout)
}
