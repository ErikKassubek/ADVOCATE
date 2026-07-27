// ADVOCATE-FILE_START

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
type AdvocateTraceControllFlow struct {
	t          int64
	numCases   int
	chosenCase int
	ct         Operation
	file       string
	line       int
}

// TODO: fix line
// AdvocateControllFlow inserts a conditional into the trace
//
// Parameter:
//   - ct string: "I" for if, "S" for switch
//   - numCases int: number of cases
//   - chosenCase int: number of chosen case
func advocateControllFlow(ct string, numCases, chosenCase int) {
	return
	// if AdvocateTracingDisabled {
	// 	return -1
	// }

	// timer := GetNextTimeStep()

	// _, file, line, _ := Caller(1)

	// if AdvocateIgnore(file) {
	// 	return -1
	// }

	// var op Operation
	// switch ct {
	// case "I":
	// 	op = OperationIf
	// case "S":
	// 	op = OperationSelect
	// default:
	// 	panic("Unknown Controll Flow type " + ct)
	// }

	// elem := AdvocateTraceControllFlow{
	// 	t:          timer,
	// 	numCases:   numCases,
	// 	chosenCase: chosenCase,
	// 	ct:         op,
	// 	file:       file,
	// 	line:       line,
	// }

	// return insertIntoTrace(elem)
}

func (self AdvocateTraceControllFlow) toString() string {
	return buildTraceElemString("I", self.t, string(self.ct), self.numCases, self.chosenCase, posToString(self.file, self.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self AdvocateTraceControllFlow) getOperation() Operation {
	return self.ct
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self AdvocateTraceControllFlow) hasCommit() bool {
	return true
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []AdvocateTraceResource: recources
func (self AdvocateTraceControllFlow) resource() []AdvocateTraceResource {
	return []AdvocateTraceResource{}
}
