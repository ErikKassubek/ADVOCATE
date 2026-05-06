// Copyright (c) 2024 Erik Kassubek, Mario Occhinegro
//
// File: comm.go
// Brief: Function to communicate between runtime and advocate
//
// Author: Erik Kassubek
// Created: 2026-04-20
//
// License: BSD-3-Clause

package comm

func (self *Communication) Run() {
	switch self.process {
	case StaticBlock:
		self.staticBlocking()
	}

}
