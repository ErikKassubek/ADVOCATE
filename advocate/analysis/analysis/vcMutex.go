// Copyright (c) 2024 Erik Kassubek
//
// File: vcMutex.go
// Brief: Update functions for vector clocks from mutex operation
//        Some of the functions start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/clock"
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// UpdateVCMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateVCMutex(mu *trace.ElementMutex) {
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
func UpdateVCMutexAlt(mu *trace.ElementMutex) {
	routine := mu.GetRoutine()
	mu.SetVc(data.CurrentVC[routine])

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)
}

// Create a new data.RelW and data.RelR if needed
//
// Parameter:
//   - index int: The id of the atomic variable
//   - nRout int: The number of routines in the trace
func newRel(index int, nRout int) {
	if _, ok := data.RelW[index]; !ok {
		data.RelW[index] = clock.NewVectorClock(nRout)
	}
	if _, ok := data.RelR[index]; !ok {
		data.RelR[index] = clock.NewVectorClock(nRout)
	}
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

	data.CurrentVC[routine].Sync(data.RelW[id])
	data.CurrentVC[routine].Sync(data.RelR[id])

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

	newRel(id, data.CurrentVC[routine].GetSize())
	data.RelW[id] = data.CurrentVC[routine].Copy()
	data.RelR[id] = data.CurrentVC[routine].Copy()

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

	newRel(id, data.CurrentVC[routine].GetSize())
	data.CurrentVC[routine].Sync(data.RelW[id])

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

	newRel(id, data.CurrentVC[routine].GetSize())
	data.RelR[id].Sync(data.CurrentVC[routine])

	data.CurrentVC[routine].Inc(routine)
	data.CurrentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(routine, id)
	// for fuzzing
	data.CurrentlyHoldLock[id] = nil
}
