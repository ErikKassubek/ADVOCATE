// Copyright (c) 2026 Erik Kassubek
//
// File: record.go
// Brief: Gui for record
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func (self *window) setFuzzing() {
	var content []fyne.CanvasObject
	content = []fyne.CanvasObject{
		widget.NewLabel("TODO: Fuzzing"),
	}

	self.setSettings(content)
}
