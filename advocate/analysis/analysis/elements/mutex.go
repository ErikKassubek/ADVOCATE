// advocate/analysis/analysis/elements/mutex.go

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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/hbcalc"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/flags"
	"advocate/utils/log"
)

// UpdateMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
//   - alt bool: if IgnoreCriticalSections is set
func UpdateMutex(mu *trace.ElementMutex, alt bool) {
	hbcalc.UpdateHBMutex(mu, alt)

	routine := mu.GetRoutine()
	id := mu.GetID()

	switch mu.GetType(true) {

	// --------- WRITE LOCK ---------
	case trace.MutexLock:
		if baseA.AnalysisCasesMap[flags.Leak] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
		}

		LockSetAddLock(mu, false)

		baseA.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	// --------- READ LOCK (RWMutex RLock) ---------
	case trace.MutexRLock:
		if baseA.AnalysisCasesMap[flags.Leak] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
		}

		LockSetAddLock(mu, true)

		baseA.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	// --------- TRY LOCK (write) ---------
	case trace.MutexTryLock:
		if mu.IsSuc() {
			if baseA.AnalysisCasesMap[flags.Leak] {
				scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
			}

			LockSetAddLock(mu, false)

			baseA.CurrentlyHoldLock[id] = mu
			scenarios.IncFuzzingCounter(mu)

			if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}

	// --------- TRY RLOCK (read) ---------
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			if baseA.AnalysisCasesMap[flags.Leak] {
				scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
			}

			LockSetAddLock(mu, true)

			baseA.CurrentlyHoldLock[id] = mu
			scenarios.IncFuzzingCounter(mu)

			if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}

	// --------- UNLOCK (write) ---------
	case trace.MutexUnlock:
		baseA.RelW[id] = &baseA.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		baseA.RelR[id] = &baseA.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		LockSetRemoveLock(mu, false)

		baseA.CurrentlyHoldLock[id] = nil

		if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	// --------- RUNLOCK (read) ---------
	case trace.MutexRUnlock:
		if baseA.RelR[id] != nil {
			baseA.RelR[id].Elem = mu
		}

		LockSetRemoveLock(mu, true)

		baseA.CurrentlyHoldLock[id] = nil

		if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	default:
		log.Error("Unknown mutex operation: " + mu.ToString())
	}
}

func ensureLockTracking(routine int) {
	if _, ok := baseA.LockSet[routine]; !ok {
		baseA.LockSet[routine] = make(map[int]string) // lockID -> TID
	}
	if _, ok := baseA.MostRecentAcquire[routine]; !ok {
		baseA.MostRecentAcquire[routine] = make(map[int]baseA.ElemWithVc)
	}
	if _, ok := baseA.MostRecentRelease[routine]; !ok {
		baseA.MostRecentRelease[routine] = make(map[int]baseA.ElemWithVc)
	}
	if _, ok := baseA.RLockCount[routine]; !ok {
		baseA.RLockCount[routine] = make(map[int]int) // lockID -> count
	}
}

// LockSetAddLock adds to LockSet and MostRecentAcquire
func LockSetAddLock(mu *trace.ElementMutex, isReadLock bool) {
	routine := mu.GetRoutine()
	id := mu.GetID()

	ensureLockTracking(routine)

	// RLock-Counter
	if isReadLock {
		baseA.RLockCount[routine][id]++
	}

	baseA.LockSet[routine][id] = mu.GetTID()

	baseA.MostRecentAcquire[routine][id] = baseA.ElemWithVc{
		Vc:   vc.CurrentVC[routine].Copy(),
		Elem: mu,
	}
}

// LockSetRemoveLock removes from Lockset and MostRecentRelease
func LockSetRemoveLock(mu *trace.ElementMutex, isReadUnlock bool) {
	routine := mu.GetRoutine()
	id := mu.GetID()

	ensureLockTracking(routine)

	baseA.MostRecentRelease[routine][id] = baseA.ElemWithVc{
		Vc:   vc.CurrentVC[routine].Copy(),
		Elem: mu,
	}

	if isReadUnlock {
		baseA.RLockCount[routine][id]--
		if baseA.RLockCount[routine][id] <= 0 {
			baseA.RLockCount[routine][id] = 0
			delete(baseA.LockSet[routine], id)
		}
	} else {
		delete(baseA.LockSet[routine], id)

		if cnt, ok := baseA.RLockCount[routine][id]; ok && cnt > 0 {
			baseA.RLockCount[routine][id] = 0
		}
	}
}
