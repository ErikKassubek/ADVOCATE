// Copyright (c) 2024 Erik Kassubek
//
// File: main.go
// Brief: Main file and starting point for the toolchain
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package main

import (
	"advocate/advoc"
	"advocate/static/s_blocking"
	"advocate/utils/flags"
	"advocate/utils/log"
)

var (
	help bool
)

// Main function
func main() {

	cont := advoc.CommandLine()
	if !cont {
		return
	}

	err := advoc.Run()
	if err != nil {
		log.Error(err)
	}
}
