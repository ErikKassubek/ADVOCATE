// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_cond.go
// Brief: Functionality for the conditional variables
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

type AdvocateTraceCond struct {
	tPre  int64
	tPost int64
	id    uint64
	op    Operation
	file  string
	line  int
}

/*
 * AdvocateCondPre adds a cond wait to the trace
 * MARK: Pre
 * Args:
 * 	id: id of the cond
 * 	op: Operation
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateCondPre(id uint64, op Operation) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()
	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceCond{
		tPre: timer,
		id:   id,
		op:   op,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

/*
 * AdvocateCondPost adds the end counter to an operation of the trace
 * MARK: Post
 * Args:
 * 	index: index of the operation in the trace
 */
func AdvocateCondPost(index int) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()
	if index == -1 {
		return
	}
	elem := currentGoRoutine().getElement(index).(AdvocateTraceCond)

	elem.tPost = timer

	currentGoRoutine().updateElement(index, elem)
}

func (elem AdvocateTraceCond) toString() string {
	var opC string
	switch elem.op {
	case OperationCondWait:
		opC = "W"
	case OperationCondSignal:
		opC = "S"
	case OperationCondBroadcast:
		opC = "B"
	}

	return buildTraceElemString("D", elem.tPre, elem.tPost, elem.id, opC, posToString(elem.file, elem.line))
}

func (elem AdvocateTraceCond) getOperation() Operation {
	return elem.op
}
