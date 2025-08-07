// GOCP-FILE_START

// File: goCR_trace_atomic.go
// Brief: Functionality for atomics
//
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
type GoCRTraceAtomic struct {
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
func GoCRAtomic[T any](addr *T, op Operation, skip int) {
	if goCRTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(skip)

	if GoCRIgnore(file) {
		return
	}

	id := pointerAddressAsString(addr, true)

	elem := GoCRTraceAtomic{
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
func (elem GoCRTraceAtomic) toString() string {
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
func (elem GoCRTraceAtomic) getOperation() Operation {
	return elem.op
}
