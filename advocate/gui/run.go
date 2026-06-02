// Copyright (c) 2026 Erik Kassubek
//
// File: run.go
// Brief: Run advocate
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

import (
	"advocate/run"
	"context"

	"fyne.io/fyne/v2"
)

type worker struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newWorker() *worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &worker{ctx, cancel}

}

func (self *window) startRunMode() {
	self.settings.disable()
	win.modeSelect.disable()
	win.projSelector.disable()
	win.runButton.disable()
	// win.cancelButton.enable() // uncomment when cxt cancel is implemented
}

func (self *window) endRunMode() {
	self.settings.enable()
	win.modeSelect.enable()
	win.projSelector.enable()
	win.runButton.enable()
	win.cancelButton.disable()
}

func (self *window) start() {
	self.worker = newWorker()

	go func() {
		if !validInput() {
			return
		}

		fyne.Do(func() {
			win.startRunMode()
			win.WriteGui("Start Run")
		})

		err := run.Run()
		if err != nil {
			fyne.Do(func() { win.writeErr(err.Error()) })
		}

		fyne.Do(func() {
			win.endRunMode()
			win.WriteGui("Finish Run")
		})
	}()
}

func (self *window) cancel() {
	self.writeErr("Cancel Run...")
	self.worker.cancel()
}
