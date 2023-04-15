/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package alphanumeric

import "fyne.io/fyne/v2/widget"

type standardButton struct {
	key    button
	labels [3]string // lower, upper, special
}

// specialButton
type specialButton struct {
	key     button
	label   string // label is always same for special button
	handler func()
}

const (
	lower   = 0
	upper   = 1
	special = 2
)

const (
	esc   = '\x1B'
	bs    = '\x08'
	alt   = '\x1A'
	enter = '\x0A'
	shift = '\x11'
	dot   = '.'
	comma = ','
	space = '\x20'
	clr   = '\x7F' // use DEL as clr
)

// init creates needed buttons
func (a *alphaNumeric) init(v Value) {
	keyboard.Value = v
	keyboard.upper = false
	keyboard.special = false

	a.Once.Do(func() {
		for _, standard := range standardButtons {
			a.buttons[standard.key] = widget.NewButton("", nil)
		}

		for _, special := range specialButtons {
			a.buttons[special.key] = widget.NewButton(special.label, special.handler)
		}

		a.buttons[enter].Importance = widget.HighImportance
	})

	a.current = a.Value.Get()
	a.updateInput()
}

func (a *alphaNumeric) updateButtons() {
	var label int
	switch {
	case a.special:
		label = special
	case a.upper:
		label = upper
	case !a.upper:
		label = lower
	}

	for _, standard := range standardButtons {
		lbl := standard.labels[label]
		a.buttons[standard.key].SetText(lbl)
		a.buttons[standard.key].OnTapped = func() {
			a.standardKey(lbl)
		}
	}
}

var (
	standardButtons = []standardButton{
		{key: '1', labels: [3]string{lower: "1", upper: "!", special: ""}},
		{key: '2', labels: [3]string{lower: "2", upper: "@", special: ""}},
		{key: '3', labels: [3]string{lower: "3", upper: "#", special: ""}},
		{key: '4', labels: [3]string{lower: "4", upper: "$", special: ""}},
		{key: '5', labels: [3]string{lower: "5", upper: "%", special: ""}},
		{key: '6', labels: [3]string{lower: "6", upper: "^", special: ""}},
		{key: '7', labels: [3]string{lower: "7", upper: "&", special: ""}},
		{key: '8', labels: [3]string{lower: "8", upper: "*", special: ""}},
		{key: '9', labels: [3]string{lower: "9", upper: "(", special: ""}},
		{key: '0', labels: [3]string{lower: "0", upper: ")", special: ""}},
		{key: 'q', labels: [3]string{lower: "q", upper: "Q", special: ""}},
		{key: 'w', labels: [3]string{lower: "w", upper: "W", special: ""}},
		{key: 'e', labels: [3]string{lower: "e", upper: "E", special: ""}},
		{key: 'r', labels: [3]string{lower: "r", upper: "R", special: ""}},
		{key: 't', labels: [3]string{lower: "t", upper: "T", special: ""}},
		{key: 'y', labels: [3]string{lower: "y", upper: "Y", special: ""}},
		{key: 'u', labels: [3]string{lower: "u", upper: "U", special: ""}},
		{key: 'i', labels: [3]string{lower: "i", upper: "I", special: ""}},
		{key: 'o', labels: [3]string{lower: "o", upper: "O", special: ""}},
		{key: 'p', labels: [3]string{lower: "p", upper: "P", special: ""}},
		{key: 'a', labels: [3]string{lower: "a", upper: "A", special: ""}},
		{key: 's', labels: [3]string{lower: "s", upper: "S", special: ""}},
		{key: 'd', labels: [3]string{lower: "d", upper: "D", special: ""}},
		{key: 'f', labels: [3]string{lower: "f", upper: "F", special: ""}},
		{key: 'g', labels: [3]string{lower: "g", upper: "G", special: ""}},
		{key: 'h', labels: [3]string{lower: "h", upper: "H", special: ""}},
		{key: 'j', labels: [3]string{lower: "j", upper: "J", special: ""}},
		{key: 'k', labels: [3]string{lower: "k", upper: "K", special: ""}},
		{key: 'l', labels: [3]string{lower: "l", upper: "L", special: ""}},
		{key: 'z', labels: [3]string{lower: "z", upper: "Z", special: ""}},
		{key: 'x', labels: [3]string{lower: "x", upper: "X", special: ""}},
		{key: 'c', labels: [3]string{lower: "c", upper: "C", special: ""}},
		{key: 'v', labels: [3]string{lower: "v", upper: "V", special: ""}},
		{key: 'b', labels: [3]string{lower: "b", upper: "B", special: ""}},
		{key: 'n', labels: [3]string{lower: "n", upper: "N", special: ""}},
		{key: 'm', labels: [3]string{lower: "m", upper: "M", special: ""}},
	}

	specialButtons = []specialButton{
		{key: esc, label: "ESC", handler: func() { keyboard.w.Close() }},
		{key: space, label: " ", handler: func() { keyboard.space() }},
		{key: bs, label: "BS", handler: func() { keyboard.bs() }},
		{key: alt, label: "Alt", handler: func() { keyboard.alt() }},
		{key: enter, label: "=", handler: func() { keyboard.enter() }},
		{key: shift, label: "Shift", handler: func() { keyboard.shift() }},
		{key: dot, label: ".", handler: func() { keyboard.dot() }},
		{key: comma, label: ",", handler: func() { keyboard.comma() }},
		{key: clr, label: "CLR", handler: func() { keyboard.clear() }},
	}
)
