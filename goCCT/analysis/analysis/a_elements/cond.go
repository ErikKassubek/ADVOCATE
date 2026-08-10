// Copyright (c) 2024 Erik Kassubek
//
// File: hbCond.go
// Brief: Update functions for happens before info for conditional variables operations
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_elements

import (
	"gocct/analysis/a_base"
	"gocct/analysis/hb/a_hbcalc"
	"gocct/trace"
)

// AnalyzeCond calculates the hb info for cond events and updates relevant
// analysis data
//
// Parameter:
//   - co *trace.ElementCond: the element
func AnalyzeCond(co *trace.ElementCond) {
	a_hbcalc.UpdateHBCond(co)

	// update currently waiting elements
	id := co.ObjID()
	if co.Committed() { // not leak
		switch co.Type(true) {
		case trace.CondWait:
			if _, ok := a_base.CurrentlyWaiting[id]; !ok {
				a_base.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
			}
			a_base.CurrentlyWaiting[id] = append(a_base.CurrentlyWaiting[id], co)
		case trace.CondSignal:
			if len(a_base.CurrentlyWaiting[id]) != 0 {
				a_base.CurrentlyWaiting[id] = a_base.CurrentlyWaiting[id][1:]
			}
		case trace.CondBroadcast:
			a_base.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
		}

	}
}
