/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package alphanumeric

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"sync"
)

type Value interface {
	Get() string
	Set(string)
}

type button rune

type alphaNumeric struct {
	Value
	current        string
	buttons        map[button]*widget.Button
	upper, special bool
	w              fyne.Window
	sync.Once
}

var (
	ErrNoAppRunning = errors.New("no app running")
	keyboard        = &alphaNumeric{buttons: make(map[button]*widget.Button)}
)

func Show(v Value) (fyne.Window, error) {
	keyboard.init(v)

	app := fyne.CurrentApp()
	if app == nil {
		return nil, ErrNoAppRunning
	}

	keyboard.w = app.NewWindow("")

	keyboard.updateButtons()
	keyboard.refresh()

	return keyboard.w, nil
}

func (a *alphaNumeric) refresh() {

	split := container.NewHSplit(a.buttons[inp], a.buttons[clr])
	split.SetOffset(1)
	view := container.NewGridWithColumns(1, split)
	for _, current := range keyboard.layout() {
		ctn := container.NewGridWithColumns(len(current))
		for _, elem := range current {
			ctn.Add(elem)
		}
		view.Add(ctn)
	}

	// Create bottom line
	ctn := container.NewGridWithColumns(3,
		container.NewGridWithColumns(2, a.buttons[alt], a.buttons[comma]),
		a.buttons[space],
		container.NewGridWithColumns(2, a.buttons[dot], a.buttons[enter]),
	)
	view.Add(ctn)
	keyboard.w.SetContent(view)
}

func (a *alphaNumeric) layout() [][]*widget.Button {
	lines := func() [][]button {
		return [][]button{
			{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'},
			{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p'},
			{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l'},
			{shift, 'z', 'x', 'c', 'v', 'b', 'n', 'm', bs},
		}
	}()

	buttons := make([][]*widget.Button, len(lines))
	for i, line := range lines {
		buttons[i] = make([]*widget.Button, len(line))
		for j, elem := range line {
			buttons[i][j] = a.buttons[elem]
		}
	}
	return buttons
}

func (a *alphaNumeric) space() {
	a.standardKey(" ")
}

func (a *alphaNumeric) bs() {
	n := len(a.current)
	if n > 0 {
		a.current = a.current[:n-1]
		a.updateInput()
	}
}

func (a *alphaNumeric) alt() {
	a.special = !a.special
	a.updateButtons()
}

func (a *alphaNumeric) enter() {
	a.Value.Set(a.current)
	a.w.Close()
}

func (a *alphaNumeric) shift() {
	a.upper = !a.upper
	a.updateButtons()
}

func (a *alphaNumeric) dot() {
	a.standardKey(".")
}

func (a *alphaNumeric) comma() {
	a.standardKey(",")
}

func (a *alphaNumeric) standardKey(key string) {
	a.current += key
	a.updateInput()
}

func (a *alphaNumeric) updateInput() {
	a.buttons[inp].SetText(a.current)
}

func (a *alphaNumeric) clear() {
	a.current = ""
	a.updateInput()
}
