// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_trace_waitgroup.go
// Brief: Functionality for wait groups
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

import "unsafe"

// Struct to store an operation on a wait group
//
// Fields
//   - tReq int64: time when the operation started
//   - tCom int64: time when the operation finished
//   - res GoCCTTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - delta int: value by which the internal counter was changed with this operation
//     for Add > 0, for Done -1 and for wait = 0
//   - val int32: value of the internal counter after the operation was executed
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCCTTraceWaitGroup struct {
	tReq  int64
	tCom  int64
	res   GoCCTTraceResource
	op    Operation
	delta int
	val   int32
	file  string
	line  int
}

// GoCCTWaitGroupAdd adds a waitgroup add or done to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//   - delta: delta of the waitgroup
//   - val: value of the waitgroup after the operation
//
// Returns:
//   - index of the operation in the trace
func GoCCTWaitGroupAdd(mem unsafe.Pointer, id uint64, delta int, val int32) int {
	if GoCCTTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	var file string
	var line int
	if delta > 0 {
		_, file, line, _ = Caller(CallerSkipWaitGroupAddWait)
	} else {
		_, file, line, _ = Caller(CallerSkipWaitGroupDone)
	}

	if GoCCTIgnore(file) {
		return -1
	}

	res := GoCCTTraceResource{id: id, addr: mem}

	elem := GoCCTTraceWaitGroup{
		tReq:  timer,
		op:    OperationWaitgroupAddDone,
		res:   res,
		delta: delta,
		val:   val,
		file:  file,
		line:  line,
	}

	return insertIntoTrace(elem)
}

// GoCCTWaitGroupWait adds a waitgroup wait to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//
// Returns:
//   - index of the operation in the trace
func GoCCTWaitGroupWait(mem unsafe.Pointer, id uint64) int {
	if GoCCTTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipWaitGroupAddWait)

	if GoCCTIgnore(file) {
		return -1
	}

	res := GoCCTTraceResource{id: id, addr: mem}

	elem := GoCCTTraceWaitGroup{
		tReq: timer,
		res:  res,
		op:   OperationWaitgroupWait,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// GoCCTWaitGroupWaitPost adds the end counter to an operation of the trace
// Wait Post
//
// Parameter:
//   - index: index of the operation in the trace
func GoCCTWaitGroupPost(index int) {
	if GoCCTTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	// internal elements are not in the trace
	if index == -1 {
		return
	}

	// only needed to fix tests

	if currentGoRoutineInfo() == nil {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(GoCCTTraceWaitGroup)

	elem.tCom = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     W,[tPre],[tPost],[id],[op],[delta],[val],[file],[line]
func (elem GoCCTTraceWaitGroup) toString() string {
	opStr := "A"
	if elem.op == OperationWaitgroupWait {
		opStr = "W"
	}

	return buildTraceElemString("W", elem.tReq, elem.tCom, elem.res.id, opStr, elem.delta, elem.val, posToString(elem.file, elem.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem GoCCTTraceWaitGroup) getOperation() Operation {
	return elem.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCCTTraceWaitGroup) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCCTTraceResource: recources
func (self GoCCTTraceWaitGroup) resource() []GoCCTTraceResource {
	return []GoCCTTraceResource{self.res}
}
