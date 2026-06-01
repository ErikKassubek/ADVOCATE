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
	"advocate/fuzzing/baseF"
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

func createInputText(labelStr string, valToSet *string, canBeDisabled bool) textInput {
	entry := widget.NewEntry()

	entry.SetText(*valToSet)

	var check *widget.Check
	var middle fyne.CanvasObject

	if canBeDisabled {
		check = widget.NewCheck("", func(b bool) {
			if b {
				*valToSet = entry.Text
				entry.Enable()
			} else {
				*valToSet = ""
				entry.Disable()
			}
		})
		middle = container.NewGridWrap(fyne.NewSize(40, 40), check)
		if *valToSet != "-1" {
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

func createInputTextFunc(labelStr string, def string, onSet func(e bool, s string), canBeDisabled, isDisabled bool) textInput {
	entry := widget.NewEntry()

	entry.SetText("")

	var check *widget.Check
	var middle fyne.CanvasObject

	if canBeDisabled {
		check = widget.NewCheck("", func(b bool) {
			if b {
				entry.Enable()
			} else {
				entry.Disable()
			}
			onSet(b, entry.Text)
		})
		check.SetChecked(!isDisabled)
		middle = container.NewGridWrap(fyne.NewSize(40, 40), check)
	} else {
		// empty spacer same size as checkbox
		middle = container.NewGridWrap(fyne.NewSize(40, 40), widget.NewLabel(""))
	}

	entry.OnChanged = func(s string) {
		onSet(check.Checked, s)
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
				*valToSet = math.ToNum[T](entry.Text)
			} else {
				entry.Disable()
				*valToSet = -1
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
	check.SetChecked(*valToSet)

	row := createSettingField(labelStr, check, nil)

	return checkInput{Container: row, check: check}
}

// -------------------------------------------------------------------------------------------------------
// Select Input
// -------------------------------------------------------------------------------------------------------

type selectInput struct {
	*fyne.Container

	sel   *widget.Select
	check *widget.Check
}

func createInputSelect(labelStr string, valToSet *string, values []string, canBeDisabled bool) selectInput {
	entry := widget.NewSelect(values, func(s string) {
		*valToSet = s
	})

	entry.SetSelected(baseF.Modes[0])

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
	} else {
		// empty spacer same size as checkbox
		middle = container.NewGridWrap(fyne.NewSize(40, 40), widget.NewLabel(""))
	}

	row := createSettingField(labelStr, middle, entry)

	return selectInput{Container: row, sel: entry, check: check}
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
