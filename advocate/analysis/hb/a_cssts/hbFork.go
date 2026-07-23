// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the cssts for fork
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_cssts

import (
	"advocate/trace"
)

// UpdateHBFork update and calculate happens before information for fork operations
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateHBFork(fo *trace.ElementFork) {
	AddEdgeFork(fo)
}
