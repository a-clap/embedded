/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import "github.com/a-clap/iot/internal/embedded/models"

type Heaters interface {
	Heaters() []models.HeaterConfig
	SetHeater(config models.HeaterConfig) error
}

type HeatersHandler struct {
}
