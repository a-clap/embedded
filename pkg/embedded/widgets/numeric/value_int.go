/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package numeric

type implInt struct {
	Value
	current string
}

var _ impl = &implInt{}

func newNumericInt(v Value) impl {
	return &implInt{
		Value:   v,
		current: v.Get(),
	}
}

func (v *implInt) Clear() {
	v.current = "0"
}
func (v *implInt) Backspace() {
	n := len(v.current)
	if n <= 1 {
		v.Clear()
		return
	}
	v.current = v.current[:n-1]
}
func (v *implInt) Dot() {
	// not implemented for int
}
func (v *implInt) Digit(d string) {
	if v.current == "0" {
		v.current = ""
	}
	v.current += d
}

func (v *implInt) Get() string {
	return v.current
}

func (v *implInt) Enter() {
	v.Value.Set(v.current)
}

func (v *implInt) Minus() {
	if len(v.current) == 0 {
		return
	}

	if v.current[0] == '-' {
		v.current = v.current[1:]
		return
	}
	v.current = "-" + v.current

}
