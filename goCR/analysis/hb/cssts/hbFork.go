//
// File: hbAtomic.go
// Brief: Update the cssts for fork
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

import (
	"goCR/trace"
)

// UpdateHBFork update and calculate happens before information for fork operations
//
// Parameter:
//   - fo *TraceElementFork: the fork element
func UpdateHBFork(fo *trace.ElementFork) {
	AddEdgeFork(fo)
}
