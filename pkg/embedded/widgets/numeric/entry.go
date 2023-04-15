/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package numeric

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Entry struct {
	*widget.Entry
	size fyne.Size
}

func NewEntry(placeholder string, w *widget.Entry) *Entry {
	e := Entry{
		Entry: w,
	}
	e.SetPlaceHolder(placeholder)
	e.ExtendBaseWidget(&e)

	e.size = e.BaseWidget.MinSize()
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

func (e *Entry) SetMinSize(size fyne.Size) {
	e.size = size
}

func (e *Entry) MinSize() fyne.Size {
	e.ExtendBaseWidget(e)

	min := e.BaseWidget.MinSize()

	if min.Width < e.size.Width {
		min.Width = e.size.Width
	}

	if min.Height < e.size.Height {
		min.Height = e.size.Height
	}

	if e.ActionItem != nil {
		min = min.Add(fyne.NewSize(theme.IconInlineSize()+theme.Padding(), 0))
	}
	if e.Validator != nil {
		min = min.Add(fyne.NewSize(theme.IconInlineSize()+theme.Padding(), 0))
	}

	return min
}
