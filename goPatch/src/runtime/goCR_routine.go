// GOCP-FILE_START

// File: goCR_routine.go
// Brief: Functionality for the routines
//
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

var GoCRRoutines map[uint64]*GoCRRoutine
var GoCRRoutinesLock = mutex{}

var projectPath string

// var atomicRecordingDisabled = false

// GoCRRoutine is a struct to store the trace of a routine
// Fields:
//   - id uint64: the id of the routine
//   - replayID int: when used in reply, id of the new routine in the replayed trace
//   - maxObjectId uint64: the maximum id of elements in the trace
//   - G *g: the g struct of the routine
//   - Trace []traceElem: the trace of the routine
type GoCRRoutine struct {
	id          uint64
	maxObjectId uint64
	G           *g
	Trace       []traceElem
	replayID    int
}

// Create a new goCR routine
// Params:
//   - g: the g struct of the routine
//   - replayRoutine int: when used in reply, id of the new routine in the replayed trace
//   - the replay ids of the routines forked from this routine
//
// Return:
//   - the new goCR routine
func newGoCRRoutine(g *g, replayRoutine int) *GoCRRoutine {
	// ignore the internal routines that are run before the main/test function starts
	if goCRTracingDisabled {
		return &GoCRRoutine{
			id:          0,
			maxObjectId: 0,
			G:           g,
			Trace:       make([]traceElem, 0),
			replayID:    replayRoutine,
		}
	}

	goCRRoutineInfo := &GoCRRoutine{
		id:          GetNewGoCRRoutineID(),
		maxObjectId: 0,
		G:           g,
		Trace:       make([]traceElem, 0),
		replayID:    replayRoutine,
	}

	lock(&GoCRRoutinesLock)
	defer unlock(&GoCRRoutinesLock)

	if GoCRRoutines == nil {
		GoCRRoutines = make(map[uint64]*GoCRRoutine)
	}

	GoCRRoutines[goCRRoutineInfo.id] = goCRRoutineInfo

	return goCRRoutineInfo
}

// setCurrentRoutineToActive will set the id of the current routine to a valid id
// and add it to goCRRoutine
// If it already contains a valid id, do nothing.
// Call when tracing gets enabled
func setCurrentRoutineToActive() {
	g := getg()

	if g.goCRRoutineInfo.id != 0 {
		return
	}

	g.goCRRoutineInfo.id = GetNewGoCRRoutineID()

	lock(&GoCRRoutinesLock)
	defer unlock(&GoCRRoutinesLock)

	if GoCRRoutines == nil {
		GoCRRoutines = make(map[uint64]*GoCRRoutine)
	}

	GoCRRoutines[g.goCRRoutineInfo.id] = g.goCRRoutineInfo
}

// Add an element to the trace of the current routine
// Params:
//   - elem: the element to add
//
// Return:
//   - the index of the element in the trace
func (gi *GoCRRoutine) addToTrace(elem traceElem) int {
	// never needed in actual code, without it the compiler tests fail
	if gi == nil {
		return -1
	}

	gi.Trace = append(gi.Trace, elem)
	return len(gi.Trace) - 1
}

func (gi *GoCRRoutine) getElement(index int) traceElem {
	return gi.Trace[index]
}

func (gi *GoCRRoutine) getLastElement() traceElem {
	return gi.Trace[len(gi.Trace)-1]
}

// Update an element in the trace of the current routine
// Params:
//   - index: the index of the element to update
//   - elem: the new element
func (gi *GoCRRoutine) updateElement(index int, elem traceElem) {
	if goCRTracingDisabled {
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
//   - *GoCRRoutine: the current routine
func currentGoRoutineInfo() *GoCRRoutine {
	return getg().goCRRoutineInfo
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

// GOCP-FILE-END
