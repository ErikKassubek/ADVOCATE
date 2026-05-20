// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_routine.go
// Brief: Functionality for the routines
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

import "unsafe"

var GocdrRoutines map[uint64]*GocdrRoutine
var GocdrRoutinesLock = mutex{}

var projectPath string

// var atomicRecordingDisabled = false

// GocdrRoutine is a struct to store the trace of a routine
// Fields:
//   - id uint64: the id of the routine
//   - maxObjectId uint64: the maximum id of elements in the trace
//   - G *g: the g struct of the routine
//   - Trace []traceElem: the trace of the routine
//   - replayID int: when used in reply, id of the new routine in the replayed trace
//   - forkFile string: file where the routine was created in, "main" for main routine
//   - forkLine int: line where ther routine was created in, 0 for main routine
//   - parkOn []unsafe.Pointer: list of elements the routine was last parked on
//   - parkPos string: position of last park in form file:line
//   - parkForeverReplay bool: if true, routine parks forever based on replay
//   - wokenByTimeout bool: in replay block was woken up by timeout
type GocdrRoutine struct {
	id                uint64
	maxObjectId       uint64
	G                 *g
	Trace             []traceElem
	replayID          int
	forkFile          string
	forkLine          int
	parkOn            []unsafe.Pointer
	parkPos           string
	parkOp            []Operation
	parkForeverReplay bool
	wokenButTimeout   bool
}

// Create a new gocdr routine
// Params:
//   - g: the g struct of the routine
//   - replayRoutine int: when used in reply, id of the new routine in the replayed trace the replay ids of the routines forked from this routine
//   - file string: file, where the routine was created
//   - line int: line, where the routine was created
//
// Return:
//   - the new gocdr routine
func newGocdrRoutine(g *g, replayRoutine int, file string, line int) *GocdrRoutine {
	// ignore the internal routines that are run before the main/test function starts
	if gocdrTracingDisabled {
		return &GocdrRoutine{
			id:          0,
			maxObjectId: 0,
			G:           g,
			Trace:       make([]traceElem, 0),
			forkFile:    file,
			forkLine:    line,
			replayID:    replayRoutine,
			parkOp:      make([]Operation, 0),
			parkOn:      make([]unsafe.Pointer, 0),
		}
	}

	gocdrRoutineInfo := &GocdrRoutine{
		id:          GetNewGocdrRoutineID(),
		maxObjectId: 0,
		G:           g,
		Trace:       make([]traceElem, 0),
		replayID:    replayRoutine,
		parkOn:      make([]unsafe.Pointer, 0),
	}

	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)

	if GocdrRoutines == nil {
		GocdrRoutines = make(map[uint64]*GocdrRoutine)
	}

	GocdrRoutines[gocdrRoutineInfo.id] = gocdrRoutineInfo

	return gocdrRoutineInfo
}

// setCurrentRoutineToActive will set the id of the current routine to a valid id
// and add it to GocdrRoutine
// If it already contains a valid id, do nothing.
// Call when tracing gets enabled
func setCurrentRoutineToActive() {
	g := getg()

	if g.gocdrRoutineInfo.id != 0 {
		return
	}

	g.gocdrRoutineInfo.id = GetNewGocdrRoutineID()

	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)

	if GocdrRoutines == nil {
		GocdrRoutines = make(map[uint64]*GocdrRoutine)
	}

	GocdrRoutines[g.gocdrRoutineInfo.id] = g.gocdrRoutineInfo
}

// Add an element to the trace of the current routine
// Params:
//   - elem: the element to add
//
// Return:
//   - the index of the element in the trace
func (gi *GocdrRoutine) addToTrace(elem traceElem) int {
	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}

	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *GocdrRoutine) getElement(index int) traceElem {
	return gi.Trace[index]
}

func (gi *GocdrRoutine) getLastElement() traceElem {
	return gi.Trace[len(gi.Trace)-1]
}

// Update an element in the trace of the current routine
// Params:
//   - index: the index of the element to update
//   - elem: the new element
func (gi *GocdrRoutine) updateElement(index int, elem traceElem) {
	if gocdrTracingDisabled {
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

func (gi *GocdrRoutine) GetForkPos() string {
	return gi.forkFile + ":" + intToString(gi.forkLine)
}

// Get the current routine
// Return:
//   - *GocdrRoutine: the current routine
func currentGoRoutineInfo() *GocdrRoutine {
	return getg().gocdrRoutineInfo
}

// GetRoutineID gets the id of the current routine
// Return:
//   - int: id of the current routine, 0 if current routine is nil
func GetRoutineID() int {
	currentRoutine := currentGoRoutineInfo()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.id)
}

// GetReplayRoutineID returns the replay id of the current routine
// Return:
//   - int: replay id of the current routine, 0 if current routine is nil
func GetReplayRoutineID() int {
	currentRoutine := currentGoRoutineInfo()
	if currentRoutine == nil {
		return 0
	}
	return int(currentRoutine.replayID)
}

// GOCDR-FILE-END
