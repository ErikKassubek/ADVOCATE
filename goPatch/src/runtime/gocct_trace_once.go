// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_trace_once.go
// Brief: Functionality for once
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a once
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - res GoCCTTraceResource: the resource the op is applied to
//   - suc bool: true if the func in the Do was executed, false otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCCTTraceOnce struct {
	tReq int64
	tCom int64
	res  GoCCTTraceResource
	suc  bool
	file string
	line int
}

// GoCCTOncePre adds a once to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id uint64: id of the once
//
// Returns:
//   - int: index of the operation in the trace
func GoCCTOncePre(mem unsafe.Pointer, id uint64) int {
	if GoCCTTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	if GoCCTIgnore(file) {
		return -1
	}

	res := GoCCTTraceResource{id: id, addr: mem}

	elem := GoCCTTraceOnce{
		tReq: timer,
		res:  res,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// Add the end counter to an operation of the trace
//
// Parameter:
//   - index int: index of the operation in the trace
//   - suc bool: true if the do on the once was called for the first time, false otherwise
func GoCCTOncePost(index int, suc bool) {
	if GoCCTTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}
	elem := currentGoRoutineInfo().getElement(index).(GoCCTTraceOnce)

	elem.tCom = timer
	elem.suc = suc

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     O,[tPre],[tPost],[id],[suc],[file],[line]
func (self GoCCTTraceOnce) toString() string {
	return buildTraceElemString("O", self.tReq, self.tCom, self.res.id, self.suc, posToString(self.file, self.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (self GoCCTTraceOnce) getOperation() Operation {
	return OperationOnceDo
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceOnce) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceOnce) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{self.res}
}
