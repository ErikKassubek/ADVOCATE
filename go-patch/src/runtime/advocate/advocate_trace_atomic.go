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

const (
	LoadOp     AtomicOp = "L"
	StoreOp    AtomicOp = "S"
	AddOp      AtomicOp = "A"
	SwapOp     AtomicOp = "W"
	CompSwapOp AtomicOp = "C"
)

/*
 * Add an atomic operation to the trace
 * Args:
 * 	index: index of the atomic event in advocateAtomicMap
 */
func AdvocateAtomic[T any](addr *T, op AtomicOp, skip int) {
	timer := GetNextTimeStep()

	_, file, line, _ := Caller(skip)

	if AdvocateIgnore(file) {
		return
	}

	index := pointerAddressAsString(addr, true)

	elem := "A," + uint64ToString(timer) + "," + index + "," + string(op) + "," + file + ":" + intToString(line)
	insertIntoTrace(elem)
}
