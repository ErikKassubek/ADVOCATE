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
	"advocate/utils/log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type componentProjectSelector struct {
	*fyne.Container

	selectProj        *fyne.Container
	selectedProjLabel *widget.Label
	openProjButton    *widget.Button
}

func createProjSelector(win *window) componentProjectSelector {
	cps := componentProjectSelector{}

	cps.selectedProjLabel = widget.NewLabel("No project selected")

	cps.openProjButton = widget.NewButtonWithIcon(
		"Choose Project",
		theme.FolderOpenIcon(),
		func() {
			fileDialog := dialog.NewFolderOpen(
				func(uri fyne.ListableURI, err error) {
					if err != nil {
						win.appendOutput("Error opening folder dialog", log.ErrorLv)
						return
					}

					if uri == nil {
						win.appendOutput("Folder selection canceled", log.GuiLv)
						return
					}

					path := uri.Path()
					cps.selectedProjLabel.SetText(filepath.Base(path))

					win.appendOutput("Selected folder: "+path, log.GuiLv)
					flags.ProgPath = path
				},
				win.w,
			)

			fileDialog.Show()
		},
	)

	cps.Container = container.NewVBox(
		widget.NewLabel("Project:"),
		cps.openProjButton,
		cps.selectedProjLabel,
	)

	return cps
}
