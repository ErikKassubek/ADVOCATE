// Copyright (c) 2026 Erik Kassubek
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
	"advocate/gui"
	"advocate/run"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/paths"
	"advocate/utils/timer"
	"context"
)

// Main function
func main() {
	cont := run.CommandLine()
	if !cont {
		return
	}

	// gui.Run()

	initialize()

	if flags.Mode == "gui" {
		gui.Run()
	} else {
		err := run.Run(context.Background())
		if err != nil {
			log.Error(err)
		}
	}
}

func initialize() {
	progPathDir := paths.GetDirectory(flags.ProgPath)
	timer.Init(progPathDir)
}
