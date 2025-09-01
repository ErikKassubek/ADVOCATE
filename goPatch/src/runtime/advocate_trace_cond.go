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

// Struct to store an operation on a conditional variable
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id string: id of the channel
//   - op Operation: operation type
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
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
	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceCond)

	elem.tPost = timer

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation
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

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceCond) getOperation() Operation {
	return elem.op
}
