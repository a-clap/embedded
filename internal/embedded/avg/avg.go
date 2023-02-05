/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package avg

import (
	"errors"
	"golang.org/x/exp/constraints"
)

var (
	ErrSizeIsZero = errors.New("size must be greater than zero")
)

type SummableAndDividable interface {
	constraints.Integer | constraints.Float
}

type Avg[T SummableAndDividable] struct {
	buffer []T
	size   uint
}

func New[T SummableAndDividable](size uint) (*Avg[T], error) {
	if size == 0 {
		return nil, ErrSizeIsZero
	}

	return &Avg[T]{
		buffer: make([]T, 0, size),
		size:   size,
	}, nil
}

func (a *Avg[T]) Add(value T) {
	p := uint(0)
	newBufSize := uint(len(a.buffer) + 1)
	if newBufSize > a.size {
		p = newBufSize - a.size
	}
	a.buffer = append(a.buffer[p:], value)

}

func (a *Avg[T]) Average() (avg T) {
	if len(a.buffer) == 0 {
		return
	}

	for _, elem := range a.buffer {
		avg += elem
	}
	return avg / T(len(a.buffer))
}

func (a *Avg[T]) Resize(newSize uint) error {
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
