/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"os"
)

var _ Onewire = (*onewire)(nil)

// STD implementation
type onewire struct {
	path string
}

func (o *onewire) WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, 0644)
}

func (o *onewire) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (o *onewire) Path() string {
	return o.path
}
