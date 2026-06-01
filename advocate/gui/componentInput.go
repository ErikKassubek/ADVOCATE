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

type textInput struct {
	*fyne.Container

	entry *widget.Entry
	check *widget.Check
}
type NumericEntry struct {
	widget.Entry
	AllowFloat bool
}

func NewNumericEntry() *NumericEntry {
	e := &NumericEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *NumericEntry) TypedRune(r rune) {
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

func createNumericInput[T math.NumberType](labelStr string, valToSet *T, canBeDisabled bool) textInput {
	label := widget.NewLabel(labelStr)

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

	row := container.NewHBox(
		container.NewGridWrap(
			fyne.NewSize(200, 40),
			label,
		),
		middle,
		container.NewGridWrap(
			fyne.NewSize(200, 40),
			entry,
		),
	)

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

func createTextInput(labelStr string, val string, canBeDisabled bool) textInput {

	label := widget.NewLabel(labelStr)
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

	row := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(200, 40), label),
		middle,
		container.NewGridWrap(fyne.NewSize(200, 40), entry),
	)

	return textInput{Container: row, entry: entry, check: check}
}
