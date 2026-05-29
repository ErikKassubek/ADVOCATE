// Copyright (c) 2026 Erik Kassubek
//
// File: helper.go
// Brief: Helper functions
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createInput(labelStr, valStr string, canBeDisabled bool) (*fyne.Container, *widget.Entry, *widget.Check) {
	label := widget.NewLabel(labelStr)

	entry := widget.NewEntry()
	entry.SetText(valStr)

	var check *widget.Check
	var middle fyne.CanvasObject

	if canBeDisabled {
		check = widget.NewCheck("", func(b bool) {
			if b {
				entry.Enable()
			} else {
				entry.Disable()
			}
		})
		middle = container.NewGridWrap(fyne.NewSize(40, 40), check)
		if valStr != "-1" {
			check.SetChecked(true)
		} else {
			entry.Disable()
		}
	} else {
		// empty spacer same size as checkbox
		middle = container.NewGridWrap(fyne.NewSize(40, 40), widget.NewLabel(""))
	}

	row := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(200, 40), label),
		middle,
		container.NewGridWrap(fyne.NewSize(200, 40), entry),
	)

	return row, entry, check
}
