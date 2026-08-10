// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_routine.go
// Brief: Functionality for the routines
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

var GoCCTRoutines map[uint64]*GoCCTRoutine
var GoCCTRoutinesLock = mutex{}

var projectPath string

// var atomicRecordingDisabled = false

// GoCCTRoutine is a struct to store the trace of a routine
// Fields:
//   - id uint64: the id of the routine
//   - maxObjectId uint64: the maximum id of elements in the trace
//   - G *g: the g struct of the routine
//   - Trace []traceElem: the trace of the routine
//   - oat []uint64: object aware trace elements, ids of blocked object the routine can access
//   - replayID int: when used in reply, id of the new routine in the replayed trace
//   - forkFile string: file where the routine was created in, "main" for main routine
//   - forkLine int: line where ther routine was created in, 0 for main routine
//   - parkOn []unsafe.Pointer: list of elements the routine was last parked on
//   - parkPos string: position of last park in form file:line
//   - parkForeverReplay bool: if true, routine parks forever based on replay
//   - wokenByTimeout bool: in replay block was woken up by timeout
//   - hasReturned bool: true if the routine has terminated
type GoCCTRoutine struct {
	id                   uint64
	maxObjectId          uint64
	G                    *g
	Trace                []traceElem
	oat                  []uint64
	replayID             int
	forkFile             string
	forkLine             int32
	parkForeverReplay    bool
	hasReturned          bool
	wokenButTimeout      bool
	wokenNoTimeout       bool
	startedWritingToFile bool
}

// Create a new gocct routine
// Params:
//   - g: the g struct of the routine
//   - replayRoutine int: when used in reply, id of the new routine in the replayed trace the replay ids of the routines forked from this routine
//   - file string: file, where the routine was created
//   - line int32: line, where the routine was created
//
// Return:
//   - the new gocct routine
func newGoCCTRoutine(g *g, replayRoutine int, file string, line int32) *GoCCTRoutine {
	// ignore the internal routines that are run before the main/test function starts
	if GoCCTTracingDisabled {
		return &GoCCTRoutine{
			id:          0,
			maxObjectId: 0,
			G:           g,
			Trace:       make([]traceElem, 0),
			forkFile:    file,
			forkLine:    line,
			replayID:    replayRoutine,
		}
	}

	gocctRoutineInfo := &GoCCTRoutine{
		id:          GetNewGoCCTRoutineID(),
		maxObjectId: 0,
		G:           g,
		Trace:       make([]traceElem, 0),
		replayID:    replayRoutine,
		forkFile:    file,
		forkLine:    line,
	}

	lock(&GoCCTRoutinesLock)
	defer unlock(&GoCCTRoutinesLock)

	if GoCCTRoutines == nil {
		GoCCTRoutines = make(map[uint64]*GoCCTRoutine)
	}

	GoCCTRoutines[gocctRoutineInfo.id] = gocctRoutineInfo

	return gocctRoutineInfo
}

// setCurrentRoutineToActive will set the id of the current routine to a valid id
// and add it to GoCCTRoutine
// If it already contains a valid id, do nothing.
// Call when tracing gets enabled
func setCurrentRoutineToActive() {
	g := getg()

	if g.gocctRoutineInfo.id != 0 {
		return
	}

	g.gocctRoutineInfo.id = GetNewGoCCTRoutineID()

	lock(&GoCCTRoutinesLock)
	defer unlock(&GoCCTRoutinesLock)

	if GoCCTRoutines == nil {
		GoCCTRoutines = make(map[uint64]*GoCCTRoutine)
	}

	GoCCTRoutines[g.gocctRoutineInfo.id] = g.gocctRoutineInfo
}

// Add an element to the trace of the current routine
// Params:
//   - elem: the element to add
//
// Return:
//   - the index of the element in the trace
func (gi *GoCCTRoutine) addToTrace(elem traceElem) int {
	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}

	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *GoCCTRoutine) getElement(index int) traceElem {
	return gi.Trace[index]
}

func (gi *GoCCTRoutine) getPosCreated() string {
	return posToString(gi.forkFile, int(int(gi.forkLine)))
}

func (gi *GoCCTRoutine) getLastElement() traceElem {
	return gi.Trace[len(gi.Trace)-1]
}

func (gi *GoCCTRoutine) GetForkPos() string {
	return posToString(gi.forkFile, int(gi.forkLine))
}

// Update an element in the trace of the current routine
// Params:
//   - index: the index of the element to update
//   - elem: the new element
func (gi *GoCCTRoutine) updateElement(index int, elem traceElem) {
	if GoCCTTracingDisabled {
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
//   - *GoCCTRoutine: the current routine
func currentGoRoutineInfo() *GoCCTRoutine {
	return getg().gocctRoutineInfo
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

// GOCCT-FILE-END
