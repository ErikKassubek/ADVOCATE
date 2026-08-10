// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_trace_cond.go
// Brief: Functionality for the conditional variables
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a conditional variable
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - res GoCCTTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCCTTraceCond struct {
	tReq int64
	tCom int64
	res  GoCCTTraceResource
	op   Operation
	file string
	line int
}

/*
 * GoCCTCondPre adds a cond wait to the trace
 * Args:
 * 	id: id of the cond
 * 	op: Operation
 * Return:
 * 	index of the operation in the trace
 */
func GoCCTCondPre(mem unsafe.Pointer, id uint64, op Operation) int {
	if GoCCTTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()
	_, file, line, _ := Caller(CallerSkipCond)

	if GoCCTIgnore(file) {
		return -1
	}

	res := GoCCTTraceResource{id: id, addr: mem}

	elem := GoCCTTraceCond{
		tReq: timer,
		res:  res,
		op:   op,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

/*
 * GoCCTCondPost adds the end counter to an operation of the trace
 * Args:
 * 	index: index of the operation in the trace
 */
func GoCCTCondPost(index int) {
	if GoCCTTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	if index == -1 {
		return
	}
	elem := currentGoRoutineInfo().getElement(index).(GoCCTTraceCond)

	elem.tCom = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
func (self GoCCTTraceCond) toString() string {
	var opC string
	switch self.op {
	case OperationCondWait:
		opC = "W"
	case OperationCondSignal:
		opC = "S"
	case OperationCondBroadcast:
		opC = "B"
	}

	return buildTraceElemString("D", self.tReq, self.tCom, self.res.id, opC, posToString(self.file, self.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self GoCCTTraceCond) getOperation() Operation {
	return self.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceCond) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceCond) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{self.res}
}
