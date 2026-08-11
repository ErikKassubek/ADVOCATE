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
type GoCDRTraceSpawn struct {
	t     int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type GoCDRTraceRoutineExit struct {
	tPost int64
}

// GoCDRSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *GoCDRRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func GoCDRSpawnCaller(callerRoutine *GoCDRRoutine, newID uint64, file string,
	line int32) {
	if GoCDRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if GoCDRIgnore(file) {
		return
	}

	elem := GoCDRTraceSpawn{
		t:     timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

// Record the finish of a routine
func AdvocatRoutineExit() {
	if GoCDRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	elem := GoCDRTraceRoutineExit{
		tPost: timer,
	}
	insertIntoTrace(elem)
	currentGoRoutineInfo().hasReturned = true
}

// Get a string representation of a trace spawn
//
// Returns:
//   - string: the string representation of the form
//     G,[tPost],[newID],[file],[line]
func (self GoCDRTraceSpawn) toString() string {
	return buildTraceElemString("G", self.t, self.newID, posToString(self.file, self.line))
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem GoCDRTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost)
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (self GoCDRTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem GoCDRTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (self GoCDRTraceSpawn) hasFinished() bool {
	return true
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GoCDRTraceRoutineExit) hasFinished() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCDRTraceSpawn) hasCommit() bool {
	return true
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCDRTraceRoutineExit) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCDRTraceResource: recources
func (self GoCDRTraceSpawn) resource() []GoCDRTraceResource {
	return []GoCDRTraceResource{}
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCDRTraceResource: recources
func (self GoCDRTraceRoutineExit) resource() []GoCDRTraceResource {
	return []GoCDRTraceResource{}
}
