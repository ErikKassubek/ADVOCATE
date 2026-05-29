// Copyright (c) 2026 Erik Kassubek
//
// File: modeSelect.go
// Brief: Create mode selector
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/log"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentOutput struct {
	*fyne.Container

	outputList    *widget.List
	appendOutput  func(string, log.InfoLevel)
	outputChannel chan log.GuiInfo
}

func createOutput() componentOutput {
	co := componentOutput{}

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
	co.outputList = widget.NewList(
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

	co.appendOutput = func(text string, mode log.InfoLevel) {
		timestamp := time.Now().Format("15:04:05")
		line := fmt.Sprintf("[%s] %s", timestamp, text)

		lines = append(lines, logLine{
			text: line,
			mode: mode,
		})

		co.outputList.Refresh()

		co.outputList.ScrollToBottom()
	}

	// self.output.SetMinSize(fyne.NewSize(700, 400))

	co.outputChannel = log.GetGuiChan()

	co.Container = container.NewBorder(nil, nil, nil, nil,
		container.NewVScroll(co.outputList),
	)

	co.appendOutput("Application started", log.GuiLv)

	go func() {
		for c := range co.outputChannel {
			fyne.Do(func() {
				co.appendOutput(c.Msg, c.Lv)
			})
		}
	}()

	return co
}
