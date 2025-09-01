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
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/timer"
)

// UpdateHBAtomic update the vector clocks for an atomic operation
//
// Parameter:
//   - at *trace.TraceElementAtomic: the atomic operation
func UpdateHBAtomic(at *trace.ElementAtomic) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBAtomic(at)
	}

	if CalcPog {
		pog.UpdateHBAtomic(at)
	}

	if CalcCssts {
		cssts.UpdateHBAtomic(at)
	}
}

// UpdateHBChannel updates the hb info of the trace for a channel operation
//
// Parameter
//   - ch *trace.TraceElementChannel: the channel trace operation
func UpdateHBChannel(ch *trace.ElementChannel) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBChannel(ch)
	}

	if CalcPog {
		pog.UpdateHBChannel(ch)
	}

	if CalcCssts {
		cssts.UpdateHBChannel(ch)
	}
}

// UpdateHBSelect updates the hb info of the trace for a select
//
// Parameter
//   - ch *trace.TraceElementChannel: the channel trace operation
func UpdateHBSelect(se *trace.ElementSelect) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBSelect(se)
	}

	if CalcPog {
		pog.UpdateHBSelect(se)
	}

	if CalcCssts {
		cssts.UpdateHBSelect(se)
	}
}

// UpdateHBCond updates the hb info of the trace for a conditional variables
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBCond(co *trace.ElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBCond(co)
	}

	if CalcPog {
		pog.UpdateHBCond(co)
	}

	if CalcCssts {
		cssts.UpdateHBCond(co)
	}
}

// UpdateHBFork updates the hb info of the trace for a fork
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
func UpdateHBFork(fo *trace.ElementFork) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	// Note: the update for the pog is done in AddEdgeSameRoutineAndFork

	if CalcVC {
		vc.UpdateHBFork(fo)
	}

	if CalcCssts {
		cssts.UpdateHBFork(fo)
	}
}

// UpdateHBMutex updates the hb info of the trace for a mutex operation
//
// Parameter
//   - co *trace.TraceElementCond: the conditional trace operation
//   - alt bool: if IgnoreCriticalSections is set
func UpdateHBMutex(mu *trace.ElementMutex, alt bool) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBMutex(mu, alt)
	}

	if CalcPog {
		pog.UpdateHBMutex(mu)
	}

	if CalcCssts {
		cssts.UpdateHBMutex(mu)
	}
}

// UpdateHBNew stores the hb info of the trace for a new element
//
// Parameter
//   - n *trace.TraceElementNew: the new trace operation
func UpdateHBNew(n *trace.ElementNew) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	// For new and routine end elements, we only store the current vc
	// Therefore, the graph based methods do not do anything

	if CalcVC {
		vc.UpdateHBNew(n)
	}
}

// UpdateHBOnce updates the hb info of the trace for a once
//
// Parameter
//   - on *trace.TraceElementOnce: the once trace operation
func UpdateHBOnce(on *trace.ElementOnce) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBOnce(on)
	}

	if CalcPog {
		pog.UpdateHBOnce(on)
	}

	if CalcCssts {
		cssts.UpdateHBOnce(on)
	}
}

// UpdateHBRoutineEnd stores the hb info of the trace for a routine end element
//
// Parameter
//   - n *trace.TraceElementNew: the new trace operation
func UpdateHBRoutineEnd(n *trace.ElementRoutineEnd) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	// For new and routine end elements, we only store the current vc
	// Therefore, the graph based methods do not do anything

	if CalcVC {
		vc.UpdateHBRoutineEnd(n)
	}
}

// UpdateHBWait updates the hb info of the trace for a fait group
//
// Parameter
//   - wa *trace.TraceElementWait: the wait group trace operation
func UpdateHBWait(wa *trace.ElementWait) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if CalcVC {
		vc.UpdateHBWait(wa)
	}

	if CalcPog {
		pog.UpdateHBWait(wa)
	}

	if CalcCssts {
		cssts.UpdateHBWait(wa)
	}
}
