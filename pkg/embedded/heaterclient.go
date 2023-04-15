/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"time"
	
	"github.com/a-clap/iot/pkg/restclient"
)

type HeaterClient struct {
	addr    string
	timeout time.Duration
}

func NewHeaterClient(addr string, timeout time.Duration) *HeaterClient {
	return &HeaterClient{addr: addr, timeout: timeout}
}

func (p *HeaterClient) Get() ([]HeaterConfig, error) {
	return restclient.Get[[]HeaterConfig, *Error](p.addr+RoutesGetHeaters, p.timeout)
}

func (p *HeaterClient) Configure(setConfig HeaterConfig) (HeaterConfig, error) {
	return restclient.Put[HeaterConfig, *Error](p.addr+RoutesConfigHeater, p.timeout, setConfig)
}
