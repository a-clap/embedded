/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"time"

	"github.com/a-clap/iot/internal/restclient"
)

type DSClient struct {
	addr    string
	timeout time.Duration
}

func NewDSClient(addr string, timeout time.Duration) *DSClient {
	return &DSClient{addr: addr, timeout: timeout}
}

func (d *DSClient) GetSensors() ([]DSConfig, error) {
	return restclient.Get[[]DSConfig, *Error](d.addr+RoutesGetDS, d.timeout)
}

func (d *DSClient) Configure(setConfig DSConfig) (DSConfig, error) {
	return restclient.Put[DSConfig, *Error](d.addr+RoutesGetDS, d.timeout, setConfig)
}

func (d *DSClient) Temperatures() ([]DSTemperature, error) {
	return restclient.Get[[]DSTemperature, *Error](d.addr+RoutesGetDSTemperatures, d.timeout)
}
