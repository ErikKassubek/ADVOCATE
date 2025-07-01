// Copyright (c) 2024 Erik Kassubek
//
// File: vcCond.go
// Brief: Update functions for vector clocks from conditional variables operations
//
// Author: Erik Kassubek
// Created: 2024-01-09
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateVCCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateVCCond(co *trace.TraceElementCond) {
	routine := co.GetRoutine()
	co.SetVc(currentVC[routine])
	co.SetWVc(currentWVC[routine])

	switch co.GetOpC() {
	case trace.WaitCondOp:
		CondWait(co)
	case trace.SignalOp:
		CondSignal(co)
	case trace.BroadcastOp:
		CondBroadcast(co)
	}
}

// CondWait updates and calculates the vector clocks given a wait operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondWait(co *trace.TraceElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := co.GetID()
	routine := co.GetRoutine()

	if co.GetTPost() != 0 { // not leak
		if _, ok := currentlyWaiting[id]; !ok {
			currentlyWaiting[id] = make([]int, 0)
		}
		currentlyWaiting[id] = append(currentlyWaiting[id], routine)
	}
	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.TraceElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := co.GetID()
	routine := co.GetRoutine()

	if len(currentlyWaiting[id]) != 0 {
		tWait := currentlyWaiting[id][0]
		currentlyWaiting[id] = currentlyWaiting[id][1:]
		currentVC[tWait].Sync(currentVC[routine])
	}

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.TraceElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := co.GetID()
	routine := co.GetRoutine()

	for _, wait := range currentlyWaiting[id] {
		currentVC[wait].Sync(currentVC[routine])
	}
	currentlyWaiting[id] = make([]int, 0)

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}
