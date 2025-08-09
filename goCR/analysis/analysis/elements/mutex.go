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
		scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 0)

		// for fuzzing
		data.CurrentlyHoldLock[id] = mu

	case trace.RLockOp:
		// for fuzzing
		data.CurrentlyHoldLock[id] = mu
	case trace.TryLockOp, trace.TryRLockOp:
	case trace.UnlockOp:
		data.RelW[id] = &data.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		data.RelR[id] = &data.ElemWithVc{
			Elem: mu,
			Vc:   vc.CurrentVC[routine].Copy(),
		}

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil
	case trace.RUnlockOp:
		data.RelR[id].Elem = mu
		scenarios.AddMostRecentAcquireTotal(mu, vc.CurrentVC[routine], 1)

		// for fuzzing
		data.CurrentlyHoldLock[id] = nil

	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}
