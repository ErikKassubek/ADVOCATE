//
// File: hbAtomic.go
// Brief: Update the cssts for wait group
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

import (
	"goCR/analysis/data"
	"goCR/trace"
	"goCR/utils/log"
)

// UpdateHBWait update the cssts for a wait group operation
// Parameter:
//   - wa *trace.TraceElementWait: the wait group operation
func UpdateHBWait(wa *trace.ElementWait) {
	switch wa.GetOpW() {
	case trace.ChangeOp:
		Change(wa)
	case trace.WaitOp:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}

// Change updates the cssts for an add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.ElementWait) {
	id := wa.GetID()

	lw := data.LastChangeWG[id]
	if lw != nil {
		AddEdge(lw, wa, false)
	}
	data.LastChangeWG[id] = wa
}

// Wait updates the pog for a wait operation
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.ElementWait) {
	id := wa.GetID()

	if wa.GetTPost() != 0 {
		lc := data.LastChangeWG[id]
		if lc != nil {
			AddEdge(lc, wa, false)
		}
	}
}
