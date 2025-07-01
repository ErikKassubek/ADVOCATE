// Copyright (c) 2025 Erik Kassubek
//
// File: helper.go
// Brief: Utils for memory
//
// Author: Erik Kassubek
// Created: 2025-03-11
//
// License: BSD-3-Clause

package memory

import (
	"advocate/utils/log"
	"runtime"
)

// printAllGoroutines prints the stack traces of all routines
func printAllGoroutines() {
	buf := make([]byte, 1<<20) // 1 MB buffer
	n := runtime.Stack(buf, true)
	log.Infof("%s", buf[:n])
}
