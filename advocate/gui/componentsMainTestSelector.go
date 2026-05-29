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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentMainTestSelect struct {
	*fyne.Container

	mainTest *widget.Select
	allOne   *widget.Select
	testName *widget.Entry

	showAllOne   bool
	showTestName bool
}

func creatMainTestSelector(w *window) componentMainTestSelect {
	csmt := componentMainTestSelect{}

	csmt.mainTest = widget.NewSelect(
		[]string{
			"Unit Tests",
			"Main",
		},
		func(value string) {
			if value == "Main" {
				flags.ModeMain = true
				csmt.showAllOne = false
				csmt.showTestName = false
			} else {
				flags.ModeMain = false
				csmt.showAllOne = true
			}

			csmt.creatMainTestSelectorContainer()
		},
	)

	csmt.allOne = widget.NewSelect(
		[]string{
			"All Tests",
			"One Test",
		},
		func(value string) {
			if value == "All Tests" {
				flags.ExecName = ""
				csmt.showTestName = false
			} else {
				csmt.showTestName = true
			}

			csmt.creatMainTestSelectorContainer()
		},
	)

	csmt.testName = widget.NewEntry()
	csmt.testName.SetPlaceHolder("Test Name")
	// TODO: add search
	csmt.testName.OnChanged = func(text string) {
		flags.ExecName = text
		w.appendOutput(text, log.GuiLv)
	}

	// create container ONCE
	csmt.Container = container.NewVBox()

	csmt.creatMainTestSelectorContainer()

	csmt.mainTest.SetSelected("Unit Tests")
	csmt.allOne.SetSelected("All Tests")

	return csmt
}

func (self *componentMainTestSelect) creatMainTestSelectorContainer() {
	objects := []fyne.CanvasObject{
		widget.NewLabel("Main/Test:"),
		self.mainTest,
	}

	if self.showAllOne {
		objects = append(objects, self.allOne)
	}

	if self.showTestName {
		objects = append(objects, self.testName)
	}

	self.Container.Objects = objects
	self.Container.Refresh()
}
