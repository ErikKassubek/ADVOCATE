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

func (self *window) setAnalysis() {
	var content []fyne.CanvasObject
	content = []fyne.CanvasObject{
		self.settings.components.mainTestSelect.Container,
		widget.NewSeparator(),
		self.settings.components.label.Container,
		self.settings.components.scen.Container,
		self.settings.components.toRecord.Container,
		self.settings.components.toReplay.Container,
		self.settings.components.maxNumberElements.Container,
		twoCheck(self.settings.components.measureTime.Container, self.settings.components.createStatistics.Container),
		twoCheck(self.settings.components.checkForNotExecuted.Container, self.settings.components.ignoreCriticalSections.Container),
		twoCheck(self.settings.components.ignoreAtomics.Container, self.settings.components.onlyAPanicAndLeak.Container),
		twoCheck(self.settings.components.noRewrite.Container, self.settings.components.deleteTrace.Container),
		twoCheck(self.settings.components.cont.Container, self.settings.components.noWarning.Container),
		twoCheck(self.settings.components.verbose.Container, self.settings.components.noProgress.Container),
		twoCheck(self.settings.components.output.Container, self.settings.components.alwaysPanic.Container),
		self.settings.components.noMemorySup.Container,
	}

	self.setSettings(content)
}
