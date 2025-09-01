// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the cssts for fork
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

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
