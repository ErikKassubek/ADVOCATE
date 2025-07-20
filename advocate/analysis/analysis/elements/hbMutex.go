// Copyright (c) 2024 Erik Kassubek
//
// File: hbMutex.go
// Brief: Update functions for happens before info for mutex operation
//        Some of the functions start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"advocate/analysis/analysis/scenarios"
	"advocate/analysis/data"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/cssts"
	"advocate/analysis/hb/pog"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// UpdateHBMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateHBMutex(mu *trace.ElementMutex) {
	routine := mu.GetRoutine()
	mu.SetVc(vc.CurrentVC[routine])
	mu.SetWVc(vc.CurrentWVC[routine])

	switch mu.GetOpM() {
	case trace.LockOp:
		Lock(mu)
		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}
	case trace.RLockOp:
		RLock(mu)
		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}
	case trace.TryLockOp:
		if mu.IsSuc() {
			if data.AnalysisCases["unlockBeforeLock"] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
			Lock(mu)
		}
	case trace.TryRLockOp:
		if mu.IsSuc() {
			RLock(mu)
			if data.AnalysisCases["unlockBeforeLock"] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}
	case trace.UnlockOp:
		Unlock(mu)
		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	case trace.RUnlockOp:
		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
		RUnlock(mu)
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}

// UpdateVectorClockAlt stores and updates the vector clock of the trace and element
// if the ignoreCriticalSections flag is set
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateHBMutexAlt(mu *trace.ElementMutex) {
	routine := mu.GetRoutine()
	mu.SetVc(vc.CurrentVC[routine])

	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)
}

// Lock updates and calculates the vector clocks given a lock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Lock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		vc.CurrentVC[routine].Inc(routine)
		vc.CurrentWVC[routine].Inc(routine)
		return
	}

	if e, ok := vc.RelW[id]; ok {
		vc.CurrentVC[routine].Sync(e.Vc)
		pog.AddEdge(e.Elem, mu, false)
		cssts.AddEdge(e.Elem, mu, false)
	}
	if e, ok := vc.RelR[id]; ok {
		vc.CurrentVC[routine].Sync(e.Vc)
		pog.AddEdge(e.Elem, mu, false)
		cssts.AddEdge(e.Elem, mu, false)
	}

	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["leak"] {
		scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 0)
	}

	scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])

	// for fuzzing
	data.CurrentlyHoldLock[id] = mu
	scenarios.IncFuzzingCounter(mu)
}

// Unlock updates and calculates the vector clocks given a unlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Unlock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if mu.GetTPost() == 0 {
		return
	}

	id := mu.GetID()
	routine := mu.GetRoutine()

	vc.RelW[id] = &data.ElemWithVc{
		Elem: mu,
		Vc:   vc.CurrentVC[routine].Copy(),
	}

	vc.RelR[id] = &data.ElemWithVc{
		Elem: mu,
		Vc:   vc.CurrentVC[routine].Copy(),
	}

	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	scenarios.LockSetRemoveLock(routine, id)

	// for fuzzing
	data.CurrentlyHoldLock[id] = nil
}

// RLock updates and calculates the vector clocks given a rlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//
// Returns:
//   - *VectorClock: The new vector clock
func RLock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		vc.CurrentVC[routine].Inc(routine)
		vc.CurrentWVC[routine].Inc(routine)
		return
	}

	if e, ok := vc.RelW[id]; ok {
		vc.CurrentVC[routine].Sync(e.Vc)
		pog.AddEdge(e.Elem, mu, false)
		cssts.AddEdge(e.Elem, mu, false)
	}

	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["leak"] {
		scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 1)
	}

	scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])

	// for fuzzing
	data.CurrentlyHoldLock[id] = mu
	scenarios.IncFuzzingCounter(mu)
}

// RUnlock updates and calculates the vector clocks given a runlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func RUnlock(mu *trace.ElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		vc.CurrentVC[routine].Inc(routine)
		vc.CurrentWVC[routine].Inc(routine)
		return
	}

	if _, ok := vc.RelR[id]; !ok {
		vc.RelR[id] = &data.ElemWithVc{
			Vc:   clock.NewVectorClock(data.GetNoRoutines()),
			Elem: nil,
		}
	}

	vc.RelR[id].Vc.Sync(vc.CurrentVC[routine])
	vc.RelR[id].Elem = mu

	vc.CurrentVC[routine].Inc(routine)
	vc.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	scenarios.LockSetRemoveLock(routine, id)
	// for fuzzing
	data.CurrentlyHoldLock[id] = nil
}
