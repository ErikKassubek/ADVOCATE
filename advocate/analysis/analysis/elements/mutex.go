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
	case trace.MutexLock:
		if data.AnalysisCasesMap[flags.Leak] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
		}

		scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])

		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if data.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	case trace.MutexRLock:
		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if data.AnalysisCasesMap[flags.Leak] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
		}

		if data.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}
	case trace.MutexTryLock:
		if mu.IsSuc() {
			if data.AnalysisCasesMap[flags.Leak] {
				scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
			}

			if data.AnalysisCasesMap[flags.UnlockBeforeLock] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			if data.AnalysisCasesMap[flags.Leak] {
				scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine])
			}

			if data.AnalysisCasesMap[flags.UnlockBeforeLock] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}
	case trace.MutexUnlock:
		data.RelW[id] = &data.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		data.RelR[id] = &data.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		if data.AnalysisCasesMap[flags.MixedDeadlock] {
			scenarios.LockSetRemoveLock(routine, id)
		}

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

		if data.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	case trace.MutexRUnlock:
		data.RelR[id].Elem = mu

		if data.AnalysisCasesMap[flags.MixedDeadlock] {
			scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])
			scenarios.LockSetRemoveLock(routine, id)
		}

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

		if data.AnalysisCasesMap[flags.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}
