/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package numeric

import (
	"errors"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Value interface {
	Set(val string)
	Get() string
}

type button rune

type buttonHandler struct {
	key     button
	label   string
	handler func()
}

const (
	esc   button = '\x1B'
	bs           = '\x08'
	enter        = '\x0A'
	clr          = '\x7F' // use DEL as clr
	inp          = '\xFF'
	dot          = '.'
	minus        = '-'
)

var (
	ErrNoAppRunning = errors.New("no app running")
	numericKeyboard = &numeric{
		w:       nil,
		impl:    nil,
		buttons: make(map[button]*widget.Button),
	}

	specialButtons = []buttonHandler{
		{key: esc, label: "ESC", handler: func() { numericKeyboard.w.Hide() }},
		{key: bs, label: "BS", handler: func() { numericKeyboard.impl.Backspace(); numericKeyboard.update() }},
		{key: enter, label: "=", handler: func() { numericKeyboard.impl.Enter(); numericKeyboard.w.Hide() }},
		{key: clr, label: "CLR", handler: func() { numericKeyboard.impl.Clear(); numericKeyboard.update() }},
		{key: inp, label: "", handler: nil}, {key: dot, label: ".", handler: func() { numericKeyboard.impl.Dot(); numericKeyboard.update() }},
		{key: minus, label: "-", handler: func() { numericKeyboard.impl.Minus(); numericKeyboard.update() }},
	}
)

type numeric struct {
	impl    impl
	buttons map[button]*widget.Button
	w       fyne.Window
	once    sync.Once
	ctn     *fyne.Container
}

func Show(v Value) (fyne.Window, error) {
	app := fyne.CurrentApp()
	if app == nil {
		return nil, ErrNoAppRunning
	}
	numericKeyboard.impl = newImpl(v)
	numericKeyboard.init(app)
	numericKeyboard.update()

	return numericKeyboard.w, nil
}

func (n *numeric) update() {
	n.buttons[inp].SetText(n.impl.Get())
}

func (n *numeric) init(app fyne.App) {
	n.once.Do(func() {
		for _, btn := range specialButtons {
			n.buttons[btn.key] = widget.NewButton(btn.label, btn.handler)
		}
		n.buttons[enter].Importance = widget.HighImportance

		for i := 0; i < 10; i++ {
			v := strconv.Itoa(i)
			n.buttons[button(i)+'0'] = widget.NewButton(v, func() {
				n.impl.Digit(v)
				n.update()
			})
		}

		layout := n.layout()
		n.ctn = container.NewGridWithColumns(1,
			n.buttons[inp],
		)

		for _, line := range layout {
			ctn := container.NewGridWithColumns(len(line))
			for _, elem := range line {
				ctn.Add(elem)
			}
			n.ctn.Add(ctn)
		}

		drv := app.Driver()
		if drv, ok := drv.(desktop.Driver); ok {
			numericKeyboard.w = drv.CreateSplashWindow()
		} else {
			numericKeyboard.w = app.NewWindow("")
		}

		numericKeyboard.w.SetContent(numericKeyboard.ctn)
		numericKeyboard.w.SetFixedSize(true)
		numericKeyboard.w.CenterOnScreen()

	})
}

func (n *numeric) layout() [][]*widget.Button {
	lines := func() [][]button {
		return [][]button{
			{'1', '2', '3', esc},
			{'4', '5', '6', clr},
			{'7', '8', '9', bs},
			{dot, '0', minus, enter},
		}
	}()

	buttons := make([][]*widget.Button, len(lines))
	for i, line := range lines {
		buttons[i] = make([]*widget.Button, len(line))
		for j, elem := range line {
			buttons[i][j] = n.buttons[elem]
		}
	}
	return buttons

}
