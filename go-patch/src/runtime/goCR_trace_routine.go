// GOCP-FILE_START

// File: goCR_trace_routine.go
// Brief: Functionality for routines
//
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
type GoCRTraceSpawn struct {
	tPost int64
	newID uint64
	file  string
	line  int
}

// Struct to store the termination of a routine
//
// Fields
//   - tPost int64: time when the routine finished
type GoCRTraceRoutineExit struct {
	tPost int64
}

// GoCRSpawnCaller adds a routine spawn to the trace
//
// Parameter:
//   - callerRoutine *GoCRRoutine: routine that created the new routine
//   - newID uint64: id of the new routine
//   - file string: file where the routine was created
//   - line int32: line where the routine was created
func GoCRSpawnCaller(callerRoutine *GoCRRoutine, newID uint64, file string,
	line int32) {
	if goCRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if GoCRIgnore(file) {
		return
	}

	elem := GoCRTraceSpawn{
		tPost: timer,
		newID: newID,
		file:  file,
		line:  int(line),
	}

	callerRoutine.addToTrace(elem)
}

// Record the finish of a routine
func AdvocatRoutineExit() {
	if goCRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	elem := GoCRTraceRoutineExit{
		tPost: timer,
	}
	insertIntoTrace(elem)
}

// Get a string representation of a trace spawn
//
// Returns:
//   - string: the string representation of the form
//     G,[tPost],[newID],[file],[line]
func (elem GoCRTraceSpawn) toString() string {
	return buildTraceElemString("G", elem.tPost, elem.newID, posToString(elem.file, elem.line))
}

// Get a string representation of the routine element
//
// Returns:
//   - string: the string representation of the form
//     E,[tPost]
func (elem GoCRTraceRoutineExit) toString() string {
	return buildTraceElemString("E", elem.tPost)
}

// getOperation is a getter for the spawn element
//
// Returns:
//   - Operation: the operation
func (elem GoCRTraceSpawn) getOperation() Operation {
	return OperationSpawn
}

// getOperation is a getter for the routine exit element
//
// Returns:
//   - Operation: the operation
func (elem GoCRTraceRoutineExit) getOperation() Operation {
	return OperationRoutineExit
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GoCRTraceSpawn) hasFinished() bool {
	return true
}

// hasFinished checks if the operation has finished or is still running/waiting
//
// Returns:
//   - bool: true if its finished, false otherwise
func (elem GoCRTraceRoutineExit) hasFinished() bool {
	return true
}
