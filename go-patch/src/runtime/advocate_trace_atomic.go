// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_atomic.go
// Brief: Functionality for atomics
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

type AdvocateTraceAtomic struct {
	timer uint64
	index string
	op Operation
	file string
	line int
}



/*
 * Add an atomic operation to the trace
 * Args:
 * 	index: index of the atomic event in advocateAtomicMap
 */
func AdvocateAtomic[T any](addr *T, op Operation, skip int) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(skip)

	if AdvocateIgnore(file) {
		return
	}

	index := pointerAddressAsString(addr, true)

	elem := AdvocateTraceAtomic {
		timer: timer,
		index: index,
		op: op,
		file: file,
		line: line,
	}

	insertIntoTrace(elem)
}

func (elem AdvocateTraceAtomic) toString() string {
	opStr := "U"
	switch elem.op {
	case OperationAtomicLoad:
		opStr = "L"
	case OperationAtomicStore:
		opStr = "S"
	case OperationAtomicAdd:
		opStr = "A"
	case OperationAtomicSwap:
		opStr = "W"
	case OperationAtomicCompareAndSwap:
		opStr = "C"
	case OperationAtomicAnd:
		opStr = "M"
	case OperationAtomicOr:
		opStr = "O"
	}

	return buildTraceElemString("A", elem.timer, elem.index, opStr, posToString(elem.file, elem.line))
}

func (elem AdvocateTraceAtomic) getOperation() Operation {
	return elem.op
}