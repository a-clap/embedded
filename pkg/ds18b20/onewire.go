/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package ds18b20

import (
	"os"
)

var _ Onewire = (*onewire)(nil)

type onewire struct {
	path string
}

func (o *onewire) Path() string {
	return o.path
}

func (o *onewire) ReadDir(dirname string) ([]string, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, len(entries))
	for i, e := range entries {
		dirs[i] = e.Name()
	}
	return dirs, nil
}

func (o *onewire) Open(name string) (File, error) {
	return os.Open(name)
}
