// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the once for fork
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_cssts

import (
	"gocdr/analysis/a_base"
	"gocdr/trace"
)

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	// suc once does not create edge -> only not suc
	if !on.GetSuc() {
		suc := a_base.OSuc[on.ObjID()]
		AddEdge(suc, on, false)
	}
}
