/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"strconv"
	"strings"
	"time"

	"github.com/a-clap/iot/internal/restclient"
)

type ProcessClient struct {
	addr    string
	timeout time.Duration
}

func NewProcessClient(addr string, timeout time.Duration) *ProcessClient {
	return &ProcessClient{addr: addr, timeout: timeout}
}

func (h *ProcessClient) GetPhaseCount() (ProcessPhaseCount, error) {
	return restclient.Get[ProcessPhaseCount, *Error](h.addr+RoutesProcessPhases, h.timeout)
}

func (h *ProcessClient) GetPhaseConfig(phaseNumber int) (ProcessPhaseConfig, error) {
	p := strconv.FormatInt(int64(phaseNumber), 10)
	addr := strings.Replace(RoutesProcessConfigPhase, ":id", p, 1)
	return restclient.Get[ProcessPhaseConfig, *Error](h.addr+addr, h.timeout)
}

func (h *ProcessClient) ConfigurePhaseCount(count ProcessPhaseCount) (ProcessPhaseCount, error) {
	return restclient.Put[ProcessPhaseCount, *Error](h.addr+RoutesProcessPhases, h.timeout, count)
}

func (h *ProcessClient) ConfigurePhase(phaseNumber int, setConfig ProcessPhaseConfig) (ProcessPhaseConfig, error) {
	p := strconv.FormatInt(int64(phaseNumber), 10)
	addr := strings.Replace(RoutesProcessConfigPhase, ":id", p, 1)
	return restclient.Put[ProcessPhaseConfig, *Error](h.addr+addr, h.timeout, setConfig)
}

func (h *ProcessClient) ValidateConfig() (ProcessConfigValidation, error) {
	return restclient.Get[ProcessConfigValidation, *Error](RoutesProcessConfigValidate, h.timeout)
}

func (h *ProcessClient) ConfigureProcess(cfg ProcessConfig) (ProcessConfig, error) {
	return restclient.Put[ProcessConfig, *Error](RoutesProcess, h.timeout, cfg)
}

func (h *ProcessClient) Status() (ProcessStatus, error) {
	return restclient.Get[ProcessStatus, *Error](RoutesProcessStatus, h.timeout)
}
