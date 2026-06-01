// Copyright (c) 2026 Erik Kassubek
//
// File: componentInput.go
// Brief: input components
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/utils/math"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// -------------------------------------------------------------------------------------------------------
// Text Input
// -------------------------------------------------------------------------------------------------------

type textInput struct {
	*fyne.Container

	entry *widget.Entry
	check *widget.Check
}

func createInputText(labelStr string, val string, canBeDisabled bool) textInput {
	entry := widget.NewEntry()

	entry.SetText(val)

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
		if val != "-1" {
			check.SetChecked(true)
		} else {
			entry.Disable()
		}
	} else {
		// empty spacer same size as checkbox
		middle = container.NewGridWrap(fyne.NewSize(40, 40), widget.NewLabel(""))
	}

	row := createSettingField(labelStr, middle, entry)

	return textInput{Container: row, entry: entry, check: check}
}

// -------------------------------------------------------------------------------------------------------
// Numeric Input
// -------------------------------------------------------------------------------------------------------

type numericEntry struct {
	widget.Entry
	AllowFloat bool
}

func NewNumericEntry() *numericEntry {
	e := &numericEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *numericEntry) TypedRune(r rune) {
	if unicode.IsDigit(r) {
		e.Entry.TypedRune(r)
		return
	}

	if e.AllowFloat &&
		r == '.' &&
		!strings.Contains(e.Text, ".") {
		e.Entry.TypedRune(r)
	}
}

func createInputNumeric[T math.NumberType](labelStr string, valToSet *T, canBeDisabled bool) textInput {
	var zero T
	entry := NewNumericEntry()

	// Allow decimals for float types.
	switch any(zero).(type) {
	case float32, float64:
		entry.AllowFloat = true
	}

	valStr := math.ToString(*valToSet)
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

		middle = container.NewGridWrap(
			fyne.NewSize(40, 40),
			check,
		)

		if valStr != "-1" {
			check.SetChecked(true)
		} else {
			entry.Disable()
		}
	} else {
		middle = container.NewGridWrap(
			fyne.NewSize(40, 40),
			widget.NewLabel(""),
		)
	}

	row := createSettingField(labelStr, middle, entry)

	check.OnChanged = func(b bool) {
		if b {
			*valToSet = math.ToNum[T](entry.Text)
			entry.Enable()
		} else {
			*valToSet = -1
			entry.Disable()
		}
	}

	entry.OnChanged = func(_ string) {
		if check.Checked {
			*valToSet = math.ToNum[T](entry.Text)
		}
	}

	return textInput{
		Container: row,
		entry:     &entry.Entry,
		check:     check,
	}
}

// -------------------------------------------------------------------------------------------------------
// Check Input
// -------------------------------------------------------------------------------------------------------

type checkInput struct {
	*fyne.Container

	check *widget.Check
}

func createInputCheck(labelStr string, valToSet *bool) checkInput {
	check := widget.NewCheck("", func(b bool) { *valToSet = b })

	row := createSettingField(labelStr, check, nil)

	return checkInput{Container: row, check: check}
}

// -------------------------------------------------------------------------------------------------------
// Helper Input
// -------------------------------------------------------------------------------------------------------

const rowHeight float32 = 40
const labelWidth float32 = 250
const checkWidth float32 = 40
const inputWidth float32 = labelWidth + checkWidth

func createSettingField(labelStr string, middle fyne.CanvasObject, right fyne.CanvasObject) *fyne.Container {
	label := widget.NewLabel(labelStr)

	if right == nil {
		right = widget.NewLabel("")
	}

	return container.NewHBox(
		container.NewGridWrap(
			fyne.NewSize(labelWidth, rowHeight),
			label,
		),
		container.NewGridWrap(
			fyne.NewSize(checkWidth, rowHeight),
			middle,
		),
		container.NewGridWrap(
			fyne.NewSize(inputWidth, rowHeight),
			right,
		),
	)
}

func twoCheck(left, right fyne.CanvasObject) *fyne.Container {
	return container.NewHBox(
		container.NewGridWrap(
			fyne.NewSize(inputWidth, rowHeight),
			left,
		),
		container.NewGridWrap(
			fyne.NewSize(inputWidth, rowHeight),
			right,
		),
	)
}
