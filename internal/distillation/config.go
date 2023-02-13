/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

type Config struct {
	PhaseCount int
	TSensors   []TSensorConfig
	Heaters    []HeaterConfigGlobal
}

type ConfigPhase struct {
	Heaters                []HeaterConfig
	Move                   MoveToNextType
	TimeToMove             int
	TemperatureThreshold   float32
	TemperatureTimeSustain int
}
