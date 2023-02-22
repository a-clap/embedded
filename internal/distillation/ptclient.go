/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"time"

	"github.com/a-clap/iot/internal/restclient"
)

type PTClient struct {
	addr    string
	timeout time.Duration
}

func NewPTClient(addr string, timeout time.Duration) *PTClient {
	return &PTClient{addr: addr, timeout: timeout}
}

func (p *PTClient) GetSensors() ([]PTConfig, error) {
	return restclient.Get[[]PTConfig, *Error](p.addr+RoutesGetPT, p.timeout)
}

func (p *PTClient) Configure(setConfig PTConfig) (PTConfig, error) {
	return restclient.Put[PTConfig, *Error](p.addr+RoutesConfigurePT, p.timeout, setConfig)
}

func (p *PTClient) Temperatures() ([]PTTemperature, error) {
	return restclient.Get[[]PTTemperature, *Error](p.addr+RoutesGetPTTemperatures, p.timeout)
}
