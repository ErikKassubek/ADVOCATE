// Copyright (c) 2024 Erik Kassubek
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
	"advocate/run"
	"advocate/utils/flags"
	"advocate/utils/log"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

	outputList    *widget.List
	output        *container.Scroll
	appendOutput  func(string, log.InfoLevel)
	outputChannel chan log.GuiInfo

	selectWidget *widget.Select

	selectedProjLabel *widget.Label
	openProjButton    *widget.Button

	runButton      *widget.Button
	settingsButton *widget.Button
}

func (self *window) build() {
	self.a = app.New()
	self.w = self.a.NewWindow("Advocate")

	// self.w.Resize(fyne.NewSize(800, 500))
	self.w.Resize(fyne.NewSize(1920, 1080))
	self.w.CenterOnScreen()

	self.createOutput()
	self.creatModeSelect()
	self.createProjSelector()
	self.createButtons()

	topControls := container.NewVBox(
		self.selectWidget,

		widget.NewSeparator(),

		self.openProjButton,
		self.selectedProjLabel,
	)

	botControls := container.New(
		layout.NewGridLayout(2),
		self.settingsButton,
		self.runButton,
	)

	content := container.NewBorder(
		topControls,
		botControls,
		nil,
		nil,
		self.output,
	)

	self.w.SetContent(content)

	self.handleClose()

}

func (self *window) showAndRun() {
	self.w.ShowAndRun()
}

func (self *window) handleClose() {
	self.w.SetCloseIntercept(func() {
		self.appendOutput("Application shutting down...", log.GuiLv)
		self.w.Close()
		os.Exit(0)
	})
}

func (self *window) createOutput() {
	// internal storage for log lines
	type logLine struct {
		text string
		mode log.InfoLevel
	}

	var lines []logLine

	// color mapping
	getColor := func(mode log.InfoLevel) color.Color {
		switch mode {
		case log.ImportantLv, log.DebugLv:
			return yellow
		case log.ResultLv:
			return green
		case log.ProgressLv:
			return blue
		case log.TimeoutLv:
			return purple
		case log.ErrorLv:
			return red
		default:
			return color.White
		}
	}

	// create list
	self.outputList = widget.NewList(
		func() int { return len(lines) },
		func() fyne.CanvasObject {
			return container.NewPadded(canvas.NewText("", color.White))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			txt := o.(*fyne.Container).Objects[0].(*canvas.Text)

			txt.Text = lines[i].text
			txt.Color = getColor(lines[i].mode)
			txt.TextSize = 14
		},
	)

	self.appendOutput = func(text string, mode log.InfoLevel) {
		timestamp := time.Now().Format("15:04:05")
		line := fmt.Sprintf("[%s] %s", timestamp, text)

		lines = append(lines, logLine{
			text: line,
			mode: mode,
		})

		self.outputList.Refresh()

		// auto-scroll to bottom
		self.outputList.ScrollToBottom()
	}

	self.output = container.NewVScroll(self.outputList)
	self.output.SetMinSize(fyne.NewSize(700, 400))

	self.outputChannel = log.GetGuiChan()

	go func() {
		for c := range self.outputChannel {
			fyne.Do(func() {
				self.appendOutput(c.Msg, c.Lv)
			})
		}
	}()

	self.appendOutput("Application started", log.GuiLv)
}

func (self *window) creatModeSelect() {
	self.selectWidget = widget.NewSelect(
		[]string{
			"Record",
			"Replay",
			"Analysis",
			"Fuzzing",
		},
		func(value string) {
			self.appendOutput("Selected: "+value, log.GuiLv)
			flags.Mode = strings.ToLower(value)
		},
	)

	self.selectWidget.SetSelected("Record")
	flags.Mode = "record"

}

func (self *window) createProjSelector() {
	self.selectedProjLabel = widget.NewLabel("No file selected")

	self.openProjButton = widget.NewButtonWithIcon(
		"Choose Project",
		theme.FolderOpenIcon(),
		func() {
			fileDialog := dialog.NewFolderOpen(
				func(uri fyne.ListableURI, err error) {
					if err != nil {
						self.appendOutput("Error opening folder dialog", log.ErrorLv)
						return
					}

					if uri == nil {
						self.appendOutput("Folder selection canceled", log.GuiLv)
						return
					}

					path := uri.Path()
					self.selectedProjLabel.SetText(filepath.Base(path))

					self.appendOutput("Selected folder: "+path, log.GuiLv)
					flags.ProgPath = path
				},
				self.w,
			)

			fileDialog.Show()
		},
	)
}

func (self *window) createButtons() {
	self.runButton = widget.NewButton("Run", func() {
		self.runButton.Disable()
		err := run.Run()
		if err != nil {
			self.appendOutput(err.Error(), log.ErrorLv)
		}
		self.runButton.Enable()
	})

	self.settingsButton = widget.NewButton("Settings", func() {
		self.appendOutput("Settings", log.GuiLv)
		self.settings()
	})
}
