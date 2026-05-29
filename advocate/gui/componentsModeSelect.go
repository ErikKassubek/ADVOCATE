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
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentModeSelect struct {
	*fyne.Container

	modeSelectWidget *widget.Select
}

func creatModeSelect(win *window) componentModeSelect {
	cms := componentModeSelect{}

	cms.modeSelectWidget = widget.NewSelect(
		[]string{
			"Record",
			"Replay",
			"Analysis",
			"Fuzzing",
		},
		func(value string) {
			flags.Mode = strings.ToLower(value)

			switch value {
			case "Record":
				win.buildRecord()
			case "Analysis":
				win.buildAnalysis()
			case "Replay":
				win.buildReplay()
			case "Fuzzing":
				win.buildFuzzing()
			}
			win.settings.Refresh()
		},
	)

	cms.Container = container.NewVBox(
		widget.NewLabel("Mode:"),
		cms.modeSelectWidget,
	)

	cms.modeSelectWidget.SetSelected("Record")

	return cms
}
