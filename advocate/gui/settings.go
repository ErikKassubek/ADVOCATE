// Copyright (c) 2024 Erik Kassubek
//
// File: setting.go
// Brief: Create settings
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/flags"
	"advocate/utils/log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (self *window) settings() {
	second := self.a.NewWindow("Settings")
	second.SetContent(widget.NewLabel("Hello from window 2"))
	second.Resize(fyne.NewSize(800, 300))

	toRecWid, _, _ := createInput("Timeout Recording [s]: ", strconv.Itoa(flags.TimeoutRecording), true)
	toRepWid, _, _ := createInput("Timeout Replay [s]: ", strconv.Itoa(flags.TimeoutReplay), true)
	toFuzWid, _, _ := createInput("Timeout Fuzzing [s]: ", strconv.Itoa(flags.TimeoutFuzzing), true)
	maxFuzRunWid, _, _ := createInput("Max. Fuzzing Runs: ", strconv.Itoa(flags.MaxFuzzingRun), true)
	maxNumElemWid, _, _ := createInput("Max. Number Elements: ", strconv.Itoa(flags.MaxNumberElements), true)

	grid := container.NewGridWithColumns(2,
		toRecWid,
		toRepWid,
		toFuzWid,
		maxFuzRunWid,
		maxNumElemWid,
	)

	save := container.New(
		layout.NewGridLayout(2),
		widget.NewButton("Save", func() {
			self.appendOutput("Save Settings", log.GuiLv)
			// TODO: store value
		}),
		widget.NewButton("Cancel", func() {
			second.Close()
		}),
	)

	content := container.NewBorder(
		nil,
		save,
		nil,
		nil,
		grid,
	)
	second.SetContent(content)

	second.Show()
}
