// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_trace_routine.go
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
type GocdrTraceSpawn struct {
	tPost int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type GocdrTraceRoutineExit struct {
	tPost int64
}

// GocdrSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *GocdrRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func GocdrSpawnCaller(callerRoutine *GocdrRoutine, newID uint64, file string,
	line int32) {
	if gocdrTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if GocdrIgnore(file) {
		return
	}

	elem := GocdrTraceSpawn{
		tPost: timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

// Record the finish of a routine
func GocdrRoutineExit() {
	if gocdrTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	elem := GocdrTraceRoutineExit{
		tPost: timer,
	}
	insertIntoTrace(elem)
}

// Get a string representation of a trace spawn
//
// Returns:
//   - string: the string representation of the form
//     G,[tPost],[newID],[file],[line]
func (elem GocdrTraceSpawn) toString() string {
	return buildTraceElemString("G", elem.tPost, elem.newID, posToString(elem.file, elem.line))
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem GocdrTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost)
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (elem GocdrTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem GocdrTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GocdrTraceSpawn) hasFinished() bool {
	return true
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GocdrTraceRoutineExit) hasFinished() bool {
	return true
}
