//
// File: hbMutex.go
// Brief: Update functions for happens before info for mutex operation
//        Some of the functions start analysis functions
//
// Created: 2023-07-25
//
// License: BSD-3-Clause

package elements

import (
	"goCR/analysis/analysis/scenarios"
	"goCR/analysis/data"
	"goCR/analysis/hb/hbcalc"
	"goCR/analysis/hb/vc"
	"goCR/trace"
	"goCR/utils/log"
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

	switch mu.GetOpM() {
	case trace.LockOp:
		if data.AnalysisCasesMap[data.Leak] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 0)
		}

		scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])

		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if data.AnalysisCasesMap[data.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	case trace.RLockOp:
		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if data.AnalysisCasesMap[data.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}
	case trace.TryLockOp:
		if mu.IsSuc() {
			if data.AnalysisCasesMap[data.UnlockBeforeLock] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}
	case trace.TryRLockOp:
		if mu.IsSuc() {
			if data.AnalysisCasesMap[data.UnlockBeforeLock] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}
	case trace.UnlockOp:
		data.RelW[id] = &data.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		data.RelR[id] = &data.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		if data.AnalysisCasesMap[data.MixedDeadlock] {
			scenarios.LockSetRemoveLock(routine, id)
		}

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

		if data.AnalysisCasesMap[data.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	case trace.RUnlockOp:
		data.RelR[id].Elem = mu
		if data.AnalysisCasesMap[data.Leak] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 1)
		}

		if data.AnalysisCasesMap[data.MixedDeadlock] {
			scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])
			scenarios.LockSetRemoveLock(routine, id)
		}

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

		if data.AnalysisCasesMap[data.UnlockBeforeLock] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}
