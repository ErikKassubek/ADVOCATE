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
	"gocct/cct"
	"gocct/utils/log"
)

var (
	help bool
)

// Main function
func main() {

	cont := cct.CommandLine()
	if !cont {
		return
	}

	err := cct.Run()
	if err != nil {
		log.Error(err)
	}
}
