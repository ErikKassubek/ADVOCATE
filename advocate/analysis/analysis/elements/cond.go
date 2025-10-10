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
	"advocate/analysis/data"
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
			if _, ok := data.CurrentlyWaiting[id]; !ok {
				data.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
			}
			data.CurrentlyWaiting[id] = append(data.CurrentlyWaiting[id], co)
		case trace.CondSignal:
			if len(data.CurrentlyWaiting[id]) != 0 {
				data.CurrentlyWaiting[id] = data.CurrentlyWaiting[id][1:]
			}
		case trace.CondBroadcast:
			data.CurrentlyWaiting[id] = make([]*trace.ElementCond, 0)
		}

	}
}
