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

// TODO: add to replay?

// var atomicRecordingDisabled = false

// AdvocateTraceFunctionStart is a struct to store the start of a function
// Fields:
//   - t int: time
//   - fileCal string: file of the function call
//   - lineCal int: line of the function call
//   - fileDef string: file of the function definition
//   - lineDef int: line of the function definition
type AdvocateTraceFunctionStart struct {
	t        int64
	fileCall string
	lineCall int
	fileDef  string
	lineDef  int
}

// AdvocateFunctionStart is a struct to store the end of a function
// Fields:
//   - t: int
type AdvocateTraceFunctionReturn struct {
	t int64
}

// AdvocateFunctionCall adds a function stall to the trace
//
// Returns:
//   - index of the operation in the trace
func AdvocateFunctionCall() int { // TODO: not yet called
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	callerSkip := 2 // TODO: set
	_, fileCall, lineCall, _ := Caller(callerSkip + 1)
	_, fileDef, lineDef, _ := Caller(callerSkip)

	if AdvocateIgnore(fileCall) && AdvocateIgnore(fileDef) {
		return -1
	}

	elem := AdvocateTraceFunctionStart{
		t:        timer,
		fileCall: fileCall,
		lineCall: lineCall,
		fileDef:  fileDef,
		lineDef:  lineDef,
	}

	return insertIntoTrace(elem)
}

// AdvocateFunctionCall adds a function stall to the trace
//
// Returns:
//   - index of the operation in the trace
func AdvocateFunctionReturn() int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	elem := AdvocateTraceFunctionReturn{
		t: timer,
	}

	return insertIntoTrace(elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem AdvocateTraceFunctionStart) toString() string {
	return buildTraceElemString("F", elem.t, posToString(elem.fileCall, elem.lineCall), posToString(elem.fileDef, elem.lineDef))
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem AdvocateTraceFunctionReturn) toString() string {
	return buildTraceElemString("R", elem.t)
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceFunctionStart) getOperation() Operation {
	return OperationFunctionCall
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceFunctionReturn) getOperation() Operation {
	return OperationFunctionReturn
}
