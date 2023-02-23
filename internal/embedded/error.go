/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNoSuchID = errors.New("specified ID doesnt' exist")
)

// Error is common struct returned via rest api
type Error struct {
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Instance  string    `json:"instance"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%s %s:%v", e.Title, e.Detail, e.Instance, e.Timestamp)
}
