// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_trace.go
// Brief: Functionality for the trace
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

type Operation string // enum for operation

const (
	OperationNone        Operation = "none"
	OperationSpawn       Operation = "routineSpawn"
	OperationSpawned     Operation = "routineSpawned"
	OperationRoutineExit Operation = "routineExit"

	OperationChannelSend  Operation = "chanSend"
	OperationChannelRecv  Operation = "chanRecv"
	OperationChannelClose Operation = "chanClose"

	OperationMutexLock       Operation = "mutexLock"
	OperationMutexUnlock     Operation = "mutexUnlock"
	OperationMutexTryLock    Operation = "mutexTrylock"
	OperationRWMutexLock     Operation = "rwmutexLock"
	OperationRWMutexUnlock   Operation = "rwmutexUnlock"
	OperationRWMutexTryLock  Operation = "rwmutexTrylock"
	OperationRWMutexRLock    Operation = "rwmutexRlock"
	OperationRWMutexRUnlock  Operation = "rwmutexrunlock"
	OperationRWMutexTryRLock Operation = "rwmutexTryrlock"

	OperationOnceDo Operation = "onceDo"

	OperationWaitgroupAddDone Operation = "wgAdddone"
	OperationWaitgroupWait    Operation = "wgWait"

	OperationSelect        Operation = "wgSelect"
	OperationSelectCase    Operation = "wgSelectcase"
	OperationSelectDefault Operation = "wgSelectdefault"

	OperationCondSignal    Operation = "condSignal"
	OperationCondBroadcast Operation = "condBroadcast"
	OperationCondWait      Operation = "condWait"

	OperationAtomicLoad           Operation = "atomicLoad"
	OperationAtomicStore          Operation = "atoicStore"
	OperationAtomicAdd            Operation = "atomicAdd"
	OperationAtomicSwap           Operation = "atomicSwap"
	OperationAtomicCompareAndSwap Operation = "atomicCompareandswap"
	OperationAtomicAnd            Operation = "atomicAnd"
	OperationAtomicOr             Operation = "atomicOr"

	OperationNewChan Operation = "newChan"

	OperationFunctionCall   Operation = "funcCall"
	OperationFunctionReturn Operation = "funcReturn"

	OperationReplayNever Operation = "replayNever"
	OperationReplayEnd   Operation = "replayEnd"
)

const posSep = "#"

type prePost int // enum for pre/post
const (
	pre prePost = iota
	post
	none
)

var gocdrTracingDisabled = true

// var gocdrTraceWritingDisabled = false

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

// Return a string representation of a given routine local trace
//
// Parameter:
//   - trace: trace to convert to string
//
// Returns:
//   - string: string representation of the trace
func traceToString(trace *[]traceElem) string {
	res := ""

	// if atomic recording is disabled
	println("START TTS: ", len(*trace))
	for i, elem := range *trace {
		if i != 0 {
			res += "\n"
		}
		res += elem.toString()
	}
	println("END TTS")
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
func TraceToStringByID(id uint64) (string, bool) {
	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)
	if routine, ok := GocdrRoutines[id]; ok {
		return traceToString(&routine.Trace), true
	}
	return "", false
}

// Return whether the trace of a routine' is empty
//
// Parameter:
//   - routine: id of the routine
//
// Returns:
//   - true if the trace is empty, false otherwise
func TraceIsEmptyByRoutine(routine int) bool {
	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)
	if routine, ok := GocdrRoutines[uint64(routine)]; ok {
		return len(routine.Trace) == 0
	}
	return true
}

// Get the trace of the routine with id 'id'.
// To minimized the needed ram the trace is sent to the channel 'c' in chunks
// of 1000 elements.
//
// Parameter:
//   - id: id of the routine
//   - c: channel to send the trace to
//   - atomic: it true, the atomic trace is returned
func TraceToStringByIDChannel(id int, c chan<- string) {
	lock(&GocdrRoutinesLock)

	if routine, ok := GocdrRoutines[uint64(id)]; ok {
		unlock(&GocdrRoutinesLock)
		res := ""

		for i, elem := range routine.Trace {
			if i != 0 {
				res += "\n"
			}
			res += elem.toString()

			if i%1000 == 0 {
				c <- res
				res = ""
			}
		}
		c <- res
	} else {
		unlock(&GocdrRoutinesLock)
	}
}

// Return a string representation of all routine
// Return:
//   - string representation of the trace of all routines
func AllTracesToString() string {
	// write warning if projectPath is empty
	res := ""
	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)

	for i := 1; i <= len(GocdrRoutines); i++ {
		res += ""
		routine := GocdrRoutines[uint64(i)]
		if routine == nil {
			panic("Trace is nil")
		}
		res += traceToString(&routine.Trace) + "\n"

	}
	return res
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

// PrintAllTraces prints all traces
func PrintAllTraces() {
	print(AllTracesToString())
}

// GetNumberOfRoutines returns the number of routines in the trace
//
// Returns:
//   - number of routines in the trace
func GetNumberOfRoutines() int {
	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)
	return len(GocdrRoutines)
}

// DeleteTrace removes all trace elements from the trace
// It does not remove the routine objects them self
// Make sure to call BlockTrace(), before calling this function
func DeleteTrace() {
	lock(&GocdrRoutinesLock)
	defer unlock(&GocdrRoutinesLock)
	for i := range GocdrRoutines {
		GocdrRoutines[i].Trace = GocdrRoutines[i].Trace[:0]
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
func GocdrIgnore(file string) bool {
	return (containsStr(file, "goPatch/src/") || containsStr(file, "go/pkg/mod")) &&
		!containsStr(file, "goPatch/src/time/tick.go") &&
		!containsStr(file, "goPatch/src/context/context.go")
}

// GOCDR-FILE-END
