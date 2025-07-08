// Copyright (c) 2024 Erik Kassubek
//
// File: vcWait.go
// Brief: Update functions for happens before info for wait group operations
//        Some function start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/concurrent/cssts"
	"advocate/analysis/concurrent/pog"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// UpdateHBWait updates and stores the vector clock of the element
// Parameter:
//   - wa *TraceElementWait: the wait trace element
func UpdateHBWait(wa *trace.ElementWait) {
	routine := wa.GetRoutine()
	wa.SetVc(data.CurrentVC[routine])
	wa.SetWVc(data.CurrentWVC[routine])

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

// Change calculate the new vector clock for a add or done operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Change(wa *trace.ElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := wa.GetID()
	routine := wa.GetRoutine()

	lw := data.LastChangeWG[id]
	if lw != nil {
		wa.GetVC().Sync(lw.GetVC())
	}
	data.LastChangeWG[id] = wa

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["doneBeforeAdd"] {
		checkForDoneBeforeAddChange(wa)
	}
}

// Wait calculates the new vector clock for a wait operation and update cv
//
// Parameter:
//   - wa *TraceElementWait: The trace element
func Wait(wa *trace.ElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := wa.GetID()
	routine := wa.GetRoutine()

	if wa.GetTPost() != 0 {
		lc := data.LastChangeWG[id]
		if lc != nil {
			data.CurrentVC[routine].Sync(lc.GetVC())
			pog.AddEdge(lc, wa, false)
			cssts.AddEdge(lc, wa, false)
		}
	}

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}
