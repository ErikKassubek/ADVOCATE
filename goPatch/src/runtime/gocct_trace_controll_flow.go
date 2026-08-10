// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: trace_conditional.go
// Brief: Functionality for recording if and switch results
//
// Author: Erik Kassubek
// Created: 2026-07-27
//
// License: BSD-3-Clause

package runtime

// Struct to store and if or switch
//
// Fields
//   - t int64: time
//   - numCases int: number of cases
//   - chosenCase int: chosen case in order given in code
//   - ct condType: if or switch
//   - file string: file
//   - line int: first line
type GoCCTTraceControllFlow struct {
	t          int64
	numCases   int
	chosenCase int
	ct         Operation
	file       string
	line       int
}

// TODO: fix line
// GoCCTControllFlow inserts a conditional into the trace
//
// Parameter:
//   - ct string: "I" for if, "S" for switch
//   - numCases int: number of cases
//   - chosenCase int: number of chosen case
func gocctControllFlow(ct string, numCases, chosenCase int) {
	if GoCCTTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(1)

	if GoCCTIgnore(file) {
		return
	}

	var op Operation
	switch ct {
	case "I":
		op = OperationControllIf
	case "S":
		op = OperationControllSwitch
	default:
		panic("Unknown Controll Flow type " + ct)
	}

	elem := GoCCTTraceControllFlow{
		t:          timer,
		numCases:   numCases,
		chosenCase: chosenCase,
		ct:         op,
		file:       file,
		line:       line,
	}

	insertIntoTrace(elem)
}

func (self GoCCTTraceControllFlow) toString() string {
	ctString := ""
	switch self.ct {
	case OperationControllIf:
		ctString = "I"
	case OperationControllSwitch:
		ctString = "S"
	default:
		panic("Invalid controll flow type: " + string(self.ct))
	}

	return buildTraceElemString("I", self.t, ctString, self.numCases, self.chosenCase, posToString(self.file, self.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self GoCCTTraceControllFlow) getOperation() Operation {
	return self.ct
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceControllFlow) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceControllFlow) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{}
}
