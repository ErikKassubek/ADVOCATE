// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_trace_waitgroup.go
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
//   - res GoCDRTraceResource: the resource the op is applied to
//   - op Operation: operation type
//   - delta int: value by which the internal counter was changed with this operation
//     for Add > 0, for Done -1 and for wait = 0
//   - val int32: value of the internal counter after the operation was executed
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GoCDRTraceWaitGroup struct {
	tReq  int64
	tCom  int64
	res   GoCDRTraceResource
	op    Operation
	delta int
	val   int32
	file  string
	line  int
}

// GoCDRWaitGroupAdd adds a waitgroup add or done to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//   - delta: delta of the waitgroup
//   - val: value of the waitgroup after the operation
//
// Returns:
//   - index of the operation in the trace
func GoCDRWaitGroupAdd(mem unsafe.Pointer, id uint64, delta int, val int32) int {
	if GoCDRTracingDisabled {
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

	if GoCDRIgnore(file) {
		return -1
	}

	res := GoCDRTraceResource{id: id, addr: mem}

	elem := GoCDRTraceWaitGroup{
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

// GoCDRWaitGroupWait adds a waitgroup wait to the trace
//
// Parameter:
//   - mem unsafe.Pointer: memory address
//   - id: id of the waitgroup
//
// Returns:
//   - index of the operation in the trace
func GoCDRWaitGroupWait(mem unsafe.Pointer, id uint64) int {
	if GoCDRTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipWaitGroupAddWait)

	if GoCDRIgnore(file) {
		return -1
	}

	res := GoCDRTraceResource{id: id, addr: mem}

	elem := GoCDRTraceWaitGroup{
		tReq: timer,
		res:  res,
		op:   OperationWaitgroupWait,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// GoCDRWaitGroupWaitPost adds the end counter to an operation of the trace
// Wait Post
//
// Parameter:
//   - index: index of the operation in the trace
func GoCDRWaitGroupPost(index int) {
	if GoCDRTracingDisabled {
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

	elem := currentGoRoutineInfo().getElement(index).(GoCDRTraceWaitGroup)

	elem.tCom = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     W,[tPre],[tPost],[id],[op],[delta],[val],[file],[line]
func (elem GoCDRTraceWaitGroup) toString() string {
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
func (elem GoCDRTraceWaitGroup) getOperation() Operation {
	return elem.op
}

// hasCommit returns if the event has committed
//
// Returns:
//   - bool: true if committed, false if only request
func (self GoCDRTraceWaitGroup) hasCommit() bool {
	return self.tCom != 0
}

// resource returns the resources for the operation. Can only be greater 1 for select
//
// Returns:
//   - []GoCDRTraceResource: recources
func (self GoCDRTraceWaitGroup) resource() []GoCDRTraceResource {
	return []GoCDRTraceResource{self.res}
}
