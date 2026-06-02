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

const (
	unitTest = "Unit Test"
	mainFunc = "Main"
	allTests = "All Tests"
	oneTest  = "One Test"
)

type componentMainTestSelect struct {
	*fyne.Container

	label *componentSectionLabel

	mainTestSel *widget.Select
	allOneSel   *widget.Select
	testNameSel *widget.Select

	showAllOne   bool
	showTestName bool

	replay bool

	testNames []string
}

func creatMainTestSelector() *componentMainTestSelect {
	csmt := &componentMainTestSelect{}

	csmt.mainTestSel = widget.NewSelect(
		[]string{
			unitTest,
			mainFunc,
		},
		func(value string) {
			if value == mainFunc {
				flags.ModeMain = true
				csmt.showAllOne = false
				csmt.showTestName = false
			} else {
				flags.ModeMain = false
				csmt.showAllOne = true
				csmt.showTestName = (csmt.allOneSel.Selected == oneTest)
				flags.ExecName = csmt.testNameSel.Selected

			}

			csmt.creatMainTestSelectorContainer()
		},
	)

	csmt.allOneSel = widget.NewSelect(
		[]string{
			allTests,
			oneTest,
		},
		func(s string) {
			if s == allTests {
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

	csmt.Container = container.NewVBox()

	csmt.creatMainTestSelectorContainer()

	csmt.mainTestSel.SetSelected(unitTest)
	csmt.allOneSel.SetSelected(allTests)

	if len(csmt.testNameSel.Options) == 0 {
		csmt.testNameSel.PlaceHolder = "No tests found"
	}

	return csmt
}

func (self *componentMainTestSelect) creatMainTestSelectorContainer() {
	self.label = createSectionLabel("Main/Test")
	objects := []fyne.CanvasObject{
		self.label.Container,
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

	if len(*names) != 0 {
		self.testNameSel.SetSelected(self.testNames[0])
	} else {
		self.testNameSel.ClearSelected()
	}
}

func (self *componentMainTestSelect) isReplay(r bool) {
	self.replay = r
	if r {
		self.allOneSel.SetSelected(oneTest)
		self.allOneSel.Hide()
	} else {
		self.allOneSel.SetSelected(allTests)
		self.allOneSel.Show()
	}
}

func (self *componentMainTestSelect) disable() {
	self.mainTestSel.Disable()
	self.allOneSel.Disable()
	self.testNameSel.Disable()
}

func (self *componentMainTestSelect) enable() {
	self.mainTestSel.Enable()
	self.allOneSel.Enable()
	self.testNameSel.Enable()
}
