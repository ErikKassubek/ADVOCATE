// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace.go
// Brief: Functionality for the trace
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

type Operation int // enum for operation

const (
	OperationNone Operation = iota
	OperationSpawn
	OperationSpawned
	OperationRoutineExit

	OperationChannelSend
	OperationChannelRecv
	OperationChannelClose

	OperationMutexLock
	OperationMutexUnlock
	OperationMutexTryLock
	OperationRWMutexLock
	OperationRWMutexUnlock
	OperationRWMutexTryLock
	OperationRWMutexRLock
	OperationRWMutexRUnlock
	OperationRWMutexTryRLock

	OperationOnceDo

	OperationWaitgroupAddDone
	OperationWaitgroupWait

	OperationSelect
	OperationSelectCase
	OperationSelectDefault

	OperationCondSignal
	OperationCondBroadcast
	OperationCondWait

	OperationAtomicLoad
	OperationAtomicStore
	OperationAtomicAdd
	OperationAtomicSwap
	OperationAtomicCompareAndSwap
	OperationAtomicAnd
	OperationAtomicOr

	OperationNewChan

	OperationReplayEnd
)

const posSep = "#"

type prePost int // enum for pre/post
const (
	pre prePost = iota
	post
	none
)

var advocateTracingDisabled = true

// var advocateTraceWritingDisabled = false

// Given an Operation enum, return a string representation
//
// Parameter:
//   - op Operation: the operation
//
// Return:
//   - string: the string representation
func getOperationObjectString(op Operation) string {
	switch op {
	case OperationNone:
		return "None"
	case OperationSpawn, OperationSpawned, OperationRoutineExit:
		return "Routine"
	case OperationChannelSend, OperationChannelRecv, OperationChannelClose:
		return "Channel"
	case OperationMutexLock, OperationMutexUnlock, OperationMutexTryLock:
		return "Mutex"
	case OperationRWMutexLock, OperationRWMutexUnlock, OperationRWMutexTryLock, OperationRWMutexRLock, OperationRWMutexRUnlock, OperationRWMutexTryRLock:
		return "RWMutex"
	case OperationOnceDo:
		return "Once"
	case OperationWaitgroupAddDone, OperationWaitgroupWait:
		return "Waitgroup"
	case OperationSelect, OperationSelectCase, OperationSelectDefault:
		return "Select"
	case OperationCondSignal, OperationCondBroadcast, OperationCondWait:
		return "Cond"
	case OperationAtomicLoad, OperationAtomicStore, OperationAtomicAdd, OperationAtomicSwap, OperationAtomicCompareAndSwap, OperationAtomicAnd, OperationAtomicOr:
		return "Atomic"
	case OperationReplayEnd:
		return "Replay"
	}
	return "Unknown"
}

// Interface to define an trace element
type traceElem interface {
	toString() string
	getOperation() Operation
}

// Return a string representation of the trace of the current go routine
//
// Returns:
//   - string representation of the trace
func CurrentTraceToString() string {
	res := ""
	for i, elem := range currentGoRoutineInfo().Trace {
		if i != 0 {
			res += "\n"
		}
		res += elem.toString()
	}

	return res
}

// Add an operation to the trace
//
// Parameter:
//   - elem: element to add to the trace
//
// Returns:
//   - index of the element in the trace
func insertIntoTrace(elem traceElem) int {
	return currentGoRoutineInfo().addToTrace(elem)
}

// Print the trace of the current routines
func PrintTrace() {
	routineID := GetRoutineID()
	println("Routine", routineID, ":", CurrentTraceToString())
}

// Return the trace of the routine by id
//
// Parameter:
//   - id: id of the routine
//
// Returns:
//   - string representation of the trace of the routine
//   - bool: true if the routine exists, false otherwise
func TraceToChanByID(id uint64) chan string {
	println("ADVOTEST 5: ", id)
	lock(&AdvocateRoutinesLock)
	println("LOCK L: 1", id)
	println("ADVOTEST 7: ", id)
	defer println("LOCK U: 1")

	c := make(chan string, 20)
	println("ADVOTEST A: ", id)
	if routine, ok := AdvocateRoutines[id]; ok {
		unlock(&AdvocateRoutinesLock)
		println("ADVOTEST B: ", id)
		println("ADVOTEST 6: ", len(routine.Trace))

		go func() {
			res := ""
			blockSize := 1000
			// if atomic recording is disabled
			for i, elem := range routine.Trace {
				if i != 0 {
					res += "\n"
				}
				res += elem.toString()

				if i%blockSize == 0 {
					c <- res
					res = ""
				}
			}

			if res != "" {
				c <- res
			}

			close(c)
		}()

		println("ADVOTEST C: ", id)
	} else {
		unlock(&AdvocateRoutinesLock)
	}

	println("ADVOTEST 8: ", id)

	return c
}

// Return whether the trace of a routine' is empty
//
// Parameter:
//   - routine: id of the routine
//
// Returns:
//   - true if the trace is empty, false otherwise
func TraceIsEmptyByRoutine(routine int) bool {
	println("LOCK L: 6")
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	defer println("LOCK U: 6")
	if routine, ok := AdvocateRoutines[uint64(routine)]; ok {
		return len(routine.Trace) == 0
	}
	return true
}

// Given a list of element, return a string representation of the elements
// separated by ,
//
// Parameter:
//   - values ...any: the elements
//
// Returns:
//   - a concatenated string representation of all values separated by ,
func buildTraceElemString(values ...any) string {
	return buildTraceElemStringSep(",", values...)
}

// Given a list of element, return a string representation of the elements
// separated by a given separator
//
// Parameter:
//   - values ...any: the elements
//   - sep string: the separator
//
// Returns:
//   - a concatenated string representation of all values separated by the separator
func buildTraceElemStringSep(sep string, values ...any) string {
	res := ""
	for i, v := range values {
		if i != 0 {
			res += sep
		}

		res += convToString(v)
	}
	return res
}

// GetNumberOfRoutines returns the number of routines in the trace
//
// Returns:
//   - number of routines in the trace
func GetNumberOfRoutines() int {
	println("LOCK L: 7")
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	println("LOCK U: 7")
	return len(AdvocateRoutines)
}

// DeleteTrace removes all trace elements from the trace
// It does not remove the routine objects them self
// Make sure to call BlockTrace(), before calling this function
func DeleteTrace() {
	lock(&AdvocateRoutinesLock)
	println("LOCK L: 8")
	defer unlock(&AdvocateRoutinesLock)
	defer println("LOCK U: W")
	for i := range AdvocateRoutines {
		AdvocateRoutines[i].Trace = AdvocateRoutines[i].Trace[:0]
	}
}

// We are only interested in the behaviour of the actual program, not the details
// of the internal implementation.
// Additionally, some operations, like garbage collection and internal operations, can
// cause the replay to get stuck or are not needed.
// For this reason, we ignore all internal operations
//
// Parameter:
//   - file: file in which the operation is executed
//
// Returns:
//   - bool: true if the operation should be ignored, false otherwise
func AdvocateIgnore(file string) bool {
	return (containsStr(file, "goPatch/src/") || containsStr(file, "go/pkg/mod")) &&
		!containsStr(file, "goPatch/src/time/tick.go") &&
		!containsStr(file, "goPatch/src/context/context.go")
}

func RemoveActive(id uint64) {
	lock(&AdvocateRoutinesLock)
	println("LOCK L: 9")
	defer unlock(&AdvocateRoutinesLock)
	defer println("LOCK U: 1")

	delete(AdvocateRoutines, uint64(id))
}

// IsActive returns if a routine of the given id has been created/started but not yet been written to file, and if it exists, if the writing of the trace has started
//
// Parameter:
//   - id int: routine id
//
// Returns:
//   - bool: true if not started or written to file
func IsActive(id int) (bool, bool) {
	lock(&AdvocateRoutinesLock)
	println("LOCK L: 10")
	defer unlock(&AdvocateRoutinesLock)
	defer println("LOCK U: 10")

	if g, ok := AdvocateRoutines[uint64(id)]; ok {
		return ok, g.startedWritingToFile
	}

	return false, false
}

// Write the trace of the current routine to file. After writing, remove from active
//
// Parameter:
//   - id int: routine id
//
// Returns:
//   - bool: true if not started or written to file
func AdvocateWriteTraceToFile() {
	if advocateTracingDisabled {
		return
	}

	g := currentGoRoutineInfo()

	if g == nil {
		return
	}

	if AdvocateIgnore(g.createdAtFile) {
		return
	}

	g.startedWritingToFile = true
	ok := writeTraceToFileFunc(int(g.id), true)
	if !ok { // writing from finishTracing has already started
		return
	}

	RemoveActive(g.id)
}

// ADVOCATE-FILE-END
