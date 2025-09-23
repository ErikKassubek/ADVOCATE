// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_waitgroup.go
// Brief: Functionality for wait groups
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

// Struct to store an operation on a wait group
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id uint64: id of the mutex
//   - op Operation: operation type
//   - delta int: value by which the internal counter was changed with this operation
//     for Add > 0, for Done -1 and for wait = 0
//   - val int32: value of the internal counter after the operation was executed
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceWaitGroup struct {
	tPre  int64
	tPost int64
	id    uint64
	op    Operation
	delta int
	val   int32
	file  string
	line  int
}

// AdvocateWaitGroupAdd adds a waitgroup add or done to the trace
//
// Parameter:
//   - id: id of the waitgroup
//   - delta: delta of the waitgroup
//   - val: value of the waitgroup after the operation
//
// Returns:
//   - index of the operation in the trace
func AdvocateWaitGroupAdd(id uint64, delta int, val int32) int {
	if advocateTracingDisabled {
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

	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceWaitGroup{
		tPre:  timer,
		id:    id,
		op:    OperationWaitgroupAddDone,
		delta: delta,
		val:   val,
		file:  file,
		line:  line,
	}

	return insertIntoTrace(elem)
}

// AdvocateWaitGroupWait adds a waitgroup wait to the trace
//
// Parameter:
//   - id: id of the waitgroup
//
// Returns:
//   - index of the operation in the trace
func AdvocateWaitGroupWait(id uint64) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipWaitGroupAddWait)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceWaitGroup{
		tPre: timer,
		id:   id,
		op:   OperationWaitgroupWait,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

// AdvocateWaitGroupWaitPost adds the end counter to an operation of the trace
// Wait Post
//
// Parameter:
//   - index: index of the operation in the trace
func AdvocateWaitGroupPost(index int) {
	if advocateTracingDisabled {
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

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceWaitGroup)

	elem.tPost = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     W,[tPre],[tPost],[id],[op],[delta],[val],[file],[line]
func (elem AdvocateTraceWaitGroup) toString() string {
	opStr := "A"
	if elem.op == OperationWaitgroupWait {
		opStr = "W"
	}

	return buildTraceElemString("W", elem.tPre, elem.tPost, elem.id, opStr, elem.delta, elem.val, posToString(elem.file, elem.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceWaitGroup) getOperation() Operation {
	return elem.op
}
