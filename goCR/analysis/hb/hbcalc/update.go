// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update hb info for the different types
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package hbcalc

import (
	"goCR/analysis/hb/vc"
	"goCR/trace"
	"goCR/utils/timer"
)

// UpdateHBAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
//   - alt bool: update if the ignoreCriticalSections tag has been set
func UpdateHBAtomic(at *trace.ElementAtomic, alt bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBAtomic(at, alt)
}

// UpdateHBChannel updates the hb info of the trace for a channel operation
//
// Parameter
//   - ch *trace.TraceElementChannel: the channel trace operation
func UpdateHBChannel(ch *trace.ElementChannel) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBChannel(ch)
}

// UpdateHBSelect updates the hb info of the trace for a select
//
// Parameter
//   - ch *trace.TraceElementChannel: the channel trace operation
func UpdateHBSelect(se *trace.ElementSelect) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBSelect(se)
}

// UpdateHBCond updates the hb info of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBCond(co)

}

// UpdateHBFork updates the hb info of the trace for a fork
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBFork(fo *trace.ElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBFork(fo)

}

// UpdateHBMutex updates the hb info of the trace for a mutex operation
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
//   - alt bool: if IgnoreCriticalSections is set
func UpdateHBMutex(mu *trace.ElementMutex, alt bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBMutex(mu, alt)
}

// UpdateHBNew stores the hb info of the trace for a new element
//
// Parameter
//   - n *trace.TraceElementNew: the new trace operation
func UpdateHBNew(n *trace.ElementNew) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBNew(n)
}

// UpdateHBOnce updates the hb info of the trace for a once
//
// Parameter
//   - on *trace.TraceElementOnce: the once trace operation
func UpdateHBOnce(on *trace.ElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBOnce(on)
}

// UpdateHBRoutineEnd stores the hb info of the trace for a routine end element
//
// Parameter
//   - n *trace.TraceElementNew: the new trace operation
func UpdateHBRoutineEnd(n *trace.ElementRoutineEnd) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBRoutineEnd(n)
}

// UpdateHBWait updates the hb info of the trace for a fait group
//
// Parameter
//   - wa *trace.TraceElementWait: the wait group trace operation
func UpdateHBWait(wa *trace.ElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	vc.UpdateHBWait(wa)

}
