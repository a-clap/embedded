/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

import (
	"errors"
)

var (
	ErrNoHeaters            = errors.New("can't execute process without heaters")
	ErrNoTSensors           = errors.New("can't execute process without temperature sensors")
	ErrPhasesBelowZero      = errors.New("phases must be greater than 0")
	ErrNoSuchPhase          = errors.New("requested phase number doesn't exist")
	ErrWrongHeaterID        = errors.New("requested heater ID doesn't exist")
	ErrWrongSensorID        = errors.New("requested sensor ID doesn't exist")
	ErrWrongGpioID          = errors.New("requested gpio ID doesn't exist")
	ErrWrongHeaterPower     = errors.New("power must be in range <0, 100>")
	ErrDifferentGPIOSConfig = errors.New("different number of configs and gpios")
	ErrNoHeatersInConfig    = errors.New("no heaters in config")
	ErrByTimeWrongTime      = errors.New("chosen ByTime, but time must be greater than 0")
	ErrByTemperatureWrongID = errors.New("chosen ByTemperature, but id doesn't exist")
	ErrUnknownType          = errors.New("MoveToNextType unknown")
	ErrAlreadyRunning       = errors.New("already running")
	ErrNotRunning           = errors.New("not running")
)
