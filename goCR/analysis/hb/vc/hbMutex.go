//
// File: hbAtomic.go
// Brief: Update the vc for mutex operations
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package vc

import (
	"goCR/analysis/data"
	"goCR/analysis/hb/clock"
	"goCR/trace"
	"goCR/utils/log"
)

// UpdateHBMutex store and update the vector clock of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
//   - alt bool: if Ignore critical sections is set
func UpdateHBMutex(mu *trace.ElementMutex, alt bool) {
	routine := mu.GetRoutine()
	mu.SetVc(CurrentVC[routine])
	mu.SetWVc(CurrentWVC[routine])

	if !alt {
		switch mu.GetOpM() {
		case trace.LockOp:
			Lock(mu)
		case trace.RLockOp:
			RLock(mu)
		case trace.TryLockOp:
			if mu.IsSuc() {
				Lock(mu)
			}
		case trace.TryRLockOp:
			if mu.IsSuc() {
				RLock(mu)
			}
		case trace.UnlockOp:
			// only increases counter, no sync
		case trace.RUnlockOp:
			RUnlock(mu)
		default:
			err := "Unknown mutex operation: " + mu.ToString()
			log.Error(err)
		}
	}

	if mu.GetTPost() != 0 {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
	}
}

// UpdateHBMutexAlt stores and updates the vector clock of the trace and element
// if the ignoreCriticalSections flag is set
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateHBMutexAlt(mu *trace.ElementMutex) {
	routine := mu.GetRoutine()
	mu.SetVc(CurrentVC[routine])
}

// Lock updates and calculates the vector clocks given a lock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Lock(mu *trace.ElementMutex) {
	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
		return
	}

	if e, ok := data.RelW[id]; ok {
		CurrentVC[routine].Sync(e.Vc)
	}
	if e, ok := data.RelR[id]; ok {
		CurrentVC[routine].Sync(e.Vc)
	}
}

// RLock updates and calculates the vector clocks given a rlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//
// Returns:
//   - *VectorClock: The new vector clock
func RLock(mu *trace.ElementMutex) {
	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
		return
	}

	if e, ok := data.RelW[id]; ok {
		CurrentVC[routine].Sync(e.Vc)
	}
}

// RUnlock updates and calculates the vector clocks given a runlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func RUnlock(mu *trace.ElementMutex) {
	id := mu.GetID()
	routine := mu.GetRoutine()

	if mu.GetTPost() == 0 {
		CurrentVC[routine].Inc(routine)
		CurrentWVC[routine].Inc(routine)
		return
	}

	if _, ok := data.RelR[id]; !ok {
		data.RelR[id] = &data.ElemWithVc{
			Vc:   clock.NewVectorClock(data.GetNoRoutines()),
			Elem: nil,
		}
	}

	data.RelR[id].Vc.Sync(CurrentVC[routine])
}
