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

// Struct to store an operation on an atomic variable
//
// Fields
//   - timer int64: time when the operation was executed
//   - id string: id of the atomic, address of the atomic
//   - op Operation: operation type
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceAtomic struct {
	timer int64
	id    string
	op    Operation
	file  string
	line  int
}

// Add an atomic operation to the trace
// Args:
//   - addr *T: memory address of the atomic
//   - op Operation: the operation type
//   - skip iny: skip for Caller
func AdvocateAtomic[T any](addr *T, op Operation, skip int) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(skip)

	if AdvocateIgnore(file) {
		return
	}

	id := pointerAddressAsString(addr, true)

	elem := AdvocateTraceAtomic{
		timer: timer,
		id:    id,
		op:    op,
		file:  file,
		line:  line,
	}

	insertIntoTrace(elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     U,[timer],[id],[operation],[file],[line]
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
		opStr = "N"
	case OperationAtomicOr:
		opStr = "O"
	}

	return buildTraceElemString("A", elem.timer, elem.id, opStr, posToString(elem.file, elem.line))
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem AdvocateTraceAtomic) getOperation() Operation {
	return elem.op
}
