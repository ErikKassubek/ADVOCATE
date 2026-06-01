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
		self.settings.components.toRecord.Container,
		self.settings.components.maxNumberElements.Container,
		twoCheck(self.settings.components.measureTime.Container, self.settings.components.createStatistics.Container),
		twoCheck(self.settings.components.checkForNotExecuted.Container, self.settings.components.ignoreAtomics.Container),
	}
	self.setSettings(objects)
}
