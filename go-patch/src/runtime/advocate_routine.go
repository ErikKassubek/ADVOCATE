// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_routine.go
// Brief: Functionality for the routines
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

var AdvocateRoutines map[uint64]*AdvocateRoutine
var AdvocateRoutinesLock = mutex{}

var projectPath string

// var atomicRecordingDisabled = false

/*
 * AdvocateRoutine is a struct to store the trace of a routine
 * id: the id of the routine
 * maxObjectId: the maximum id of elements in the trace
 * G: the g struct of the routine
 * Trace: the trace of the routine
 */
type AdvocateRoutine struct {
	id          uint64
	maxObjectId uint64
	G           *g
	Trace       []traceElem
	// Atomics     []string
	// lock    *mutex
}

/*
 * Create a new advocate routine
 * Params:
 * 	g: the g struct of the routine
 * Return:
 * 	the new advocate routine
 */
func newAdvocateRoutine(g *g) *AdvocateRoutine {
	routine := &AdvocateRoutine{id: GetAdvocateRoutineID(), maxObjectId: 0,
		G:     g,
		Trace: make([]traceElem, 0),
	}

	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	if AdvocateRoutines == nil {
		AdvocateRoutines = make(map[uint64]*AdvocateRoutine)
	}

	AdvocateRoutines[routine.id] = routine

	return routine
}

/*
 * Add an element to the trace of the current routine
 * Params:
 * 	elem: the element to add
 * Return:
 * 	the index of the element in the trace
 */
func (gi *AdvocateRoutine) addToTrace(elem traceElem) int {
	// do nothing if tracer disabled
	// TODO ADVOCATE Remove when checked in all pre and post func
	if advocateTracingDisabled {
		return -1
	}

	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}

	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *AdvocateRoutine) getElement(index int) traceElem {
	return gi.Trace[index]
}

/*
 * Update an element in the trace of the current routine
 * Params:
 * 	index: the index of the element to update
 * 	elem: the new element
 */
func (gi *AdvocateRoutine) updateElement(index int, elem traceElem) {
	if advocateTracingDisabled {
		return
	}

	if gi == nil {
		return
	}

	if gi.Trace == nil {
		panic("Tried to update element in nil trace")
	}

	if index >= len(gi.Trace) {
		panic("Tried to update element out of bounds")
	}

	gi.Trace[index] = elem
}

/*
 * Get the current routine
 * Return:
 * 	the current routine
 */
func currentGoRoutine() *AdvocateRoutine {
	return getg().advocateRoutineInfo
}

/*
 * GetRoutineID gets the id of the current routine
 * Return:
 * 	id of the current routine, 0 if current routine is nil
 */
func GetRoutineID() int {
	currentRoutine := currentGoRoutine()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.id)
}

// ADVOCATE-FILE-END
