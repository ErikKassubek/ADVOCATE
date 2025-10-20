// Copyright (c) 2025 Erik Kassubek
//
// File: independentTraces.go
// Brief: Some elements in traces can be swapped without changing a program
//    execution regarding concurrency bugs. Functions in this file check
//    if two operations fulfill the requirements for this.
//
// Author: Erik Kassubek
// Created: 2025-10-10
//
// License: BSD-3-Clause

package guided

import (
	"advocate/analysis/hb/concurrent"
	"advocate/trace"
)

// areIndependent checks if two operations are independent.
// To operations are independent if changing the order of those operations in
// the trace cannot transform the execution of the trace from one not containing
// any concurrency bugs to one, that contains them.
// Non-Concurrent operations are always concurrent

// Parameter:
//   - tr *trace.Trace: the trace
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must happen before op1 and op2 for them to be independent
func areIndependent(tr *trace.Trace, op1, op2 trace.Element) (bool, []trace.Element) {
	if !concurrent.IsConcurrent(op1, op2) {
		return false, make([]trace.Element, 0)
	}
	return areIndependentType(tr, op1, op2)
}

// areIndependentType checks if two operations are independent based on there type.
// To operations are independent if changing the order of those operations in
// the trace cannot transform the execution of the trace from one not containing
// any concurrency bugs to one, that contains them.
// In this function, we do not care if the elements are concurrent or not.

// Parameter:
//   - tr *trace.Trace: the trace
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations are independent, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func areIndependentType(tr *trace.Trace, op1, op2 trace.Element) (bool, []trace.Element) {
	// TODO: allow channel/select and select/select pairs
	if !op1.IsSameElement(op2) {
		return true, make([]trace.Element, 0)
	}

	switch op1.GetType(false) {
	case trace.Atomic:
		return isIndependentAtomic(op1, op2)
	case trace.Channel, trace.Select:
		return isIndependentChannelSelect(op1, op2)
	case trace.Mutex:
		return isIndependentMutex(op1, op2)
	case trace.Wait:
		return isIndependentWait(op1, op2)
	case trace.Cond:
		return isIndependentCond(op1, op2)
	case trace.Once:
		return isIndependentOnce(op1, op2)
	case trace.Fork:
		return true, make([]trace.Element, 0)
	case trace.Replay, trace.New, trace.End:
		return true, make([]trace.Element, 0)
	}

	return false, make([]trace.Element, 0)
}

// areIndependentAtomics checks if two atomic operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentAtomic(op1, op2 trace.Element) (bool, []trace.Element) {
	return (op1.GetType(true) == trace.AtomicLoad && op2.GetType(true) == trace.AtomicLoad), make([]trace.Element, 0)
}

// areIndependentAtomics checks if two channel or select operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentChannelSelect(op1, op2 trace.Element) (bool, []trace.Element) {
	// TODO: implement
	return false, make([]trace.Element, 0)
}

// areIndependentAtomics checks if two atomic operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentChannel(op1, op2 trace.Element) (bool, []trace.Element) {
	return (op1.GetType(true) == trace.AtomicLoad && op2.GetType(true) == trace.AtomicLoad), make([]trace.Element, 0)
}

// areIndependentMutex checks if two mutex operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentMutex(op1, op2 trace.Element) (bool, []trace.Element) {
	t1 := op1.GetType(true)
	t2 := op2.GetType(true)

	// TODO: RLock -> RUnlock

	if t1 == trace.MutexRLock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true, make([]trace.Element, 0)
	} else if t1 == trace.MutexRUnlock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true, make([]trace.Element, 0)
	} else if t1 == trace.MutexTryRLock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true, make([]trace.Element, 0)
	}

	return false, make([]trace.Element, 0)
}

// areIndependentMutex checks if two mutex operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentWait(op1, op2 trace.Element) (bool, []trace.Element) {
	t1 := op1.GetType(true)
	t2 := op2.GetType(true)

	if t1 == t2 {
		return true, make([]trace.Element, 0)
	}

	// TODO: done, add
	return false, make([]trace.Element, 0)
}

// areIndependentCond checks if two conditional operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentCond(op1, op2 trace.Element) (bool, []trace.Element) {
	return (op1.GetType(true) == trace.CondBroadcast && op2.GetType(true) == trace.CondBroadcast), make([]trace.Element, 0)
}

// areIndependentCond checks if two once operations are independent based
// on there operation type
//
// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - []trace.Element: elements that must be executed before op1 and op2 for them to be independent
func isIndependentOnce(op1, op2 trace.Element) (bool, []trace.Element) {
	// TODO
	return false, make([]trace.Element, 0)
}
