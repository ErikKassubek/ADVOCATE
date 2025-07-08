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

package analysis

import (
	"advocate/analysis/concurrent/clock"
	"advocate/analysis/concurrent/cssts"
	"advocate/analysis/concurrent/pog"
	"advocate/analysis/data"
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
	mu.SetVc(data.CurrentVC[routine])
	mu.SetWVc(data.CurrentWVC[routine])

	switch mu.GetOpM() {
	case trace.LockOp:
		Lock(mu)
		if data.AnalysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockLock(mu)
		}
	case trace.RLockOp:
		RLock(mu)
		if data.AnalysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockLock(mu)
		}
	case trace.TryLockOp:
		if mu.IsSuc() {
			if data.AnalysisCases["unlockBeforeLock"] {
				checkForUnlockBeforeLockLock(mu)
			}
			Lock(mu)
		}
	case trace.TryRLockOp:
		if mu.IsSuc() {
			RLock(mu)
			if data.AnalysisCases["unlockBeforeLock"] {
				checkForUnlockBeforeLockLock(mu)
			}
		}
	case trace.UnlockOp:
		Unlock(mu)
		if data.AnalysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockUnlock(mu)
		}
	case trace.RUnlockOp:
		if data.AnalysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockUnlock(mu)
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
	mu.SetVc(data.CurrentVC[routine])

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
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
		data.CurrentVC[routine].Inc(routine)
		data.CurrentWVC[routine].Inc(routine)
		return
	}

	if e, ok := data.RelW[id]; ok {
		data.CurrentVC[routine].Sync(e.Vc)
		pog.AddEdge(e.Elem, mu, false)
		cssts.AddEdge(e.Elem, mu, false)
	}
	if e, ok := data.RelR[id]; ok {
		data.CurrentVC[routine].Sync(e.Vc)
		pog.AddEdge(e.Elem, mu, false)
		cssts.AddEdge(e.Elem, mu, false)
	}

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["leak"] {
		addMostRecentAcquireTotal(mu, data.CurrentVC[routine], 0)
	}

	lockSetAddLock(mu, data.CurrentWVC[routine])

	// for fuzzing
	data.CurrentlyHoldLock[id] = mu
	incFuzzingCounter(mu)
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

	data.RelW[id] = &data.ElemWithVc{
		Elem: mu,
		Vc:   data.CurrentVC[routine].Copy(),
	}

	data.RelR[id] = &data.ElemWithVc{
		Elem: mu,
		Vc:   data.CurrentVC[routine].Copy(),
	}

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(routine, id)

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
		data.CurrentVC[routine].Inc(routine)
		data.CurrentWVC[routine].Inc(routine)
		return
	}

	if e, ok := data.RelW[id]; ok {
		data.CurrentVC[routine].Sync(e.Vc)
		pog.AddEdge(e.Elem, mu, false)
		cssts.AddEdge(e.Elem, mu, false)
	}

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if data.AnalysisCases["leak"] {
		addMostRecentAcquireTotal(mu, data.CurrentVC[routine], 1)
	}

	lockSetAddLock(mu, data.CurrentWVC[routine])

	// for fuzzing
	data.CurrentlyHoldLock[id] = mu
	incFuzzingCounter(mu)
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
		data.CurrentVC[routine].Inc(routine)
		data.CurrentWVC[routine].Inc(routine)
		return
	}

	if _, ok := data.RelR[id]; !ok {
		data.RelR[id] = &data.ElemWithVc{
			Vc:   clock.NewVectorClock(data.GetNoRoutines()),
			Elem: nil,
		}
	}

	data.RelR[id].Vc.Sync(data.CurrentVC[routine])
	data.RelR[id].Elem = mu

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(routine, id)
	// for fuzzing
	data.CurrentlyHoldLock[id] = nil
}
