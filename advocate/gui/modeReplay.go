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

func (self *window) setReplay() {
	objects := []fyne.CanvasObject{
		self.settings.components.mainTestSelect.Container,
		widget.NewSeparator(),
		self.settings.components.tracePath.Container,
		widget.NewSeparator(),
		self.settings.components.label.Container,
		self.settings.components.toReplay.Container,
		twoCheck(self.settings.components.measureTime.Container, self.settings.components.createStatistics.Container),
		twoCheck(self.settings.components.ignoreAtomics.Container, self.settings.components.cont.Container),
	}
	self.setSettings(objects)
}
