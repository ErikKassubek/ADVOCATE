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

// AdvocateRoutine is a struct to store the trace of a routine
// Fields:
//   - id uint64: the id of the routine
//   - replayID int: when used in reply, id of the new routine in the replayed trace
//   - maxObjectId uint64: the maximum id of elements in the trace
//   - G *g: the g struct of the routine
//   - Trace []traceElem: the trace of the routine
type AdvocateRoutine struct {
	id          uint64
	replayID    int
	maxObjectId uint64
	G           *g
	Trace       []traceElem
}

// Create a new advocate routine
// Params:
//   - g: the g struct of the routine
//   - replayRoutine int: when used in reply, id of the new routine in the replayed trace
//
// Return:
//   - the new advocate routine
func newAdvocateRoutine(g *g, replayRoutine int) *AdvocateRoutine {

	// ignore the internal routines that are run before the main function starts
	if advocateTracingDisabled && GetNextAdvocateRoutineID() != 1 {
		return &AdvocateRoutine{
			id:          0,
			maxObjectId: 0,
			G:           g,
			Trace:       make([]traceElem, 0),
			replayID:    replayRoutine,
		}
	}

	routine := &AdvocateRoutine{
		id:          GetNewAdvocateRoutineID(),
		maxObjectId: 0,
		G:           g,
		Trace:       make([]traceElem, 0),
		replayID:    replayRoutine,
	}

	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	if AdvocateRoutines == nil {
		AdvocateRoutines = make(map[uint64]*AdvocateRoutine)
	}

	AdvocateRoutines[routine.id] = routine

	return routine
}

// Add an element to the trace of the current routine
// Params:
//   - elem: the element to add
//
// Return:
//   - the index of the element in the trace
func (gi *AdvocateRoutine) addToTrace(elem traceElem) int {
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

// Update an element in the trace of the current routine
// Params:
//   - index: the index of the element to update
//   - elem: the new element
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

// Get the current routine
// Return:
//   - *AdvocateRoutine: the current routine
func currentGoRoutine() *AdvocateRoutine {
	return getg().advocateRoutineInfo
}

// GetRoutineID gets the id of the current routine
// Return:
//   - int: id of the current routine, 0 if current routine is nil
func GetRoutineID() int {
	currentRoutine := currentGoRoutine()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.id)
}

// GetReplayRoutineID returns the replay id of the current routine
// Return:
//   - int: replay id of the current routine, 0 if current routine is nil
func GetReplayRoutineID() int {
	currentRoutine := currentGoRoutine()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.replayID)
}

// ADVOCATE-FILE-END
