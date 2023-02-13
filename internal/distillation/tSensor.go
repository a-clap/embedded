/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

// TSensorConfig is structure to configure each temperature sensor, no matter what type
type TSensorConfig struct {
	ID          string  `json:"id"`
	Enabled     bool    `json:"enabled"`
	Correction  float32 `json:"correction"`
	Temperature float32 `json:"temperature"`
}
