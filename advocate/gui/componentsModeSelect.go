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

	label componentSectionLabel

	modeSelectWidget *widget.Select
}

const (
	record   = "Record"
	replay   = "Replay"
	analysis = "Analysis"
	fuzzing  = "Fuzzing"
)

func createModeSelect() componentModeSelect {
	cms := componentModeSelect{}

	cms.modeSelectWidget = widget.NewSelect(
		[]string{
			record,
			replay,
			analysis,
			fuzzing,
		},
		func(value string) {
			flags.Mode = strings.ToLower(value)

			if value == replay {
				win.settings.components.mainTestSelect.isReplay(true)
			} else {
				win.settings.components.mainTestSelect.isReplay(false)
			}

			switch value {
			case record:
				win.setRecord()
			case analysis:
				win.setAnalysis()
			case replay:
				win.setReplay()
			case fuzzing:
				win.setFuzzing()
			}

		},
	)

	cms.label = createSectionLabel("Mode")

	cms.Container = container.NewVBox(
		cms.label.Container,
		cms.modeSelectWidget,
	)

	cms.modeSelectWidget.SetSelected("Record")

	return cms
}
