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
		widget.NewLabel("TODO: Replay"),
		// self.settings.components.toRecord.Container,
		// self.settings.components.maxNumberElements.Container,
		// twoCheck(self.settings.components.measureTime.Container, self.settings.components.createStatistics.Container),
		// twoCheck(self.settings.components.checkForNotExecuted.Container, self.settings.components.ignoreAtomics.Container),
		// twoCheck(self.settings.components.deleteTrace.Container, self.settings.components.cont.Container),
		// twoCheck(self.settings.components.noWarning.Container, self.settings.components.noInfo.Container),
		// twoCheck(self.settings.components.noProgress.Container, self.settings.components.output.Container),
		// twoCheck(self.settings.components.alwaysPanic.Container, self.settings.components.noMemorySup.Container),

	}
	self.setSettings(objects)
}
