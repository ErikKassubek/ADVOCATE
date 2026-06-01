// Copyright (c) 2026 Erik Kassubek
//
// File: gui.go
// Brief: Entry point for graphical interface
//
// Author: Erik Kassubek
// Created: 2026-05-29
//
// License: BSD-3-Clause

package gui

var win = window{}

func Run() {
	win.create()
	win.build()
	win.showAndRun()
}
