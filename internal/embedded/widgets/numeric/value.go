/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package numeric

import (
	"strings"
)

type impl interface {
	Value
	Clear()
	Backspace()
	Dot()
	Digit(d string)
	Enter()
	Minus()
}

func newImpl(v Value) impl {
	current := v.Get()
	var val impl
	// make it simple, for now
	if strings.Contains(current, ".") {
		val = newFloat(v)
	} else {
		val = newNumericInt(v)
	}
	return val
}
