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
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type componentSectionLabel struct {
	*fyne.Container

	label *canvas.Text
}

func createSectionLabel(text string) *componentSectionLabel {
	label := canvas.NewText(
		text,
		theme.Color(theme.ColorNameError),
	)

	return &componentSectionLabel{
		Container: container.NewHBox(label),
		label:     label,
	}
}

func createBoxedLabel(text string) fyne.CanvasObject {
	label := widget.NewLabel(text)

	if text == "" {
		return label
	}

	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = color.White
	border.StrokeWidth = 1

	return container.NewStack(
		border,
		label,
	)
}
