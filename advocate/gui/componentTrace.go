// Copyright (c) 2026 Erik Kassubek
//
// File: componentTrace.go
// Brief: Trace Viewer
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type componentTraceViewer struct {
	*container.Scroll

	rows, cols int
	cells      map[string]fyne.CanvasObject
	grid       *fyne.Container
}

func createTraceViewer(rows, cols int) *componentTraceViewer {
	grid := container.NewGridWithColumns(cols)
	scroll := container.NewScroll(grid)
	scroll.SetMinSize(fyne.NewSize(0, 0))

	v := &componentTraceViewer{
		Scroll: scroll,

		rows:  rows,
		cols:  cols,
		cells: make(map[string]fyne.CanvasObject),
		grid:  grid,
	}

	return v
}

func key(r, c int) string {
	return fmt.Sprintf("%d:%d", r, c)
}

// AddCell places content into a specific row/col
func (self *componentTraceViewer) AddCell(row, col int, content string) {
	self.cells[key(row, col)] = self.colorText(content)
	self.rebuild()
}

// TODO: color/fix text
func (self *componentTraceViewer) colorText(text string) *widget.Label {
	return widget.NewLabel(text)
}

func (self *componentTraceViewer) rebuild() {
	self.grid.Objects = nil

	for r := 0; r < self.rows; r++ {
		for c := 0; c < self.cols; c++ {
			k := key(r, c)

			if obj, ok := self.cells[k]; ok {
				self.grid.Add(obj)
			} else {
				// empty placeholder so grid stays aligned
				self.grid.Add(widget.NewLabel(""))
			}
		}
	}

	self.grid.Refresh()
}

func (self *componentTraceViewer) clear() {
	self.cells = make(map[string]fyne.CanvasObject)
}
