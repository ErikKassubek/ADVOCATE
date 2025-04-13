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

/*
 * AdvocateWaitGroupAdd adds a waitgroup add or done to the trace
 * MARK: Add
 * Args:
 * 	id: id of the waitgroup
 *  delta: delta of the waitgroup
 * 	val: value of the waitgroup after the operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateWaitGroupAdd(id uint64, delta int, val int32) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	var file string
	var line int
	if delta > 0 {
		_, file, line, _ = Caller(2)
	} else {
		_, file, line, _ = Caller(3)
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

/*
 * AdvocateWaitGroupWait adds a waitgroup wait to the trace
 * MARK: Wait Pre
 * Args:
 * 	id: id of the waitgroup
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateWaitGroupWait(id uint64) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

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

/*
 * AdvocateWaitGroupWaitPost adds the end counter to an operation of the trace
 * MARK: Wait Post
 * Args:
 * 	index: index of the operation in the trace
 */
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

	if currentGoRoutine() == nil {
		return
	}

	elem := currentGoRoutine().getElement(index).(AdvocateTraceWaitGroup)

	elem.tPost = timer

	currentGoRoutine().updateElement(index, elem)
}

func (elem AdvocateTraceWaitGroup) toString() string {
	opStr := "A"
	if elem.op == OperationWaitgroupWait {
		opStr = "W"
	}

	return buildTraceElemString("W", elem.tPre, elem.tPost, elem.id, opStr, elem.delta, elem.val, posToString(elem.file, elem.line))
}

func (elem AdvocateTraceWaitGroup) getOperation() Operation {
	return elem.op
}
