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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"advocate/utils/log"
)

// UpdateHBMutex updates the pog of the trace and element
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - mu *trace.TraceElementMutex: the mutex trace element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func UpdateHBMutex(graph *PoGraph, mu *trace.ElementMutex, recorded bool) {
	switch mu.GetType(true) {
	case trace.MutexLock:
		Lock(graph, mu)
	case trace.MutexRLock:
		RLock(graph, mu, recorded)
	case trace.MutexTryLock:
		if mu.IsSuc() {
			Lock(graph, mu)
		}
	case trace.MutexTryRLock:
		if mu.IsSuc() {
			RLock(graph, mu, recorded)
		}
	case trace.MutexUnlock:
	case trace.MutexRUnlock:
		RUnlock(graph, mu, recorded)
	default:
		err := "Unknown mutex operation: " + mu.ToString()
		log.Error(err)
	}
}

// Lock updates the pog given a lock operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - mu *TraceElementMutex: The trace element
func Lock(graph *PoGraph, mu *trace.ElementMutex) {
	id := mu.GetObjId()

	if mu.GetTPost() == 0 {
		return
	}

	if e, ok := baseA.RelW[id]; ok {
		if graph != nil {
			graph.AddEdge(e.Elem, mu)
		} else {
			AddEdge(e.Elem, mu, false)
		}
	}
	if e, ok := baseA.RelR[id]; ok {
		if graph != nil {
			graph.AddEdge(e.Elem, mu)
		} else {
			AddEdge(e.Elem, mu, false)
		}
	}
}

// RLock updates the pog given a rlock operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - mu *TraceElementMutex: The trace element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func RLock(graph *PoGraph, mu *trace.ElementMutex, recorded bool) {
	id := mu.GetObjId()

	if recorded && mu.GetTPost() == 0 {
		return
	}

	if e, ok := baseA.RelW[id]; ok {
		if graph != nil {
			graph.AddEdge(e.Elem, mu)
		} else {
			AddEdge(e.Elem, mu, false)
		}
	}
}

// RUnlock updates the pog given a runlock operation
//
// Parameter:
//   - graph *PoGraph: if nil, use the standard po/poivert, otherwise add to given
//   - mu *TraceElementMutex: The trace element
//   - recorded bool: true if it is a recorded trace, false if it is rewritten/mutated
func RUnlock(graph *PoGraph, mu *trace.ElementMutex, recorded bool) {
	id := mu.GetObjId()

	if recorded && mu.GetTPost() == 0 {
		return
	}

	if _, ok := baseA.RelR[id]; !ok {
		baseA.RelR[id] = &baseA.ElemWithVc{
			Vc:   clock.NewVectorClock(baseA.GetNoRoutines()),
			Elem: nil,
		}
	} else {
		if graph != nil {
			graph.AddEdge(mu, baseA.RelR[id].Elem)
		} else {
			AddEdge(mu, baseA.RelR[id].Elem, false)
		}
	}
}
