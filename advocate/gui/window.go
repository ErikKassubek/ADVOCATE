// Copyright (c) 2026 Erik Kassubek
//
// File: window.go
// Brief: Create main window
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/flags"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	red    = color.RGBA{255, 99, 99, 255}
	green  = color.RGBA{80, 220, 120, 255}
	blue   = color.RGBA{100, 170, 255, 255}
	purple = color.RGBA{200, 140, 255, 255}
	yellow = color.RGBA{255, 220, 90, 255}
	pink   = color.RGBA{255, 120, 200, 255}
	gray   = color.RGBA{200, 200, 200, 255}
)

type window struct {
	a fyne.App
	w fyne.Window

	left  *fyne.Container
	right *fyne.Container

	modeSelect   *componentModeSelect
	projSelector *componentPathSelector
	runButton    *componentButton
	cancelButton *componentButton
	output       *componentOutput
	progressBar  *componentProgress
	settings     *componentSetting

	worker *worker
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

func (self *window) build() {
	self.left = container.NewBorder(
		container.NewVBox(
			self.modeSelect.Container,

			widget.NewSeparator(),

			self.projSelector.Container,
		),
		container.NewVBox(
			self.runButton.Container,
			self.cancelButton.Container,
		),
		nil,
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			self.settings.Container,
		),
	)

	self.right = container.NewBorder(
		nil,
		self.progressBar.Container,
		nil,
		nil,
		self.output.Container,
	)

	content := container.NewHSplit(self.left, self.right)
	content.SetOffset(0.33)

	self.w.SetContent(content)
}

func (self *window) createComponents() {
	self.output = createOutput()

	self.projSelector = createPathSelector("Project", &flags.ProgPath)
	self.runButton = createRunButton()
	self.cancelButton = createCancelButton()
	self.progressBar = createProgressBar()
	self.settings = createSettings()

	self.modeSelect = createModeSelect() // must be created last

	self.cancelButton.disable()
}

func (self *window) showAndRun() {
	self.w.ShowAndRun()
}

func (self *window) handleClose() {
	self.w.SetCloseIntercept(func() {
		self.WriteGui("Application shutting down...")
		self.w.Close()
		os.Exit(0)
	})
}
