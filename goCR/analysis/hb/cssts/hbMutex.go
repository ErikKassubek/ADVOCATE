//
// File: hbAtomic.go
// Brief: Update the cssts for mutex operations
//
// Created: 2025-07-20
//
// License: BSD-3-Clause

package cssts

import (
	"goCR/analysis/data"
	"goCR/analysis/hb/clock"
	"goCR/trace"
	"goCR/utils/log"
)

// UpdateHBMutex updates the cssts of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateHBMutex(mu *trace.ElementMutex) {
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
	case trace.RUnlockOp:
		RUnlock(mu)
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}

// Lock updates the cssts given a lock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Lock(mu *trace.ElementMutex) {
	id := mu.GetID()

	if mu.GetTPost() == 0 {
		return
	}

	if e, ok := data.RelW[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
	if e, ok := data.RelR[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
}

// RLock updates the cssts given a rlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
//
// Returns:
//   - *VectorClock: The new vector clock
func RLock(mu *trace.ElementMutex) {
	id := mu.GetID()

	if mu.GetTPost() == 0 {
		return
	}

	if e, ok := data.RelW[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
}

// RUnlock updates the cssts given a runlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func RUnlock(mu *trace.ElementMutex) {
	id := mu.GetID()

	if mu.GetTPost() == 0 {
		return
	}

	if _, ok := data.RelR[id]; !ok {
		data.RelR[id] = &data.ElemWithVc{
			Vc:   clock.NewVectorClock(data.GetNoRoutines()),
			Elem: nil,
		}
	} else {
		AddEdge(mu, data.RelR[id].Elem, false)
	}
}
