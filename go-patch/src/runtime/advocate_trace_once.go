// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_once.go
// Brief: Functionality for once
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause


package runtime

type AdvocateTraceOnce struct {
	tPre uint64
	tPost uint64
	id uint64
	suc bool
	file string
	line int
}

/*
 * AdvocateOncePre adds a once to the trace
 * Args:
 * 	id: id of the once
 * Return:
 * 	index of the operation in the trace
 */
func AdvocateOncePre(id uint64) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(2)

	if AdvocateIgnore(file) {
		return -1
	}

	elem := AdvocateTraceOnce {
		tPre: timer,
		id: id,
		file: file,
		line: line,
	}

	return insertIntoTrace(elem)
}

/*
 * Add the end counter to an operation of the trace
 * Args:
 * 	index: index of the operation in the trace
 * 	suc: true if the do on the once was called for the first time, false otherwise
 */
func AdvocateOncePost(index int, suc bool) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}
	elem := currentGoRoutine().getElement(index).(AdvocateTraceOnce)

	elem.tPost = timer
	elem.suc = suc

	currentGoRoutine().updateElement(index, elem)
}

func (elem AdvocateTraceOnce) toString() string {
	return buildTraceElemString("O", elem.tPre, elem.tPost, elem.id, elem.suc, posToString(elem.file, elem.line))
}

func (elen AdvocateTraceOnce) getOperation() Operation {
	return OperationOnceDo
}