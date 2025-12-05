// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the pog for wait group
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/baseA"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBWait update the pog for a wait group operation
// Parameter:
//   - wa *trace.TraceElementWait: the wait group operation
func UpdateHBWait(wa *trace.ElementWait) {
	switch wa.GetOpW() {
	case trace.WaitAdd, trace.WaitDone:
		Change(wa)
	case trace.WaitWait:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}

// Change updates the pog for an add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.ElementWait) {
	id := wa.GetObjId()

	lw := baseA.LastChangeWG[id]
	if lw != nil {
		AddEdge(lw, wa, false)
	}
	baseA.LastChangeWG[id] = wa
}

// Wait updates the pog for a wait operation
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.ElementWait) {
	id := wa.GetObjId()

	if wa.GetTPost() != 0 {
		lc := baseA.LastChangeWG[id]
		if lc != nil {
			AddEdge(lc, wa, false)
		}
	}
}
