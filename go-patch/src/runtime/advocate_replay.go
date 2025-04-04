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

const (
	releaseOldestWaitLastMax  int64 = 6
	releaseWaitMaxWait              = 20
	releaseWaitMaxNoWait            = 10
	acknowledgementMaxWaitSec       = 5
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

var (
	hasReturnedExitCode = false
	ignoreAtomicsReplay = true
	printDebug          = false

	tPostWhenOldestFirstReleased = 0
	tPostWhenReplayDisabled      = 0
	tPostWhenAckFirstTimeout     = 0
)

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
	case OperationOnceDo:
		return "OperationOnceDo"
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

func (elem *ReplayElement) key() string {
	return buildReplayKey(elem.Routine, elem.File, elem.Line)
}

func buildReplayKey(routine int, file string, line int) string {
	return intToString(routine) + ":" + file + ":" + intToString(line)
}

type AdvocateReplayTrace []ReplayElement
type AdvocateReplayTraces map[uint64]AdvocateReplayTrace // routine -> trace

var (
	replayEnabled  bool // replay is on
	replayLock     mutex
	replayDone     int
	replayDoneLock mutex

	// read trace
	replayData            = make(AdvocateReplayTraces, 0)
	numberElementsInTrace int
	traceElementPositions = make(map[string][]int) // file -> []line

	replayIndex = make(map[uint64]int)

	// exit code
	replayExitCode   bool
	expectedExitCode int

	// for leak, TimePre of stuck elem
	stuckReplayExecutedSuc = false

	// for replay timeout
	lastKey               string
	lastTime              int64
	lastTimeWithoutOldest int64
	releaseOldestWait     = releaseOldestWaitLastMax
)

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
	replayIndex[routine] = 0

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
		sleep(0.5)
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
	lastTime = currentTime()
	lastTimeWithoutOldest = currentTime()

	for {
		// wait for acknowledgement of element that was directly
		// released when it was called, because it was the next element
		if waitForAck.waitForAck {
			releaseElement(waitForAck.replayC, waitForAck.replayE, false, true)
			waitForAck.waitForAck = false
		}

		counter++
		routine, replayElem := getNextReplayElement()

		if routine == -1 {
			break
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

		key := replayElem.key()
		if key == lastKey {
			if hasTimePast(lastTime, releaseOldestWait) {
				var oldest = replayChan{nil, nil, -1, false, 0, "", 0}
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
					if tPostWhenOldestFirstReleased == 0 {
						tPostWhenAckFirstTimeout = replayElem.Time
					}

					releaseElement(oldest, replayElemFromKey(oldestKey), true, false)

					if releaseOldestWait > 1 {
						releaseOldestWait--
					}

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
			if (len(waitingOps) == 0 && hasTimePast(lastTimeWithoutOldest, releaseWaitMaxNoWait)) || hasTimePast(lastTimeWithoutOldest, releaseWaitMaxWait) {
				tPostWhenReplayDisabled = replayElem.Time
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
		if waitOp, ok := waitingOps[key]; ok {
			unlock(&waitingOpsMutex)

			releaseElement(waitOp, replayElem, true, true)

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

func replayElemFromKey(key string) ReplayElement {
	keySplit := split(key, ':')
	return ReplayElement{
		Routine: stringToInt(keySplit[0]),
		File:    keySplit[1],
		Line:    stringToInt(keySplit[2]),
		Suc:     true,
		Blocked: false,
	}
}

type replayChan struct {
	chWait  chan ReplayElement
	chAck   chan struct{}
	counter int
	waitAck bool
	routine int
	file    string
	line    int
}

type released struct {
	replayC    replayChan
	replayE    ReplayElement
	waitForAck bool
}

// Map of all currently waiting operations
var waitingOps = make(map[string]replayChan)
var waitingOpsMutex mutex
var counter = 0

// element
var waitForAck released

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
	// without this the runtime build runs into an deadlock message, no idea why
	if !replayEnabled {
		return false, nil, nil
	}

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

	routine := getg().advocateRoutineInfo.replayRoutine

	// routine := GetRoutineID()
	key := buildReplayKey(routine, file, line)

	if printDebug {
		println("Wait: ", op.ToString(), file, line)
	}

	chWait := make(chan ReplayElement, 1)
	chAck := make(chan struct{}, 1)

	replayElem := replayChan{chWait, chAck, counter, waitForResponse, routine, file, line}

	_, nextElem := getNextReplayElement()
	if key == nextElem.key() {
		// if it is the next element, release directly and add elems to waitForAck
		replayElem.chWait <- nextElem
		if replayElem.waitAck {
			waitForAck = released{
				replayC:    replayElem,
				replayE:    nextElem,
				waitForAck: true,
			}
		}
	} else {
		// add to waiting list
		lock(&waitingOpsMutex)
		if _, ok := waitingOps[key]; ok {
			println("Override key: ", key)
		}
		waitingOps[key] = replayElem
		unlock(&waitingOpsMutex)
	}

	return true, chWait, chAck
}

/*
 * Release a waiting operation and, if required, wait for the acknowledgement
 * Args:
 *  elem (replayChan): the element to be released
 *  elemReplay (ReplayElement): the corresponding replay element
 *  rel (bool): the waiting channel should be released
 *  next (bool): true if the next element in the trace was released, false if the oldest has been released
 */
func releaseElement(elem replayChan, elemReplay ReplayElement, rel, next bool) {
	if rel {
		elem.chWait <- elemReplay
	}

	if elem.waitAck {
		select {
		case <-elem.chAck:
		case <-after(sToNs(acknowledgementMaxWaitSec)):
			if tPostWhenAckFirstTimeout == 0 {
				tPostWhenAckFirstTimeout = elemReplay.Time
			}
		}
	}

	lastTime = currentTime()
	if next {
		lastTimeWithoutOldest = currentTime()
		releaseOldestWait = releaseOldestWaitLastMax
		foundReplayElement(elemReplay.Routine)
	} else {
		releasedElementOldest(elemReplay.key())
	}
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

/*
 * When called, the calling routine is blocked and cannot be woken up again
 */
func BlockForever() {
	gopark(nil, nil, waitReasonZero, traceBlockForever, 1)
}

var alreadyExecutedAsOldest = make(map[string]int)

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
		if replayIndex[id] >= len(trace) {
			continue
		}
		elem := trace[replayIndex[id]]
		if minTime == -1 || elem.Time < minTime {
			minTime = elem.Time
			routine = int(id)
		}
	}

	if routine == -1 {
		return -1, ReplayElement{}
	}

	elem := replayData[uint64(routine)][replayIndex[uint64(routine)]]

	// if the elem was already executed as an oldest before, do not get again
	elemKey := elem.key()
	if val, ok := alreadyExecutedAsOldest[elemKey]; ok && val > 0 {
		foundReplayElement(elem.Routine)
		alreadyExecutedAsOldest[elemKey]--
		return getNextReplayElement()
	}

	return routine, replayData[uint64(routine)][0]
}

func releasedElementOldest(key string) {
	alreadyExecutedAsOldest[key]++
}

func foundReplayElement(routine int) {
	lock(&replayLock)
	defer unlock(&replayLock)
	replayIndex[uint64(routine)]++
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

/*
- Exit the program with the given code.
- Args:
- code: the exit code
*/
func ExitReplayWithCode(code int) {
	if !hasReturnedExitCode {
		// TODO: is this correct?
		if isExitCodeConfOnEndElem(code) && !stuckReplayExecutedSuc {
			return
		}
		println("\nExit Replay with code ", code, ExitCodeNames[code])
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
		} else if hasPrefix(m, "test timed out after") || m == "Timeout" {
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

/*
 * Return the replay status
 * Return:
 * 	(int): what was the next tPost, when the manager released an oldest for the first time, if never return 0
 * 	(int): what was the next tPost, when the manager disables the replay because it was stuck, if not return 0
 * 	(int): what was the next tPost, when an expected Acknowledgement timed out, if never return 0
 */
func GetReplayStatus() (int, int, int) {
	return tPostWhenOldestFirstReleased, tPostWhenReplayDisabled, tPostWhenAckFirstTimeout
}
