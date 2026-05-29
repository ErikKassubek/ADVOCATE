// Copyright (c) 2024 Erik Kassubek
//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package main

import (
	"advocate/app"
)

// Main function
func main() {
	mode := app.CommandLine()
	if mode == "" {
		return
	}

	app.Run(mode)
}
