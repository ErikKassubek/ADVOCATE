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
//   - name string: name of the function
//   - fileCal string: file of the function call
//   - lineCal int: line of the function call
//   - fileDef string: file of the function definition
//   - lineDef int: line of the function definition
type AdvocateTraceFunctionStart struct {
	t        int64
	name     string
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
func advocateFunctionCall() {
	// println("A")
	if advocateTracingDisabled {
		return
	}

	// move to slow required because advocateFunctionCall is handled in the compiler
	timer := GetNextTimeStep()

	callerSkip := 1
	_, fileCall, lineCall, _ := Caller(callerSkip + 1)
	pc, fileDef, lineDef, _ := Caller(callerSkip)

	if AdvocateIgnore(fileCall) && AdvocateIgnore(fileDef) {
		return
	}

	funcName := FuncForPC(pc).Name()

	elem := AdvocateTraceFunctionStart{
		t:        timer,
		name:     funcName,
		fileCall: fileCall,
		lineCall: lineCall,
		fileDef:  fileDef,
		lineDef:  lineDef,
	}

	insertIntoTrace(elem)
}

// AdvocateFunctionCall adds a function stall to the trace
//
// Returns:
//   - index of the operation in the trace
func advocateFunctionReturn() {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	elem := AdvocateTraceFunctionReturn{
		t: timer,
	}

	insertIntoTrace(elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (elem AdvocateTraceFunctionStart) toString() string {
	return buildTraceElemString("F", elem.t, elem.name, posToString(elem.fileDef, elem.lineDef), posToString(elem.fileCall, elem.lineCall))
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
