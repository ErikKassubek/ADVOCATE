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
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type componentPathSelector struct {
	*fyne.Container

	label *componentSectionLabel

	selectPath        *fyne.Container
	selectedPathLabel *widget.Label
	openPathSelButton *widget.Button

	path string
}

func createPathSelector(label string, valToSet *string, onChange func(path string), parent fyne.Window) *componentPathSelector {
	cps := &componentPathSelector{}

	cps.selectedPathLabel = widget.NewLabel(fmt.Sprintf("No %s selected", strings.ToLower(label)))

	cps.openPathSelButton = widget.NewButtonWithIcon(
		"Select",
		theme.FolderOpenIcon(),
		func() {
			fileDialog := dialog.NewFolderOpen(
				func(uri fyne.ListableURI, err error) {
					if err != nil {
						win.writeErr("Error opening folder dialog")
						return
					}

					if uri == nil {
						return
					}

					path := uri.Path()
					cps.selectedPathLabel.SetText(filepath.Base(path))

					cps.path = path
					if valToSet != nil {
						*valToSet = path
					}
					if onChange != nil {
						onChange(path)
					}
				},
				parent,
			)

			fileDialog.Show()
		},
	)

	cps.label = createSectionLabel(label)

	cps.Container = container.NewVBox(
		cps.label.Container,
		cps.openPathSelButton,
		cps.selectedPathLabel,
	)

	return cps
}

func (self *componentPathSelector) getPath() string {
	return self.path
}

func (self *componentPathSelector) disable() {
	self.openPathSelButton.Disable()
}

func (self *componentPathSelector) enable() {
	self.openPathSelButton.Enable()
}
