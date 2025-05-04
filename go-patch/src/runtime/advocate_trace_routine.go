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

// Struct to store the spawn of a new routine (go func() {}())
//
// Fields
//   - tPost int64: time when the operation finished
//   - newID uint64: id of new routine
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceSpawn struct {
	tPost int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type AdvocateTraceRoutineExit struct {
	tPost int64
}

// AdvocateSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *AdvocateRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func AdvocateSpawnCaller(callerRoutine *AdvocateRoutine, newID uint64, file string,
	line int32) {
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

// Record the finish of a routine
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

// Get a string representation of a trace spawn
//
// Returns:
//   - string: the string representation of the form
//     G,[tPost],[newID],[file],[line]
func (elem AdvocateTraceSpawn) toString() string {
	return buildTraceElemString("G", elem.tPost, elem.newID, posToString(elem.file, elem.line))
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem AdvocateTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost)
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}
