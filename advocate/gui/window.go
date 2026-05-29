// Copyright (c) 2026 Erik Kassubek
//
// File: gui.go
// Brief: Create main window
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/log"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	red    = color.RGBA{255, 0, 0, 255}
	green  = color.RGBA{0, 255, 0, 255}
	blue   = color.RGBA{0, 0, 255, 255}
	purple = color.RGBA{128, 0, 128, 255}
	yellow = color.RGBA{255, 255, 0, 255}
)

type window struct {
	a fyne.App
	w fyne.Window

	left     *fyne.Container
	settings *fyne.Container

	modeSelect     componentModeSelect
	projSelector   componentProjectSelector
	mainTestSelect componentMainTestSelect
	runButton      componentRunButton
	output         componentOutput
}

func (self *window) create() {
	self.a = app.New()
	self.w = self.a.NewWindow("Advocate")

	// self.w.Resize(fyne.NewSize(800, 500))
	self.w.Resize(fyne.NewSize(1920, 1080))
	self.w.CenterOnScreen()

	self.handleClose()

	self.createComponents()
}

func (self *window) appendOutput(msg string, lv log.InfoLevel) {
	self.output.appendOutput(msg, lv)
}

func (self *window) build() {
	self.left = container.NewBorder(
		container.NewVBox(
			self.modeSelect.Container,

			widget.NewSeparator(),

			self.projSelector.Container,

			widget.NewSeparator(),

			self.mainTestSelect.Container,
		),
		self.runButton.Container,
		nil,
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			self.settings,
		),
	)

	content := container.NewHSplit(self.left, self.output.Container)
	content.SetOffset(0.33)

	self.w.SetContent(content)
}

func (self *window) createComponents() {
	self.projSelector = createProjSelector(self)
	self.mainTestSelect = creatMainTestSelector(self)
	self.runButton = createRunButton(self)
	self.output = createOutput()

	self.settings = container.NewVBox()

	self.modeSelect = creatModeSelect(self) // must be last create

}

func (self *window) showAndRun() {
	self.w.ShowAndRun()
}

func (self *window) handleClose() {
	self.w.SetCloseIntercept(func() {
		self.output.appendOutput("Application shutting down...", log.GuiLv)
		self.w.Close()
		os.Exit(0)
	})
}
