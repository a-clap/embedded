/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

type MoveToNextType int

const (
	ByTime MoveToNextType = iota
	ByTemperature
)

type MoveToNext func() bool

type Phase struct {
	config *ConfigPhase
}
