// Copyright (c) 2024 Erik Kassubek
//
// File: hbCond.go
// Brief: Update functions for happens before info for conditional variables operations
//
// Author: Erik Kassubek
// Created: 2024-01-09
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb/hbcalc"
	"advocate/trace"
)

// AnalyzeCond calculates the hb info for cond events and updates relevant
// analysis data
//
// Parameter:
//   - co *trace.ElementCond: the element
func AnalyzeCond(co *trace.ElementCond) {
	hbcalc.UpdateHBCond(co)

	// update currently waiting elements
	id := co.GetID()
	if co.GetTPost() != 0 { // not leak
		switch co.GetType(true) {
		case trace.CondWait:
			if _, ok := baseA.CurrentlyWaiting[id]; !ok {
				baseA.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
			}
			baseA.CurrentlyWaiting[id] = append(baseA.CurrentlyWaiting[id], co)
		case trace.CondSignal:
			if len(baseA.CurrentlyWaiting[id]) != 0 {
				baseA.CurrentlyWaiting[id] = baseA.CurrentlyWaiting[id][1:]
			}
		case trace.CondBroadcast:
			baseA.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
		}

	}
}
