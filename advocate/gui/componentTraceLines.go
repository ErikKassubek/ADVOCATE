// Copyright (c) 2026 Erik Kassubek
//
// File: componentTraceLines.go
// Brief: Draw lines between traces
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
)

type traceViewerCanvas struct {
	lines []*canvas.Line
	size  fyne.Size
}

func (t *traceViewerCanvas) MinSize() fyne.Size {
	return t.size
}

func (t *traceViewerCanvas) Resize(size fyne.Size) {
	t.size = size
}

func (t *traceViewerCanvas) Move(pos fyne.Position) {}

func (t *traceViewerCanvas) Position() fyne.Position {
	return fyne.NewPos(0, 0)
}

func (t *traceViewerCanvas) Size() fyne.Size {
	return t.size
}

func (t *traceViewerCanvas) Visible() bool {
	return true
}

func (t *traceViewerCanvas) Show() {}

func (t *traceViewerCanvas) Hide() {}

func (t *traceViewerCanvas) Refresh() {
	for _, l := range t.lines {
		l.Refresh()
	}
}

func (t *traceViewerCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &traceViewerRenderer{
		obj: t,
	}
}

type traceViewerRenderer struct {
	obj *traceViewerCanvas
}

func (r *traceViewerRenderer) Layout(size fyne.Size) {}

func (r *traceViewerRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *traceViewerRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, len(r.obj.lines))

	for i, l := range r.obj.lines {
		objects[i] = l
	}

	return objects
}

func (r *traceViewerRenderer) Refresh() {
	canvas.Refresh(r.obj)
}

func (r *traceViewerRenderer) Destroy() {}

func (r *traceViewerRenderer) BackgroundColor() color.Color {
	return color.Transparent
}
