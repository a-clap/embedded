/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

import (
	"time"
)

type clock struct {
}

func (*clock) Unix() int64 {
	return time.Now().Unix()
}
