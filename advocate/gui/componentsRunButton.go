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
	"advocate/run"
	"advocate/utils/log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentRunButton struct {
	*fyne.Container

	runButton *widget.Button
}

func createRunButton(win *window) componentRunButton {
	rb := componentRunButton{}

	rb.runButton = widget.NewButton("Run", func() {
		rb.runButton.Disable()
		err := run.Run()
		if err != nil {
			win.appendOutput(err.Error(), log.ErrorLv)
		}
		rb.runButton.Enable()
	})

	rb.Container = container.NewVBox(
		rb.runButton,
	)

	return rb
}
