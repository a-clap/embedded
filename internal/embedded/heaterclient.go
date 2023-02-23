/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

// type HeaterClient struct {
// 	addr    string
// 	timeout time.Duration
// }
//
// func NewHeaterClient(addr string, timeout time.Duration) *HeaterClient {
// 	return &HeaterClient{addr: addr, timeout: timeout}
// }
//
// func (p *HeaterClient) GetSensors() ([]HeaterConfig, error) {
// 	return restclient.Get[[]HeaterConfig, *Error](p.addr+RoutesGetPT, p.timeout)
// }
//
// func (p *HeaterClient) Configure(setConfig HeaterConfig) (HeaterConfig, error) {
// 	return restclient.Put[HeaterConfig, *Error](p.addr+RoutesConfigurePT, p.timeout, setConfig)
// }
//
// func (p *HeaterClient) Temperatures() ([]PTTemperature, error) {
// 	return restclient.Get[[]PTTemperature, *Error](p.addr+RoutesGetPTTemperatures, p.timeout)
// }
//
