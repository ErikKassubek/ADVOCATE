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

func (this *window) setReplay() {
	objects := []fyne.CanvasObject{
		this.settings.components.mainTestSelect.Container,
		widget.NewSeparator(),
		this.settings.components.tracePath.Container,
		widget.NewSeparator(),
		this.settings.components.label.Container,
		this.settings.components.toReplay.Container,
		twoCheck(this.settings.components.measureTime.Container, this.settings.components.createStatistics.Container),
		twoCheck(this.settings.components.ignoreAtomics.Container, this.settings.components.cont.Container),
	}
	this.setSettings(objects)
}
