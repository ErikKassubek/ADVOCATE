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
	"advocate/run"
	"advocate/utils/log"
)

var (
	help bool
)

// Main function
func main() {
	cont := run.CommandLine()
	if !cont {
		return
	}

	err := run.Run()
	if err != nil {
		log.Error(err)
	}
}
