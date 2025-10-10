// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the pog for mutex operations
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package pog

import (
	"advocate/analysis/data"
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBMutex updates the pog of the trace and element
//
// Parameter:
//   - mu *trace.TraceElementMutex: the mutex trace element
func UpdateHBMutex(mu *trace.ElementMutex) {
	switch mu.GetType(true) {
	case trace.MutexLock:
		Lock(mu)
	case trace.MutexRLock:
		RLock(mu)
	case trace.MutexTryLock:
		if mu.IsSuc() {
			Lock(mu)
		}
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			RLock(mu)
		}
	case trace.MutexUnlock:
	case trace.MutexRUnlock:
		RUnlock(mu)
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}

// Lock updates the pog given a lock operation
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

// RLock updates the pog given a rlock operation
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

// RUnlock updates the pog given a runlock operation
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
