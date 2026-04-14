// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the vc for wait group
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package vc

import (
	"advocate/analysis/baseA"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBWait update the vector clocks for a wait group operation
// Parameter:
//   - wa *trace.TraceElementWait: the wait group operation
func UpdateHBWait(wa *trace.ElementWait) {
	routine := wa.GetRoutine()
	wa.SetVc(CurrentVC[routine])
	wa.SetWVc(CurrentWVC[routine])

	switch wa.GetType(true) {
	case trace.WaitAdd, trace.WaitDone:
		Change(wa)
	case trace.WaitWait:
		Wait(wa)
	default:
		err := "Unknown operation on wait group: " + wa.ToString()
		log.Error(err)
	}
}

// Change calculate the new vector clock for a add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.ElementWait) {
	id := wa.GetObjId()
	routine := wa.GetRoutine()

	lw := baseA.LastChangeWG[id]
	if lw != nil {
		wa.GetVC().Sync(lw.GetVC())
	}
	baseA.LastChangeWG[id] = wa

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}

// Wait calculates the new vector clock for a wait operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.ElementWait) {
	id := wa.GetObjId()
	routine := wa.GetRoutine()

	if wa.GetTPost() != 0 {
		lc := baseA.LastChangeWG[id]
		if lc != nil {
			CurrentVC[routine].Sync(lc.GetVC())
		}
	}

	CurrentVC[routine].Inc(routine)
	CurrentWVC[routine].Inc(routine)
}
