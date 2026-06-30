// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the once for fork
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package a_cssts

import (
	"advocate/analysis/a_base"
	"advocate/trace"
)

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	// suc once does not create edge -> only not suc
	if !on.GetSuc() {
		suc := a_base.OSuc[on.GetObjId()]
		AddEdge(suc, on, false)
	}
}
