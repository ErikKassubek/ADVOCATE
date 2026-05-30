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

type componentMainTestSelect struct {
	*fyne.Container

	mainTestSel *widget.Select
	allOneSel   *widget.Select
	testNameSel *widget.Select

	showAllOne   bool
	showTestName bool

	testNames []string
}

func creatMainTestSelector(w *window) componentMainTestSelect {
	csmt := componentMainTestSelect{}

	csmt.mainTestSel = widget.NewSelect(
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

	csmt.allOneSel = widget.NewSelect(
		[]string{
			"All Tests",
			"One Test",
		},
		func(s string) {
			if s == "All Tests" {
				flags.ExecName = ""
				csmt.showTestName = false
			} else {
				csmt.showTestName = true
			}

			csmt.creatMainTestSelectorContainer()
		},
	)

	csmt.testNameSel = widget.NewSelect(
		[]string{},
		func(s string) {
			flags.ExecName = s
		},
	)

	// create container ONCE
	csmt.Container = container.NewVBox()

	csmt.creatMainTestSelectorContainer()

	csmt.mainTestSel.SetSelected("Unit Tests")
	csmt.allOneSel.SetSelected("All Tests")

	return csmt
}

func (self *componentMainTestSelect) creatMainTestSelectorContainer() {
	objects := []fyne.CanvasObject{
		widget.NewLabel("Main/Test:"),
		self.mainTestSel,
	}

	if self.showAllOne {
		objects = append(objects, self.allOneSel)
	}

	if self.showTestName {
		objects = append(objects, self.testNameSel)
	}

	self.Container.Objects = objects
	self.Container.Refresh()
}

func (self *componentMainTestSelect) setTestNames(names *[]string) {
	self.testNames = *names
	self.testNameSel.Options = self.testNames

	self.testNameSel.SetSelected(self.testNames[0])
}
