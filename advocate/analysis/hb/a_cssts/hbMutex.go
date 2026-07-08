// Copyright (c) 2026 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Update the cssts for mutex operations
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package a_cssts

import (
	"advocate/analysis/a_base"
	"advocate/analysis/hb/a_clock"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBMutex updates the cssts of the trace and element
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

// Lock updates the cssts given a lock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func Lock(mu *trace.ElementMutex) {
	id := mu.GetObjId()

	if mu.GetTPost() == 0 {
		return
	}

	if e, ok := a_base.RelW[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
	if e, ok := a_base.RelR[id]; ok {
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
	id := mu.GetObjId()

	if mu.GetTPost() == 0 {
		return
	}

	if e, ok := a_base.RelW[id]; ok {
		AddEdge(e.Elem, mu, false)
	}
}

// RUnlock updates the cssts given a runlock operation
//
// Parameter:
//   - mu *TraceElementMutex: The trace element
func RUnlock(mu *trace.ElementMutex) {
	id := mu.GetObjId()

	if mu.GetTPost() == 0 {
		return
	}

	if _, ok := a_base.RelR[id]; !ok {
		a_base.RelR[id] = &a_base.ElemWithVc{
			Vc:   a_clock.NewVectorClock(a_base.GetNoRoutines()),
			Elem: nil,
		}
	} else {
		AddEdge(mu, a_base.RelR[id].Elem, false)
	}
}
