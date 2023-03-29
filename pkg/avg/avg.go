/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package avg

type Avg struct {
	buffer []float64
	size   uint
}

// New creates Avg struct
// Minimum size is 1, 0 is silently changed to 1
func New(size uint) *Avg {
	if size == 0 {
		size = 1
	}

	return &Avg{
		buffer: make([]float64, 0, size),
		size:   size,
	}
}

// Add adds value to buffer
func (a *Avg) Add(value float64) {
	p := uint(0)
	newBufSize := uint(len(a.buffer) + 1)
	if newBufSize > a.size {
		p = newBufSize - a.size
	}
	a.buffer = append(a.buffer[p:], value)

}

// Average returns current average value based on internal buffer
func (a *Avg) Average() (avg float64) {
	if len(a.buffer) == 0 {
		return
	}

	for _, elem := range a.buffer {
		avg += elem
	}
	return avg / float64(len(a.buffer))
}

// Resize changes internal buffer
// Minimum size is 1, 0 is silently changed to 1
func (a *Avg) Resize(newSize uint) {
	if newSize == 0 {
		newSize = 1
	}

	if newSize > a.size {
		a.size = newSize
		return
	}

	a.size = newSize
	bufSize := uint(len(a.buffer))
	if bufSize >= newSize {
		a.buffer = a.buffer[bufSize-newSize:]
	}

}
