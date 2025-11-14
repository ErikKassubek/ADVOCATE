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

		scenarios.LockSetAddLock(mu, vc.CurrentVC[routine])

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

		scenarios.LockSetAddLock(mu, vc.CurrentVC[routine])

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

			scenarios.LockSetAddLock(mu, vc.CurrentVC[routine])

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

			scenarios.LockSetAddLock(mu, vc.CurrentVC[routine])

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

		scenarios.LockSetRemoveLock(mu, vc.CurrentVC[routine])

		baseA.CurrentlyHoldLock[id] = nil

		if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	// --------- RUNLOCK (read) ---------
	case trace.MutexRUnlock:
		if baseA.RelR[id] != nil {
			baseA.RelR[id].Elem = mu
		}

		scenarios.LockSetRemoveLock(mu, vc.CurrentVC[routine])

		baseA.CurrentlyHoldLock[id] = nil

		if baseA.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}

	default:
		log.Error("Unknown mutex operation: " + mu.ToString())
	}
}
