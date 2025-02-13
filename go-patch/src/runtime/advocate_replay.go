// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Functions for the replay
//
// Author: Erik Kassubek
// Created: 2023-10-24
//
// License: BSD-3-Clause

package runtime

const (
	ExitCodeDefault          = 0
	ExitCodePanic            = 3
	ExitCodeTimeout          = 10
	ExitCodeLeakUnbuf        = 20
	ExitCodeLeakBuf          = 21
	ExitCodeLeakMutex        = 22
	ExitCodeLeakCond         = 23
	ExitCodeLeakWG           = 24
	ExitCodeSendClose        = 30
	ExitCodeRecvClose        = 31
	ExitCodeNegativeWG       = 32
	ExitCodeUnlockBeforeLock = 33
	ExitCodeCyclic           = 41
)

var ExitCodeNames = map[int]string{
	0:  "The replay terminated without confirming the predicted bug",
	3:  "The program panicked unexpectedly",
	10: "Timeout",
	20: "Leak: Leaking unbuffered channel or select was unstuck",
	21: "Leak: Leaking buffered channel was unstuck",
	22: "Leak: Leaking Mutex was unstuck",
	23: "Leak: Leaking Cond was unstuck",
	24: "Leak: Leaking WaitGroup was unstuck",
	30: "Send on close",
	31: "Receive on close",
	32: "Negative WaitGroup counter",
	33: "Unlock of unlocked mutex",
	41: "Cyclic deadlock",
}

var hasReturnedExitCode = false
var ignoreAtomicsReplay = true

var printDebug = false

func SetReplayAtomic(repl bool) {
	ignoreAtomicsReplay = !repl
}

func GetReplayAtomic() bool {
	return !ignoreAtomicsReplay
}

/*
 * String representation of the replay operation.
 * Return:
 * 	string: string representation of the replay operation
 */
func (ro Operation) ToString() string {
	switch ro {
	case OperationNone:
		return "OperationNone"
	case OperationSpawn:
		return "OperationSpawn"
	case OperationSpawned:
		return "OperationSpawned"
	case OperationChannelSend:
		return "OperationChannelSend"
	case OperationChannelRecv:
		return "OperationChannelRecv"
	case OperationChannelClose:
		return "OperationChannelClose"
	case OperationMutexLock:
		return "OperationMutexLock"
	case OperationMutexUnlock:
		return "OperationMutexUnlock"
	case OperationMutexTryLock:
		return "OperationMutexTryLock"
	case OperationRWMutexLock:
		return "OperationRWMutexLock"
	case OperationRWMutexUnlock:
		return "OperationRWMutexUnlock"
	case OperationRWMutexTryLock:
		return "OperationRWMutexTryLock"
	case OperationRWMutexRLock:
		return "OperationRWMutexRLock"
	case OperationRWMutexRUnlock:
		return "OperationRWMutexRUnlock"
	case OperationRWMutexTryRLock:
		return "OperationRWMutexTryRLock"
	case OperationOnce:
		return "OperationOnce"
	case OperationWaitgroupAddDone:
		return "OperationWaitgroupAddDone"
	case OperationWaitgroupWait:
		return "OperationWaitgroupWait"
	case OperationSelect:
		return "OperationSelect"
	case OperationSelectCase:
		return "OperationSelectCase"
	case OperationSelectDefault:
		return "OperationSelectDefault"
	case OperationCondSignal:
		return "OperationCondSignal"
	case OperationCondBroadcast:
		return "OperationCondBroadcast"
	case OperationCondWait:
		return "OperationCondWait"
	case OperationReplayEnd:
		return "OperationReplayEnd"
	default:
		return "Unknown"
	}
}

/*
 * The replay data structure.
 * The replay data structure is used to store the trace of the program.
 * op: identifier of the operation
 * time: time of the operation
 * timePre: pre time
 * file: file in which the operation is executed
 * line: line number of the operation
 * blocked: true if the operation is blocked (never finised, tpost=0), false otherwise
 * suc: success of the opeartion
 *     - for mutexes: trylock operations true if the lock was acquired, false otherwise
 * 			for other operations always true
 *     - for once: true if the once was chosen (was the first), false otherwise
 *     - for others: always true
 * PFile: file of the partner (mainly for channel/select)
 * PLine: line of the partner (mainly for channel/select)
 * SelIndex: index of the select case (only for select, otherwise)
 */
type ReplayElement struct {
	Routine  int
	Op       Operation
	Time     int
	TimePre  int
	File     string
	Line     int
	Blocked  bool
	Suc      bool
	PFile    string
	PLine    int
	SelIndex int
}

type AdvocateReplayTrace []ReplayElement
type AdvocateReplayTraces map[uint64]AdvocateReplayTrace // routine -> trace

var replayEnabled bool // replay is on
var replayLock mutex
var replayDone int
var replayDoneLock mutex

// read trace
var replayData = make(AdvocateReplayTraces, 0)
var numberElementsInTrace int
var traceElementPositions = make(map[string][]int) // file -> []line

// exit code
var replayExitCode bool
var expectedExitCode int

// for leak, TimePre of stuck elem
var lastTPreReplay int
var stuckReplayExecutedSuc = false

var timeoutHappened = false

/*
 * Add a replay trace to the replay data.
 * Arguments:
 * 	routine: the routine id
 * 	trace: the replay trace
 */
func AddReplayTrace(routine uint64, trace AdvocateReplayTrace) {
	if _, ok := replayData[routine]; ok {
		panic("Routine already exists")
	}
	replayData[routine] = trace

	numberElementsInTrace += len(trace)

	for _, e := range trace {
		if _, ok := traceElementPositions[e.File]; !ok {
			traceElementPositions[e.File] = make([]int, 0)
		}
		if !containsInt(traceElementPositions[e.File], e.Line) {
			traceElementPositions[e.File] = append(traceElementPositions[e.File], e.Line)
		}
	}
}

/*
 * Print the replay data.
 */
func (t AdvocateReplayTraces) Print() {
	for id, trace := range t {
		println("\nRoutine: ", id)
		trace.Print()
	}
}

/*
 * Print the replay trace for one routine.
 */
func (t AdvocateReplayTrace) Print() {
	for _, e := range t {
		println(e.Op.ToString(), e.Time, e.File, e.Line, e.Blocked, e.Suc)
	}
}

/*
 * Enable the replay.
 */
func EnableReplay() {
	go ReleaseWaits()

	replayEnabled = true
}

/*
 * Disable the replay. This is called when a stop character in the trace is
 * encountered.
 */
func DisableReplay() {
	lock(&replayLock)
	defer unlock(&replayLock)

	if !replayEnabled {
		return
	}

	replayEnabled = false

	lock(&waitingOpsMutex)
	for _, replCh := range waitingOps {
		replCh.chWait <- ReplayElement{Blocked: false}
	}
	unlock(&waitingOpsMutex)
}

/*
 * Wait until all operations in the trace are executed.
 * This function should be called after the main routine is finished, to prevent
 * the program to terminate before the trace is finished.
 */
func WaitForReplayFinish(exit bool) {
	if !IsReplayEnabled() {
		for {
			lock(&replayDoneLock)
			if replayDone >= numberElementsInTrace {
				unlock(&replayDoneLock)
				break
			}
			unlock(&replayDoneLock)

			if !replayEnabled {
				break
			}

			sleep(0.001)
		}

		DisableReplay()

		// wait long enough, that all operations that have been released in the displayReplay
		// can record the pre
		sleep(1)
	}

	if stuckReplayExecutedSuc {
		ExitReplayWithCode(expectedExitCode)
	}
}

func IsReplayEnabled() bool {
	return replayEnabled
}

/*
 * Function to run in the background and to release the waiting operations
 */
func ReleaseWaits() {
	lastKey := ""
	lastTime := currentTime()
	lastTimeWithoutOldest := currentTime()

	const releaseOldestWaitLastMax int64 = 6
	const releaseWaitMaxWait = 20
	const releaseWaitMaxNoWait = 10
	const acknowledgementMaxWaitSec = 5
	releaseOldestWait := releaseOldestWaitLastMax

	for {
		counter++
		routine, replayElem := getNextReplayElement()

		if routine == -1 {
			continue
		}

		if replayElem.Op == OperationReplayEnd {
			if isExitCodeConfOnEndElem(replayElem.Line) {
				ExitReplayWithCode(replayElem.Line)
			}

			// wait long enough, that all operations that have been released but not
			// finished executing can execute
			sleep(0.5)

			DisableReplay()
			// foundReplayElement(routine)
			return
		}

		// key := intToString(routine) + ":" + replayElem.File + ":" + intToString(replayElem.Line)
		key := replayElem.File + ":" + intToString(replayElem.Line)
		if key == lastKey {
			if hasTimePast(lastTime, releaseOldestWait) {
				timeoutHappened = true
				var oldest = replayChan{nil, nil, -1, false}
				oldestKey := ""
				lock(&waitingOpsMutex)
				for key, ch := range waitingOps {
					if oldest.counter == -1 || ch.counter < oldest.counter {
						oldest = ch
						oldestKey = key
					}
				}
				unlock(&waitingOpsMutex)
				if oldestKey != "" {
					oldest.chWait <- replayElem

					if printDebug {
						println("RelO: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
					}

					// wait for the acknowledgement
					if oldest.wait {
						select {
						case <-oldest.chAck:
							if printDebug {
								println("AckO: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
							}
						case <-after(sToNs(acknowledgementMaxWaitSec)):
							if printDebug {
								println("TOO: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
							}
							timeoutHappened = true
						}
					} else {
						if printDebug {
							println("AckO: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
						}
					}

					lastTime = currentTime()
					if releaseOldestWait > 1 {
						releaseOldestWait--
					}

					foundReplayElement(replayElem.Routine)

					lock(&replayDoneLock)
					replayDone++
					unlock(&replayDoneLock)

					lock(&waitingOpsMutex)
					if printDebug {
						println("Deli: ", oldestKey)
					}
					delete(waitingOps, oldestKey)
					unlock(&waitingOpsMutex)
				}
			}
			if (hasTimePast(lastTimeWithoutOldest, releaseWaitMaxNoWait) && len(waitingOps) == 0) || hasTimePast(lastTimeWithoutOldest, releaseWaitMaxWait) {
				DisableReplay()
			}
		}

		if AdvocateIgnoreReplay(replayElem.Op, replayElem.File) {
			foundReplayElement(routine)

			lock(&replayDoneLock)
			replayDone++
			unlock(&replayDoneLock)
			continue
		}

		if key != lastKey {
			lastKey = key
			if printDebug {
				println("\n\n===================\nNext: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
				println("Currently Waiting: ", len(waitingOps))
				for key := range waitingOps {
					println(key)
				}
				println("===================\n\n")
			}
		}

		lock(&waitingOpsMutex)
		if waitCh, ok := waitingOps[key]; ok {
			unlock(&waitingOpsMutex)
			waitCh.chWait <- replayElem

			if printDebug {
				println("RelR: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
			}

			// wait for the acknowledgement
			if waitCh.wait {
				select {
				case <-waitCh.chAck:
					if printDebug {
						println("AckR: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
					}
				case <-after(sToNs(acknowledgementMaxWaitSec)):
					if printDebug {
						println("TOR: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
					}
				}
			} else {
				if printDebug {
					println("AckO: ", replayElem.Op.ToString(), replayElem.File, replayElem.Line)
				}
			}

			lastTime = currentTime()
			lastTimeWithoutOldest = currentTime()
			releaseOldestWait = releaseOldestWaitLastMax

			foundReplayElement(routine)

			lock(&replayDoneLock)
			replayDone++
			unlock(&replayDoneLock)

			lock(&waitingOpsMutex)
			if printDebug {
				println("Deli: ", key)
			}
			delete(waitingOps, key)
		}
		unlock(&waitingOpsMutex)

		if !replayEnabled {
			return
		}
	}
}

type replayChan struct {
	chWait  chan ReplayElement
	chAck   chan struct{}
	counter int
	wait    bool
}

// Map of all currently waiting operations
var waitingOps = make(map[string]replayChan)
var waitingOpsMutex mutex
var counter = 0

/*
 * Wait until the correct operation is about to be executed.
 * Arguments:
 * 	op: the operation type that is about to be executed
 * 	skip: number of stack frames to skip
 * 	waitForResponse bool: whether the wait should wait for a response after finished
 * Return:
 * 	bool: true if the operation should wait, false otherwise
 * 	chan ReplayElement: channel to wait on
 * 	chan struct{}: chan to report back on when finished
 */
func WaitForReplay(op Operation, skip int, waitForResponse bool) (bool, chan ReplayElement, chan struct{}) {
	_, file, line, _ := Caller(skip)

	return WaitForReplayPath(op, file, line, waitForResponse)
}

/*
 * Wait until the correct operation is about to be executed.
 * Arguments:
 * 	op: the operation type that is about to be executed
 * 	file: file in which the operation is executed
 * 	line: line number of the operation
 * 	waitForResponse bool: whether the wait should wait for a response after finished
 * Return:
 * 	bool: true if the operation should wait, false otherwise
 * 	chan ReplayElement: channel to wait on
 */
func WaitForReplayPath(op Operation, file string, line int, waitForResponse bool) (bool, chan ReplayElement, chan struct{}) {
	if !replayEnabled {
		return false, nil, nil
	}

	if AdvocateIgnoreReplay(op, file) {
		return false, nil, nil
	}

	// routine := GetRoutineID()
	// key := uint64ToString(routine) + ":" + file + ":" + intToString(line)
	key := file + ":" + intToString(line)

	if printDebug {
		println("Wait: ", op.ToString(), file, line)
	}

	chWait := make(chan ReplayElement, 0)
	chResp := make(chan struct{}, 1)
	lock(&waitingOpsMutex)
	if _, ok := waitingOps[key]; ok {
		println("Override key: ", key)
	}
	waitingOps[key] = replayChan{chWait, chResp, counter, waitForResponse}
	unlock(&waitingOpsMutex)

	return true, chWait, chResp
}

/*
 * Check if the position is in the trace.
 * Args:
 * 	file: file in which the operation is executed
 * 	line: line number of the operation
 * Return:
 * 	bool: true if the position is in the trace, false otherwise
 */
func isPositionInTrace(file string, line int) bool {
	if _, ok := traceElementPositions[file]; !ok {
		return false
	}

	if !containsInt(traceElementPositions[file], line) {
		return false
	}

	return true
}

func correctSelect(next Operation, op Operation) bool {
	if op != OperationSelect {
		return false
	}

	if next != OperationSelectCase && next != OperationSelectDefault {
		return false
	}

	return true
}

func BlockForever() {
	gopark(nil, nil, waitReasonZero, traceBlockForever, 1)
}

/*
 * Get the next replay element.
 * Return:
 * 	uint64: the routine of the next replay element or -1 if the trace is empty
 * 	ReplayElement: the next replay element
 */
func getNextReplayElement() (int, ReplayElement) {
	lock(&replayLock)
	defer unlock(&replayLock)

	routine := -1
	// set mintTime to max int
	var minTime = -1

	for id, trace := range replayData {
		if len(trace) == 0 {
			continue
		}
		elem := trace[0]
		if minTime == -1 || elem.Time < minTime {
			minTime = elem.Time
			routine = int(id)
		}
	}

	if routine == -1 {
		return -1, ReplayElement{}
	}

	return routine, replayData[uint64(routine)][0]
}

func foundReplayElement(routine int) {
	lock(&replayLock)
	defer unlock(&replayLock)

	// remove the first element from the trace for the routine
	replayData[uint64(routine)] = replayData[uint64(routine)][1:]
}

/*
 * Set the replay code
 * Args:
 * 	code: the replay code
 */
func SetExitCode(code bool) {
	replayExitCode = code
}

/*
 * Set the expected exit code
 * Args:
 * 	code: the expected exit code
 */
func SetExpectedExitCode(code int) {
	expectedExitCode = code
}

func SetLastTPre(tPre int) {
	lastTPreReplay = tPre
}

/*
  - Set the time of the

/*
  - Exit the program with the given code.
  - Args:
  - code: the exit code
*/
func ExitReplayWithCode(code int) {
	if !hasReturnedExitCode {
		if isExitCodeConfOnEndElem(code) && !stuckReplayExecutedSuc {
			return
		}
		println("Exit Replay with code ", code, ExitCodeNames[code])
		hasReturnedExitCode = true
	}
	if replayExitCode && ExitCodeNames[code] != "" {
		if !advocateTracingDisabled { // do not exit if recording is enabled
			return
		}
		exit(int32(code))
	}
}

/*
 * For some exit codes, the replay is seen as confirmed, if the replay end
 * element is reached. This function returns wether the exit code is
 * such a code
 * The codes are
 *    20 - 29: Leak
 *    40 - 49: Deadlocks
 *
 */
func isExitCodeConfOnEndElem(code int) bool {
	return (code >= 20 && code < 30) || (code >= 40 && code <= 49)
}

/*
 * Exit the program with the given code if the program panics.
 * Args:
 * 	msg: the panic message
 */
func ExitReplayPanic(msg any) {
	if !IsReplayEnabled() {
		return
	}

	println("Exit with panic")
	switch m := msg.(type) {
	case plainError:
		if expectedExitCode == ExitCodeSendClose && m.Error() == "send on closed channel" {
			ExitReplayWithCode(ExitCodeSendClose)
		}
	case string:
		if expectedExitCode == ExitCodeNegativeWG && m == "sync: negative WaitGroup counter" {
			ExitReplayWithCode(ExitCodeNegativeWG)
		} else if expectedExitCode == ExitCodeUnlockBeforeLock {
			if m == "sync: RUnlock of unlocked RWMutex" ||
				m == "sync: Unlock of unlocked RWMutex" ||
				m == "sync: unlock of unlocked mutex" {
				ExitReplayWithCode(ExitCodeUnlockBeforeLock)
			}
		} else if hasPrefix(m, "test timed out after") {
			ExitReplayWithCode(ExitCodeTimeout)
		}
	}

	ExitReplayWithCode(ExitCodePanic)
}

func AdvocateIgnoreReplay(operation Operation, file string) bool {
	if ignoreAtomicsReplay && getOperationObjectString(operation) == "Atomic" {
		return true
	}

	if contains(file, "go/pkg/mod/") {
		return true
	}

	return AdvocateIgnore(file)
}
