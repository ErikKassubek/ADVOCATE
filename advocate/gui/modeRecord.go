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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type settingsRecord struct {
	*fyne.Container
}

func createSettingsRecord(comp *settingComponents) *settingsRecord {
	sr := settingsRecord{}

	sr.Container = container.NewVBox()

	return &sr
}

func (self *window) setRecord() {
	objects := []fyne.CanvasObject{
		self.settings.components.mainTestSelect.Container,
		widget.NewSeparator(),
		self.settings.components.label.Container,
		self.settings.components.label,
		self.settings.components.toRecord.Container,
		self.settings.components.maxNumberElements.Container,
		twoCheck(self.settings.components.measureTime.Container, self.settings.components.createStatistics.Container),
		twoCheck(self.settings.components.checkForNotExecuted.Container, self.settings.components.ignoreAtomics.Container),
		twoCheck(self.settings.components.deleteTrace.Container, self.settings.components.cont.Container),
		twoCheck(self.settings.components.noWarning.Container, self.settings.components.verbose.Container),
		twoCheck(self.settings.components.noProgress.Container, self.settings.components.output.Container),
		twoCheck(self.settings.components.alwaysPanic.Container, self.settings.components.noMemorySup.Container),
	}
	self.setSettings(objects)
}
