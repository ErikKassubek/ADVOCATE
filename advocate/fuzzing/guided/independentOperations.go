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

type independenceCondition int

// Condition that must be met when the first operation is executed
const (
	none                   independenceCondition = iota
	closeBefore                                  // there must be a close before the operations
	mutexRWCounterAtLeast2                       // the rw counter must be 2 or greater
	waitCounterAtLeast2                          // the wg counter must be 2 or greater
	onceBefore                                   // there must be a once.Do before

)

// areIndependent checks if two operations are independent.
// To operations are independent if changing the order of those operations in
// the trace cannot transform the execution of the trace from one not containing
// any concurrency bugs to one, that contains them.
// Non-Concurrent operations are always concurrent

// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - independenceCondition: condition that must be true for the two operations to be independent
func areIndependent(op1, op2 trace.Element) (bool, independenceCondition) {
	if !concurrent.IsConcurrent(op1, op2) {
		return false, none
	}
	return areIndependentType(op1, op2)
}

// areIndependentType checks if two operations are independent based on there type.
// To operations are independent if changing the order of those operations in
// the trace cannot transform the execution of the trace from one not containing
// any concurrency bugs to one, that contains them.
// In this function, we do not care if the elements are concurrent or not.

// Parameter:
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
//
// Returns:
//   - bool: true, if the operations are independent, false otherwise
//   - independenceCondition: condition that must be true for the two operations to be independent
func areIndependentType(op1, op2 trace.Element) (bool, independenceCondition) {
	t := op1.GetType(false)
	// TODO: allow channel/select and select/select pairs
	if !op1.IsSameElement(op2) && !(t == trace.Channel || t == trace.Select) {
		return true, none
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
		return isIndependentOnce()
	case trace.Fork:
		return true, none
	case trace.Replay, trace.New, trace.End:
		return true, none
	}

	return false, none
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
//   - independenceCondition: condition that must be true for the two operations to be independent
func isIndependentAtomic(op1, op2 trace.Element) (bool, independenceCondition) {
	return (op1.GetType(true) == trace.AtomicLoad && op2.GetType(true) == trace.AtomicLoad), none
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
//   - independenceCondition: condition that must be true for the two operations to be independent
func isIndependentChannelSelect(op1, op2 trace.Element) (bool, independenceCondition) {
	op1Type := op1.GetType(false)
	op2Type := op2.GetType(false)

	if op1Type == trace.Channel && op2Type == trace.Channel {
		op1Op := op1.GetType(true)
		op2Op := op2.GetType(true)
		if op1Op == trace.ChannelRecv && op2Op == trace.ChannelRecv {
			return true, closeBefore
		}
		return false, none
	} else if op1Type == trace.Select && op2Type == trace.Select {
		if !op1.(*trace.ElementSelect).HasCommonChannel(op2.(*trace.ElementSelect)) {
			return true, none
		}
	} else { // one select and one channel
		if op1Type == trace.Channel { // op1 -> channel, op2 -> select
			if !op2.(*trace.ElementSelect).IsInCases(op1.(*trace.ElementChannel)) {
				return true, none
			}
		} else { // op1 -> select, op2 -> channel
			if !op1.(*trace.ElementSelect).IsInCases(op2.(*trace.ElementChannel)) {
				return true, none
			}
		}
	}

	return false, none
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
//   - independenceCondition: condition that must be true for the two operations to be independent
func isIndependentMutex(op1, op2 trace.Element) (bool, independenceCondition) {
	t1 := op1.GetType(true)
	t2 := op2.GetType(true)

	if t1 == trace.MutexRLock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true, none
	} else if t1 == trace.MutexRUnlock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true, none
	} else if t1 == trace.MutexTryRLock && (t2 == trace.MutexRLock || t2 == trace.MutexTryRLock) {
		return true, none
	} else if t1 == trace.MutexRLock && t2 == trace.MutexRUnlock {
		return true, mutexRWCounterAtLeast2
	}

	return false, none
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
//   - independenceCondition: condition that must be true for the two operations to be independent
func isIndependentWait(op1, op2 trace.Element) (bool, independenceCondition) {
	t1 := op1.GetType(true)
	t2 := op2.GetType(true)

	if t1 == t2 {
		return true, none
	} else if (t1 == trace.WaitAdd && t2 == trace.WaitDone) || (t1 == trace.WaitDone && t2 == trace.WaitAdd) {
		return true, waitCounterAtLeast2
	}

	return false, none
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
//   - independenceCondition: condition that must be true for the two operations to be independent
func isIndependentCond(op1, op2 trace.Element) (bool, independenceCondition) {
	return (op1.GetType(true) == trace.CondBroadcast && op2.GetType(true) == trace.CondBroadcast), none
}

// areIndependentCond checks if two once operations are independent based
// on there operation type
//
// Returns:
//   - bool: true, if the operations can be switched, false otherwise
//   - independenceCondition: condition that must be true for the two operations to be independent
func isIndependentOnce() (bool, independenceCondition) {
	return true, onceBefore
}
