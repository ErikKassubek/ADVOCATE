// ADVOCATE-FILE-START

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

	OperationOnce

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

	OperationReplayEnd
)

type prePost int // enum for pre/post
const (
	pre prePost = iota
	post
	none
)

var advocateTracingDisabled = true
var advocatePanicWriteBlock chan struct{}
var advocatePanicDone chan struct{}

// var advocateTraceWritingDisabled = false

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
	case OperationOnce:
		return "Once"
	case OperationWaitgroupAddDone, OperationWaitgroupWait:
		return "Waitgroup"
	case OperationSelect, OperationSelectCase, OperationSelectDefault:
		return "Select"
	case OperationCondSignal, OperationCondBroadcast, OperationCondWait:
		return "Cond"
	case OperationAtomicLoad, OperationAtomicStore, OperationAtomicAdd, OperationAtomicSwap, OperationAtomicCompareAndSwap:
		return "Atomic"
	case OperationReplayEnd:
		return "Replay"
	}
	return "Unknown"
}

/*
 * Get the channels used to write the trace on certain panics
 * Args:
 *    apwb (chan struct{}): advocatePanicWriteBlock
 *    apd (chan struct{}): advocatePanicDone
 */
func GetAdvocatePanicChannels(apwb, apd chan struct{}) {
	advocatePanicWriteBlock = apwb
	advocatePanicDone = apd
}

/*
 * Return a string representation of the trace
 * Return:
 * 	string representation of the trace
 */
func CurrentTraceToString() string {
	res := ""
	for i, elem := range currentGoRoutine().Trace {
		if i != 0 {
			res += "\n"
		}
		res += elem
	}

	return res
}

/*
 * Return a string representation of the trace
 * Args:
 * 	trace: trace to convert to string
 * Return:
 * 	string representation of the trace
 */
func traceToString(trace *[]string, atomics *[]string) string {
	res := ""

	println("TraceToString", len(*trace), len(*atomics), len(*trace)+len(*atomics))

	// if atomic recording is disabled
	for i, elem := range *trace {
		if i != 0 {
			res += "\n"
		}
		res += elem
	}
	return res
}

func getTpre(elem string) int {
	split := splitStringAtCommas(elem, []int{1, 2})
	return stringToInt(split[1])
}

/*
 * Add an operation to the trace
 * Args:
 *  elem: element to add to the trace
 * Return:
 * 	index of the element in the trace
 */
func insertIntoTrace(elem string) int {
	return currentGoRoutine().addToTrace(elem)
}

/*
 * Print the trace of the current routines
 */
func PrintTrace() {
	routineID := GetRoutineID()
	println("Routine", routineID, ":", CurrentTraceToString())
}

/*
 * Return the trace of the routine with id 'id'
 * Args:
 * 	id: id of the routine
 * Return:
 * 	string representation of the trace of the routine
 * 	bool: true if the routine exists, false otherwise
 */
func TraceToStringByID(id uint64) (string, bool) {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	if routine, ok := AdvocateRoutines[id]; ok {
		return traceToString(&routine.Trace, &routine.Atomics), true
	}
	return "", false
}

/*
 * Return whether the trace of a routine' is empty
 * Args:
 * 	routine: id of the routine
 * Return:
 * 	true if the trace is empty, false otherwise
 */
func TraceIsEmptyByRoutine(routine int) bool {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	if routine, ok := AdvocateRoutines[uint64(routine)]; ok {
		return len(routine.Trace) == 0
	}
	return true
}

/*
 * Get the trace of the routine with id 'id'.
 * To minimized the needed ram the trace is sent to the channel 'c' in chunks
 * of 1000 elements.
 * Args:
 * 	id: id of the routine
 * 	c: channel to send the trace to
 *  atomic: it true, the atomic trace is returned
 */
func TraceToStringByIDChannel(id int, c chan<- string) {
	lock(&AdvocateRoutinesLock)

	if routine, ok := AdvocateRoutines[uint64(id)]; ok {
		unlock(&AdvocateRoutinesLock)
		res := ""

		// if atomic recording is disabled
		for i, elem := range routine.Trace {
			if i != 0 {
				res += "\n"
			}
			res += elem

			if i%1000 == 0 {
				c <- res
				res = ""
			}
		}
		c <- res
	} else {
		unlock(&AdvocateRoutinesLock)
	}
}

/*
 * Return the trace of all traces
 * Return:
 * 	string representation of the trace of all routines
 */
func AllTracesToString() string {
	// write warning if projectPath is empty
	res := ""
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)

	for i := 1; i <= len(AdvocateRoutines); i++ {
		res += ""
		routine := AdvocateRoutines[uint64(i)]
		if routine == nil {
			panic("Trace is nil")
		}
		res += traceToString(&routine.Trace, &routine.Atomics) + "\n"

	}
	return res
}

/*
* PrintAllTraces prints the trace of all routines
 */
func PrintAllTraces() {
	print(AllTracesToString())
}

/*
 * GetNumberOfRoutines returns the number of routines in the trace
 * Return:
 *	number of routines in the trace
 */
func GetNumberOfRoutines() int {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	return len(AdvocateRoutines)
}

/*
 * InitAdvocate enables the collection of the trace
 * Args:
 * 	size: size of the channel used to link the atomic recording to the main
 *    recording.
 */
func InitAdvocate() {
	advocateTracingDisabled = false
}

/*
 * DisableTrace disables the collection of the trace
 */
func DisableTrace() {
	advocateTracingDisabled = true
}

/*
 * GetAdvocateDisabled returns if the trace collection is disabled
 * Return:
 * 	true if the trace collection is disabled, false otherwise
 */
func GetAdvocateDisabled() bool {
	return advocateTracingDisabled
}

// /*
//  * BockTrace blocks the trace collection
//  * Resume using UnblockTrace
//  */
// func BlockTrace() {
// 	advocateTraceWritingDisabled = true
// }

// /*
//  * UnblockTrace resumes the trace collection
//  * Block using BlockTrace
//  */
// func UnblockTrace() {
// 	advocateTraceWritingDisabled = false
// }

/*
 * DeleteTrace removes all trace elements from the trace
 * Do not remove the routine objects them self
 * Make sure to call BlockTrace(), before calling this function
 */
func DeleteTrace() {
	lock(&AdvocateRoutinesLock)
	defer unlock(&AdvocateRoutinesLock)
	for i := range AdvocateRoutines {
		AdvocateRoutines[i].Trace = AdvocateRoutines[i].Trace[:0]
	}
}

// ====================== Ignore =========================

/*
 * Some operations, like garbage collection and internal operations, can
 * cause the replay to get stuck or are not needed.
 * For this reason, we ignore them.
 * Arguments:
 * 	operation: operation that is about to be executed
 * 	file: file in which the operation is executed
 * 	line: line number of the operation
 * Return:
 * 	bool: true if the operation should be ignored, false otherwise
 */
func AdvocateIgnore(operation Operation, file string, line int) bool {
	if hasSuffix(file, "advocate/advocate.go") ||
		hasSuffix(file, "runtime/advocate_replay.go") ||
		hasSuffix(file, "runtime/advocate_routine.go") ||
		hasSuffix(file, "runtime/advocate_trace.go") ||
		hasSuffix(file, "runtime/advocate_utile.go") ||
		hasSuffix(file, "runtime/advocate_atomic.go") { // internal
		return true
	} else if hasSuffix(file, "syscall/env_unix.go") {
		return true
	} else if hasSuffix(file, "runtime/signal_unix.go") {
		return true
	} else if hasSuffix(file, "runtime/mgc.go") { // garbage collector
		return true
	} else if hasSuffix(file, "runtime/panic.go") {
		return true
	}

	switch operation {
	case OperationMutexLock, OperationMutexUnlock:
		// mutex operations in the once can cause the replay to get stuck,
		// if the once was called by the poll/fd_poll_runtime.go init.
		if hasSuffix(file, "sync/once.go") && (line == 113 || line == 114 ||
			line == 119 || line == 123) {
			return true
		}
		// pools
		if hasSuffix(file, "sync/pool.go") && (line == 226 || line == 243) {
			return true
		}
		// mutex in rwmutex
		// if hasSuffix(file, "sync/rwmutex.go") && (line == 270 || line == 396) {
		// 	return true
		// }
	case OperationOnce:
		// once operations in the poll/fd_poll_runtime.go init can cause the replay to get stuck.
		if hasSuffix(file, "internal/poll/fd_poll_runtime.go") && line == 40 {
			return true
		}
	}
	return false
}

// ADVOCATE-FILE-END
