/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package avg

import (
	"errors"
)

var (
	ErrSizeIsZero = errors.New("size must be greater than zero")
)

type Avg struct {
	buffer []float64
	size   uint
}

func New(size uint) (*Avg, error) {
	if size == 0 {
		return nil, ErrSizeIsZero
	}

	return &Avg{
		buffer: make([]float64, 0, size),
		size:   size,
	}, nil
}

func (a *Avg) Add(value float64) {
	p := uint(0)
	newBufSize := uint(len(a.buffer) + 1)
	if newBufSize > a.size {
		p = newBufSize - a.size
	}
	a.buffer = append(a.buffer[p:], value)

}

func (a *Avg) Average() (avg float64) {
	if len(a.buffer) == 0 {
		return
	}

	for _, elem := range a.buffer {
		avg += elem
	}
	return avg / float64(len(a.buffer))
}

func (a *Avg) Resize(newSize uint) error {
	if newSize == 0 {
		return ErrSizeIsZero
	}
	defer func() {
		a.size = newSize
	}()

	if newSize > a.size {
		return nil
	}

	bufSize := uint(len(a.buffer))
	if bufSize >= newSize {
		a.buffer = a.buffer[bufSize-newSize:]
		return nil
	}

	return nil

}
