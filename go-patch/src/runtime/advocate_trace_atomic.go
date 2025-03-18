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

type AtomicOp string

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

	opStr := "U"
	switch op {
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

	elem := "A," + uint64ToString(timer) + "," + index + "," + opStr + "," + file + ":" + intToString(line)
	insertIntoTrace(elem)
}
