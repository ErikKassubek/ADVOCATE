// Copyright (c) 2024 Erik Kassubek
//
// File: hbCond.go
// Brief: Update functions for happens before info for conditional variables operations
//
// Author: Erik Kassubek
// Created: 2024-01-09
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateHBCond updates the vector clock of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	routine := co.GetRoutine()
	co.SetVc(data.CurrentVC[routine])
	co.SetWVc(data.CurrentWVC[routine])

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
func CondWait(co *trace.ElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := co.GetID()
	routine := co.GetRoutine()

	if co.GetTPost() != 0 { // not leak
		if _, ok := data.CurrentlyWaiting[id]; !ok {
			data.CurrentlyWaiting[id] = make([]int, 0)
		}
		data.CurrentlyWaiting[id] = append(data.CurrentlyWaiting[id], routine)
	}
	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}

// CondSignal updates and calculates the vector clocks given a signal operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondSignal(co *trace.ElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := co.GetID()
	routine := co.GetRoutine()

	if len(data.CurrentlyWaiting[id]) != 0 {
		tWait := data.CurrentlyWaiting[id][0]
		data.CurrentlyWaiting[id] = data.CurrentlyWaiting[id][1:]
		data.CurrentVC[tWait].Sync(data.CurrentVC[routine])
	}

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}

// CondBroadcast updates and calculates the vector clocks given a broadcast operation
//
// Parameter:
//   - co *TraceElementCond: The trace element
func CondBroadcast(co *trace.ElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := co.GetID()
	routine := co.GetRoutine()

	for _, wait := range data.CurrentlyWaiting[id] {
		data.CurrentVC[wait].Sync(data.CurrentVC[routine])
	}
	data.CurrentlyWaiting[id] = make([]int, 0)

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}
