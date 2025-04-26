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
	"analyzer/clock"
	"analyzer/timer"
	"analyzer/trace"
	"analyzer/utils"
)

// UpdateVCMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateVCMutex(mu *trace.TraceElementMutex) {
	routine := mu.GetRoutine()
	mu.SetVc(currentVC[routine])
	mu.SetWVc(currentWVC[routine])

	switch mu.GetOpM() {
	case trace.LockOp:
		Lock(mu)
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockLock(mu)
		}
	case trace.RLockOp:
		RLock(mu)
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockLock(mu)
		}
	case trace.TryLockOp:
		if mu.IsSuc() {
			if analysisCases["unlockBeforeLock"] {
				checkForUnlockBeforeLockLock(mu)
			}
			Lock(mu)
		}
	case trace.TryRLockOp:
		if mu.IsSuc() {
			RLock(mu)
			if analysisCases["unlockBeforeLock"] {
				checkForUnlockBeforeLockLock(mu)
			}
		}
	case trace.UnlockOp:
		Unlock(mu)
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockUnlock(mu)
		}
	case trace.RUnlockOp:
		if analysisCases["unlockBeforeLock"] {
			checkForUnlockBeforeLockUnlock(mu)
		}
		RUnlock(mu)
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		utils.LogError(err)
	}
}

// UpdateVectorClockAlt stores and updates the vector clock of the trace and element
// if the ignoreCriticalSections flag is set
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateVCMutexAlt(mu *trace.TraceElementMutex) {
	routine := mu.GetRoutine()
	mu.SetVc(currentVC[routine])

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)
}

// Create a new relW and relR if needed
//
// Parameter:
//   - index int: The id of the atomic variable
//   - nRout int: The number of routines in the trace
func newRel(index int, nRout int) {
	if _, ok := relW[index]; !ok {
		relW[index] = clock.NewVectorClock(nRout)
	}
	if _, ok := relR[index]; !ok {
		relR[index] = clock.NewVectorClock(nRout)
	}
}

// Lock updates and calculates the vector clocks given a lock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Lock(mu *trace.TraceElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		currentVC[routine].Inc(routine)
		currentWVC[routine].Inc(routine)
		return
	}

	currentVC[routine].Sync(relW[id])
	currentVC[routine].Sync(relR[id])

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["leak"] {
		addMostRecentAcquireTotal(mu, currentVC[routine], 0)
	}

	lockSetAddLock(mu, currentWVC[routine])

	// for fuzzing
	currentlyHoldLock[id] = mu
	incFuzzingCounter(mu)
}

// Unlock updates and calculates the vector clocks given a unlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Unlock(mu *trace.TraceElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if mu.GetTPost() == 0 {
		return
	}

	id := mu.GetID()
	routine := mu.GetRoutine()

	newRel(id, currentVC[routine].GetSize())
	relW[id] = currentVC[routine].Copy()
	relR[id] = currentVC[routine].Copy()

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(routine, id)

	// for fuzzing
	currentlyHoldLock[id] = nil
}

// RLock updates and calculates the vector clocks given a rlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//
// Returns:
//   - *VectorClock: The new vector clock
func RLock(mu *trace.TraceElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		currentVC[routine].Inc(routine)
		currentWVC[routine].Inc(routine)
		return
	}

	newRel(id, currentVC[routine].GetSize())
	currentVC[routine].Sync(relW[id])

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["leak"] {
		addMostRecentAcquireTotal(mu, currentVC[routine], 1)
	}

	lockSetAddLock(mu, currentWVC[routine])

	// for fuzzing
	currentlyHoldLock[id] = mu
	incFuzzingCounter(mu)
}

// RUnlock updates and calculates the vector clocks given a runlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func RUnlock(mu *trace.TraceElementMutex) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		currentVC[routine].Inc(routine)
		currentWVC[routine].Inc(routine)
		return
	}

	newRel(id, currentVC[routine].GetSize())
	relR[id].Sync(currentVC[routine])

	currentVC[routine].Inc(routine)
	currentWVC[routine].Inc(routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(routine, id)
	// for fuzzing
	currentlyHoldLock[id] = nil
}
