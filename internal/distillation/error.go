/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

import (
	"time"
)

// Error is common struct returned via rest api
type Error struct {
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Instance  string    `json:"instance"`
	Timestamp time.Time `json:"timestamp"`
}
