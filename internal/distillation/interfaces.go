/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

type TemperatureSensor interface {
	Temperature() (float32, error)
}

type Heater interface {
	Set(power uint) error
}
