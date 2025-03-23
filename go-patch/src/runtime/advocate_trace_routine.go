// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_routine.go
// Brief: Functionality for routines
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

type AdvocateTraceSpawn struct {
	tPost int64
	newID uint64
	file  string
	line  int
}

type AdvocateTraceRoutineExit struct {
	tPost int64
}

/*
 * AdvocateSpawnCaller adds a routine spawn to the trace
 * Args:
 * 	callerRoutine: routine that created the new routine
 * 	newID: id of the new routine
 * 	file: file where the routine was created
 * 	line: line where the routine was created
 */
func AdvocateSpawnCaller(callerRoutine *AdvocateRoutine, newID uint64, file string, line int32) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if AdvocateIgnore(file) {
		return
	}

	elem := AdvocateTraceSpawn{
		tPost: timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

/*
 * Record the finish of a routine
 */
func AdvocatRoutineExit() {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	elem := AdvocateTraceRoutineExit{
		tPost: timer,
	}
	insertIntoTrace(elem)
}

func (elem AdvocateTraceSpawn) toString() string {
	return buildTraceElemString("G", elem.tPost, elem.newID, posToString(elem.file, elem.line))
}

func (elem AdvocateTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost)
}

func (elem AdvocateTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

func (elem AdvocateTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}
