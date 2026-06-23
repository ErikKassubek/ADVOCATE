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
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type traceConnection struct {
	a string
	b string
}

type componentTraceViewer struct {
	*container.Scroll

	rows, cols int
	cells      map[string]fyne.CanvasObject
	grid       *fyne.Container

	lineLayer *traceViewerCanvas
	links     []traceConnection
}

func createTraceViewer(cols, rows int) *componentTraceViewer {
	grid := container.NewGridWithColumns(cols)

	lineLayer := &traceViewerCanvas{
		lines: make([]*canvas.Line, 0),
	}

	overlay := container.NewStack(
		lineLayer,
		grid,
	)

	scroll := container.NewScroll(overlay)
	scroll.SetMinSize(fyne.NewSize(0, 0))

	return &componentTraceViewer{
		Scroll:    scroll,
		rows:      rows,
		cols:      cols,
		cells:     make(map[string]fyne.CanvasObject),
		grid:      grid,
		lineLayer: lineLayer,
	}
}

func key(rout, row int) string {
	return fmt.Sprintf("%d:%d", rout, row)
}

// AddCell places content into a specific row/col
func (this *componentTraceViewer) AddCell(rout, row int, content string) {
	this.cells[key(rout, row)] = this.buildCell(content)
}

// TODO: color/fix text
func (this *componentTraceViewer) buildCell(text string) fyne.CanvasObject {
	return container.NewPadded(createBoxedLabel(text))
}

func (this *componentTraceViewer) rebuild() {
	this.grid.Objects = nil

	for r := 0; r < this.rows+1; r++ {
		for c := 1; c < this.cols+1; c++ {
			k := key(c, r)

			if r == 0 {
				this.grid.Add(this.buildCell(fmt.Sprint("Routine ", c)))
			} else {
				if obj, ok := this.cells[k]; ok {
					this.grid.Add(obj)
					println("PLOT: ", k)
				} else {
					// empty placeholder so grid stays aligned
					this.grid.Add(this.buildCell(""))
				}
			}
		}
	}

	this.grid.Refresh()
	this.refreshLines()
}

func (this *componentTraceViewer) clear() {
	this.cells = make(map[string]fyne.CanvasObject)
}

func (v *componentTraceViewer) connect(a, b string) {
	v.links = append(v.links, traceConnection{
		a: a,
		b: b,
	})
}

func (v *componentTraceViewer) refreshLines() {
	v.lineLayer.lines = nil

	for _, c := range v.links {
		cellA := v.cells[c.a]
		cellB := v.cells[c.b]

		if cellA == nil || cellB == nil {
			continue
		}

		pa := cellA.Position()
		pb := cellB.Position()

		line := canvas.NewLine(color.White)
		line.StrokeWidth = 2

		line.Position1 = fyne.NewPos(
			pa.X+cellA.Size().Width/2,
			pa.Y+cellA.Size().Height,
		)

		line.Position2 = fyne.NewPos(
			pb.X+cellB.Size().Width/2,
			pb.Y,
		)

		v.lineLayer.lines = append(v.lineLayer.lines, line)
	}

	v.lineLayer.Refresh()
}
