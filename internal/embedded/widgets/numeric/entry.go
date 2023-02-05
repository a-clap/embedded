/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package numeric

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Entry struct {
	*widget.Entry
}

func NewEntry(placeholder string, w *widget.Entry) *Entry {
	e := Entry{
		Entry: w,
	}
	e.SetPlaceHolder(placeholder)
	e.ExtendBaseWidget(e)

	return &e
}

func (e *Entry) Tapped(*fyne.PointEvent) {
	if e.Disabled() {
		return
	}

	w, _ := Show(e)
	w.Show()
}

func (e *Entry) Get() string {
	return e.Text
}

func (e *Entry) Set(value string) {
	e.SetText(value)
}
