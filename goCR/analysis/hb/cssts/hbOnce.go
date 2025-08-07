//
// File: hbAtomic.go
// Brief: Update the once for fork
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

import (
	"goCR/analysis/data"
	"goCR/trace"
)

// UpdateHBOnce update the vector clock of the trace and element
// Parameter:
//   - on *trace.TraceElementOnce: the once trace element
func UpdateHBOnce(on *trace.ElementOnce) {
	// suc once does not create edge -> only not suc
	if !on.GetSuc() {
		suc := data.OSuc[on.GetID()]
		AddEdge(suc, on, false)
	}
}
