// Copyright (c) 2026 Erik Kassubek
//
// File: componentsMainTestSelector.go
// Brief: Create main/test selector
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/flags"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentButton struct {
	*fyne.Container

	runButton *widget.Button
}

func createButton(label string, f func()) *componentButton {
	rb := &componentButton{}

	rb.runButton = widget.NewButton(label, f)

	rb.Container = container.NewVBox(
		rb.runButton,
	)

	return rb
}

func createRunButton() *componentButton {
	return createButton("Run", win.start)
}

func createCancelButton() *componentButton {
	return createButton("Cancel", func() { win.worker.cancel() })
}

func (self *componentButton) disable() {
	self.runButton.Disable()
}

func (self *componentButton) enable() {
	self.runButton.Enable()
}

func validInput() bool {
	if flags.ProgPath == "" {
		win.writeErr("Program path not set")
		win.writeErr("Abort...")
		return false
	}

	return true
}
