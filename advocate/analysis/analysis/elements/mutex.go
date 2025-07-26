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
	"advocate/analysis/hb/hbCalc"
	"advocate/analysis/hb/vc"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
//   - alt bool: if IgnoreCriticalSections is set
func UpdateMutex(mu *trace.ElementMutex, alt bool) {
	hbCalc.UpdateHBMutex(mu, alt)

	routine := mu.GetRoutine()
	id := mu.GetID()

	switch mu.GetOpM() {
	case trace.LockOp:
		if data.AnalysisCases["leak"] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 0)
		}

		scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])

		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}

	case trace.RLockOp:
		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
		scenarios.IncFuzzingCounter(mu)

		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockLock(mu)
		}
	case trace.TryLockOp:
		if mu.IsSuc() {
			if data.AnalysisCases["unlockBeforeLock"] {
				scenarios.CheckForUnlockBeforeLockLock(mu)
			}
		}
	case trace.TryRLockOp:
		if mu.IsSuc() {
			if data.AnalysisCases["unlockBeforeLock"] {
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

		scenarios.LockSetRemoveLock(routine, id)

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	case trace.RUnlockOp:
		data.RelR[id].Elem = mu
		if data.AnalysisCases["leak"] {
			scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 1)
		}

		scenarios.LockSetAddLock(mu, vc.CurrentWVC[routine])

		scenarios.LockSetRemoveLock(routine, id)
		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

		if data.AnalysisCases["unlockBeforeLock"] {
			scenarios.CheckForUnlockBeforeLockUnlock(mu)
		}
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}
