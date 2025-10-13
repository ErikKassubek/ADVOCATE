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
func areIndependent(tr *trace.Trace, op1, op2 trace.Element) bool {
	return areIndependentType(tr, op1, op2) && concurrent.IsConcurrent(op1, op2)
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
//   - bool: true, if the operations can be switched, false otherwise
func areIndependentType(tr *trace.Trace, op1, op2 trace.Element) bool {
	if op1.GetType(false) != op2.GetType(false) {
		return true
	}

	switch op1.GetType(false) {
	case trace.Atomic:
		return op1.GetType(true) == trace.AtomicLoad && op2.GetType(true) == trace.AtomicLoad
	case trace.Channel:
		// TODO: channel
	case trace.Select:
		// TODO: select
	case trace.Mutex:
		return isIndependentMutex(op1, op2)
	case trace.Wait:
		return isIndependentWait(op1, op2)
	case trace.Cond:
		return op1.GetType(true) == trace.CondBroadcast && op2.GetType(true) == trace.CondBroadcast
	case trace.Once:
		// TODO: once
	case trace.Fork:
		return true
	case trace.Replay, trace.New, trace.End:
		return true
	}

	return false
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
func isIndependentMutex(op1, op2 trace.Element) bool {
	t1 := op1.GetType(true)
	t2 := op2.GetType(true)

	// TODO: RLock -> RUnlock

	if t1 == trace.MutexRLock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true
	} else if t1 == trace.MutexRUnlock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true
	} else if t1 == trace.MutexTryRLock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true
	}

	return false
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
func isIndependentWait(op1, op2 trace.Element) bool {
	t1 := op1.GetType(true)
	t2 := op2.GetType(true)

	if t1 == t2 {
		return true
	}

	// TODO: done, add
	return false
}
