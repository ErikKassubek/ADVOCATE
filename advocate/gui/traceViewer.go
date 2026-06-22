// Copyright (c) 2026 Erik Kassubek
//
// File: traceViewer.go
// Brief: Create trace viewer
//
// Author: Erik Kassubek
// Created: 2026-06-22
//
// License: BSD-3-Clause

package gui

import (
	"advocate/io"
	"advocate/trace"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type traceViewer struct {
	window fyne.Window

	pathSelector *componentPathSelector
	closeButton  *componentButton
	traceViewer  *componentTraceViewer

	path  string
	trace *trace.Trace
}

func (self *window) openTraceViewer() {
	tv := traceViewer{}

	tv.window = self.app.NewWindow("Trace")

	tv.pathSelector = createPathSelector("Trace", &tv.path, tv.updateTrace, tv.window)
	tv.closeButton = createButton("Close", tv.window.Close)

	tv.traceViewer = createTraceViewer(0, 0)

	content := container.NewBorder(
		tv.pathSelector.Container,
		tv.closeButton.Container,
		nil,
		nil,
		container.NewStack(tv.traceViewer.Scroll),
	)

	tv.window.SetContent(content)

	tv.window.Resize(fyne.NewSize(1728, 972))
	tv.window.Show()
}

func (self *traceViewer) updateTrace(dir string) {
	err := self.readTrace(dir)
	if err != nil {
		win.writeErr("Could not read trace: ", err)
		return
	}

	self.trace.AsRequestCommit()
	self.trace.NormalizeRequestCommit()

	self.clear()

	traceIter := self.trace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		elem_name := "C:"
		if elem.IsRequest() {
			elem_name = "R:"
		}
		elem_name += elem.ToString()

		// println(elem.GetRoutine(), elem.GetTSort()+1, elem_name)
		self.AddEntry(elem.GetRoutine(), elem.GetTSort()+1, elem_name)
	}

	self.rebuild()

}

func (self *traceViewer) AddEntry(rout, elem int, text string) {
	self.traceViewer.AddCell(rout, elem, text)
}

func (self *traceViewer) clear() {
	self.traceViewer.clear()
}

func (self *traceViewer) rebuild() {
	self.traceViewer.rebuild()
}

func (self *traceViewer) readTrace(dir string) error {
	var err error
	self.traceViewer.cols, self.traceViewer.rows, self.trace, err = io.CreateTraceFromFiles(dir)
	return err
}
