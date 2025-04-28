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
	ExitCodeCloseClose       = 32
	ExitCodeCloseNil         = 33
	ExitCodeNegativeWG       = 34
	ExitCodeUnlockBeforeLock = 35
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
	32: "Close on close",
	33: "Close on nil",
	34: "Negative WaitGroup counter",
	35: "Unlock of unlocked mutex",
	41: "Cyclic deadlock",
}

var (
	hasReturnedExitCode = false
	ignoreAtomicsReplay = true
	printDebug          = false

	tPostWhenFirstTimeout    = 0
	tPostWhenReplayDisabled  = 0
	tPostWhenAckFirstTimeout = 0

	releasedFork = 0
)

func SetReplayAtomic(repl bool) {
	ignoreAtomicsReplay = !repl
}

func GetReplayAtomic() bool {
	return !ignoreAtomicsReplay
}

// String representation of the replay operation.
//
// Returns:
//   - string: string representation of the replay operation
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

// The replay data structure.
// The replay data structure is used to store the routine local trace of the replay.
//
// Fields:
// - Routine int: id of the represented routine
//   - op: identifier of the operation
//   - time: time of the operation
//   - timePre: pre time
//   - file: file in which the operation is executed
//   - line: line number of the operation
//   - blocked: true if the operation is blocked (never finised, tpost=0), false otherwise
//   - suc: success of the opeartion
//     for mutexes: trylock operations true if the lock was acquired, false otherwise
//     for other operations always true
//     for once: true if the once was chosen (was the first), false otherwise
//     for others: always true
//   - Index: Index of the select case (only for select) or index of the new routine (only for spawn), otherwise 0
type ReplayElement struct {
	Routine int
	Op      Operation
	Time    int
	TimePre int
	File    string
	Line    int
	Blocked bool
	Suc     bool
	Index   int
}

// Get the key (id) of a replay element
//
// Returns:
//   - the key of elem
func (elem *ReplayElement) key() string {
	return BuildReplayKey(elem.Routine, elem.File, elem.Line)
}

// TODO (Erik): fix replay routine, if fixed also add routine in active key

// Build the key (id) of a replay element
//
// Parameters:
//   - routine int: replay routine of the element
//   - file string: code position file of the element
//   - line int: code position line of the element
func BuildReplayKey(routine int, file string, line int) string {
	// return intToString(routine) + ":" + file + ":" + intToString(line)
	return file + ":" + intToString(line)
}

type AdvocateReplayTrace []ReplayElement

var (
	replayEnabled  bool // replay is on
	replayLock     mutex

	// only run partial replay based on active
	partialReplay        bool
	partialReplayCounter = make(map[string]int)

	// detection of deadlock
	waitDeadlockDetect     bool
	waitDeadlockDetectLock mutex

	// replay info
	replayData            = make(AdvocateReplayTrace, 0)
	numberElementsInTrace int
	active                map[string][]int

	replayIndex = 0

	// exit code
	replayForceExit  bool
	expectedExitCode int

	// for leak, TimePre of stuck elem
	stuckReplayExecutedSuc = false

	// for replay timeout
	lastKey               string
	lastTime              int64
	lastTimeWithoutOldest int64
	releaseOldestWait     = releaseOldestWaitLastMax
)

// Add a routine local replay trace to the replay data.
//
// Parameters:
//   - trace trace: the replay trace
func GetReplayTrace() *AdvocateReplayTrace {
	return &replayData
}


// Print the replay trace for one routine.
func (t AdvocateReplayTrace) Print() {
	for _, e := range t {
		println(e.Op.ToString(), e.Time, e.File, e.Line, e.Blocked, e.Suc)
	}
}


// AddActiveTrace adds the set of active trace elements to the trace
// and sets the replay to be partial
//
// Parameter
//   - active map[string][int]: the map of active operations where the map
//     key is equal to the replay element key (buildReplayKey) and value is the
//     list of occurrences when the replay should be active for the element,
//     e.g. if the value is [3, 4], the operation in the key is scheduled by
//     the replay if it is executed the 3rd and 4th time, but not for the
//     1st and 2nd time.
func AddActiveTrace(activeMap map[string][]int) {
	active = activeMap
	partialReplay = true
}

// Enable the replay by starting the replay manager
func EnableReplay() {
	numberElementsInTrace = len(replayData)

	if printDebug {
		println("\nTrace\n")
		replayData.Print()
		println("\n\n")
	}


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
	if printDebug {
		println("Wait for replay finish")
	}
	if IsReplayEnabled() { // Is this correct?
		for {
			if replayIndex >= numberElementsInTrace {
				break
			}

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

	// Ensure that the deadlock detector is finished
	for {
		lock(&waitDeadlockDetectLock)
		if !waitDeadlockDetect {
			unlock(&waitDeadlockDetectLock)
			break
		}
		unlock(&waitDeadlockDetectLock)

		sleep(0.001)
	}


	if stuckReplayExecutedSuc {
		ExitReplayWithCode(expectedExitCode)
	}
}

func IsReplayEnabled() bool {
	return replayEnabled
}

// Function to run in the background and to release the waiting operations
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
			println("Found ReplayEnd Marker with exit code", replayElem.Line)
			// wait long enough, that all operations that have been released but not
			// finished executing can execute
			if replayElem.Line == ExitCodeCyclic {
				lock(&waitDeadlockDetectLock)
				waitDeadlockDetect = true
				unlock(&waitDeadlockDetectLock)
			}
			sleep(0.5)

			DisableReplay()
			// foundReplayElement()
			sleep(0.1)

			// Check if a deadlock has been reached
			if replayElem.Line == ExitCodeCyclic {
				stuckRoutines := checkForStuckRoutines(1.0, 100)

				stuckMutexCounter := 0
				for id, reason := range stuckRoutines {
					println("Routine", id, "is possibly stuck. Waiting with reason:", waitReasonStrings[reason])
					// TODO invert to everything that could NOT be a deadlock
					if reason == waitReasonSyncMutexLock || reason == waitReasonSyncRWMutexLock || reason == waitReasonSyncRWMutexRLock {
						stuckMutexCounter++
					}
				}

				println("Number of routines waiting on mutexes:", stuckMutexCounter)

				if stuckMutexCounter > 0 {
					SetForceExit(true)
					ExitReplayWithCode(replayElem.Line)
				}

				lock(&waitDeadlockDetectLock)
				waitDeadlockDetect = false
				unlock(&waitDeadlockDetectLock)
			} else if isExitCodeConfOnEndElem(replayElem.Line) {
				ExitReplayWithCode(replayElem.Line)
			}
			return
		}

		key := replayElem.key()

		if false && key == lastKey {
			if !waitForAck.waitForAck && hasTimePast(lastTime, releaseOldestWait) { // timeout
				if printDebug {
					println("TIMEOUT")
				}
				if tPostWhenFirstTimeout == 0 {
					tPostWhenAckFirstTimeout = replayElem.Time
				}

				// we either release the longest waiting operation
				// or skip the current next element in the trace
				// If no elements are waiting we always skip the current element in the trace
				// otherwise we choose either option with a prop of 0.5 (we use nanotime()%2 == 0 as a quasi random number generator from {0,1})
				if len(waitingOps) == 0 || nanotime()%2 == 0 {
					// skip the next element in the trace
					foundReplayElement()
				} else {
					// release the currently waiting element
					var oldest = replayChan{nil, nil, -1, false, 0, "", 0, false}
					oldestKey := ""
					lock(&waitingOpsMutex)
					for key, ch := range waitingOps {
						if oldest.counter == -1 || ch.counter < oldest.counter {
							oldest = ch
							oldestKey = key
						}
					}
					unlock(&waitingOpsMutex)

					suc := releaseElement(oldest, replayElemFromKey(oldestKey), true, false)

					if releaseOldestWait > 1 {
						releaseOldestWait--
					}

					lock(&waitingOpsMutex)
					if printDebug && suc {
						println("Release Oldes: ", oldestKey)
					}
					delete(waitingOps, oldestKey)
					unlock(&waitingOpsMutex)
				}
			}
		}

		if (len(waitingOps) == 0 && hasTimePast(lastTimeWithoutOldest, releaseWaitMaxNoWait)) || hasTimePast(lastTimeWithoutOldest, releaseWaitMaxWait) {
			tPostWhenReplayDisabled = replayElem.Time
			DisableReplay()
		}

		if AdvocateIgnoreReplay(replayElem.Op, replayElem.File) {
			foundReplayElement()
			continue
		}

		if key != lastKey {
			lastKey = key
			if printDebug {
				println("\n\n===================\nNext: ", key)
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

			lock(&waitingOpsMutex)
			delete(waitingOps, key)
		}
		unlock(&waitingOpsMutex)

		if !replayEnabled {
			return

		}
	}
}

// Given an replay element key, create the corresponding replay element
//
// Parameter:
//   - key (string): te replay key
//
// Returns:
//   - ReplayElement: a replay element that fits to the key
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

// Returns all routines for which the wait reason has not changed within checkStuckTime seconds.
// Used to detect if replay of deadlock was successful
//
// Parameters:
//   - checkStuckTime float64: find routines that have been waiting for at least this many seconds
//   - checkStuckIterations int: iterations to check
func checkForStuckRoutines(checkStuckTime float64, checkStuckIterations int) map[uint64]waitReason {
	stuckRoutines := make(map[uint64]waitReason)

	lock(&AdvocateRoutinesLock)
	for id, routine := range AdvocateRoutines {
		stuckRoutines[id] = routine.G.waitreason
	}
	unlock(&AdvocateRoutinesLock)

	// Repeatedly check if wait reason has changed
	for i := 0; i < checkStuckIterations; i++ {
		sleep(checkStuckTime / float64(checkStuckIterations))
		lock(&AdvocateRoutinesLock)
		for id, routine := range AdvocateRoutines {
			if _, ok := stuckRoutines[id]; ok && routine.G.waitreason != stuckRoutines[id] {
				delete(stuckRoutines, id)
			}
		}
		unlock(&AdvocateRoutinesLock)
	}
	return stuckRoutines
}

type replayChan struct {
	chWait   chan ReplayElement
	chAck    chan struct{}
	counter  int
	waitAck  bool
	routine  int
	file     string
	line     int
	released bool
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

// Called by operations during replay. If the operations is the next operation
// to be executed it is release, otherwise the operation is stored into the
// waiting operations. The function provides channels to the operation
// on which they wait to be released and send an acknowledgement when they have
// finished executing
//
// Parameter:
//   - op: the operation type that is about to be executed
//   - skip: number of stack frames to skip
//   - waitForResponse bool: whether the wait should wait for a response after finished
//
// Returns:
//   - bool: true if the operation should wait, otherwise it can immediately execute
//     and ignore all replay mechanisms, e.g. used if replay is enabled
//   - chan ReplayElement: channel to wait on
//   - chan struct{}: chan to report back on when finished
func WaitForReplay(op Operation, skip int, waitForResponse bool) (bool, chan ReplayElement, chan struct{}) {
	// without this the runtime build runs into an deadlock message, no idea why
	if !replayEnabled {
		return false, nil, nil
	}

	_, file, line, _ := Caller(skip)

	return WaitForReplayPath(op, file, line, waitForResponse)
}

// Called by operations during replay. If the operations is the next operation
// to be executed it is release, otherwise the operation is stored into the
// waiting operations. The function provides channels to the operation
// on which they wait to be released and send an acknowledgement when they have
// finished executing
//
// Parameter:
//   - op: the operation type that is about to be executed
//   - file: file in which the operation is executed
//   - line: line number of the operation
//   - waitForResponse bool: whether the wait should wait for a response after finished
//
// Returns:
//   - bool: true if the operation should wait, otherwise it can immediately execute
//     and ignore all replay mechanisms, e.g. used if replay is enabled
//   - chan ReplayElement: channel to wait on
//   - chan struct{}: chan to report back on when finished
func WaitForReplayPath(op Operation, file string, line int, waitForResponse bool) (bool, chan ReplayElement, chan struct{}) {
	if !replayEnabled {
		return false, nil, nil
	}

	if AdvocateIgnoreReplay(op, file) {
		return false, nil, nil
	}

	routine := GetRoutineID()

	// routine := GetRoutineID()
	key := BuildReplayKey(routine, file, line)

	// ignore not active operations if partial replay is active
	if partialReplay {
		// the operation is never active
		c, ok := active[key]
		if !ok {
			return false, nil, nil
		}


		partialReplayCounter[key] += 1
		currentCounter := partialReplayCounter[key]

		// the operation is sometimes active, but not this time
		if !isInSlice(c, currentCounter) {
			return false, nil, nil
		}
	}

	if printDebug {
		println("Wait: ", key)
	}

	chWait := make(chan ReplayElement, 1)
	chAck := make(chan struct{}, 1)

	replayElem := replayChan{chWait, chAck, counter, waitForResponse, routine, file, line, false}

	_, nextElem := getNextReplayElement()
	if key == nextElem.key() && !waitForAck.waitForAck {
		// if it is the next element, release directly and add elems to waitForAck
		replayElem.chWait <- nextElem
		if printDebug {
			println("ReleaseDir: ", key, waitForResponse)
		}
		if replayElem.waitAck {
			waitForAck = released{
				replayC:    replayElem,
				replayE:    nextElem,
				waitForAck: true,
			}
		} else {
			foundReplayElement()
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

// Release a waiting operation and, if required, wait for the acknowledgement
//
// Parameter:
//   - elem (replayChan): the element to be released
//   - elemReplay (ReplayElement): the corresponding replay element
//   - rel (bool): the waiting channel should be released
//   - next (bool): true if the next element in the trace was released, false if the oldest has been released
//
// Returns:
//   - bool: true if the element was released, false if not, especially if it already has been released before
func releaseElement(elem replayChan, elemReplay ReplayElement, rel, next bool) bool {
	if elem.released {
		if printDebug {
			println("Already released: ", elemReplay.key())
		}
		return false
	}

	if rel {
		elem.chWait <- elemReplay
		elem.released = true
		if printDebug {
			println("Release: ", elemReplay.key())
		}
	}

	if elem.waitAck {
		if printDebug {
			println("Wait Ack: ", elemReplay.key())
		}
		select {
		case <-elem.chAck:
			if printDebug {
				println("Ack: ", elemReplay.key())
			}
		case <-after(sToNs(acknowledgementMaxWaitSec)):
			if printDebug {
				println("AckTimeout: ", elemReplay.key())
			}
			if tPostWhenAckFirstTimeout == 0 {
				tPostWhenAckFirstTimeout = elemReplay.Time
			}
		}
	}

	if printDebug {
		println("Complete: ", elemReplay.key())
	}

	lastTime = currentTime()
	if next {
		lastTimeWithoutOldest = currentTime()
		releaseOldestWait = releaseOldestWaitLastMax
		foundReplayElement()
	} else {
		releasedElementOldest(elemReplay.key())
	}

	return true
}

// When called, the calling routine is blocked and cannot be woken up again
func BlockForever() {
	gopark(nil, nil, waitReasonZero, traceBlockForever, 1)
}

var alreadyExecutedAsOldest = make(map[string]int)

// Get the next element to be executed from the replay trace
//
// Returns:
//   - uint64: the routine of the next replay element or -1 if the trace is empty
//   - ReplayElement: the next replay element
func getNextReplayElement() (int, ReplayElement) {
	lock(&replayLock)
	defer unlock(&replayLock)

	if replayIndex >= numberElementsInTrace{
		return -1, ReplayElement{}
	}

	elem := replayData[replayIndex]

	// if the elem was already executed as an oldest before, do not get again
	elemKey := elem.key()
	if val, ok := alreadyExecutedAsOldest[elemKey]; ok && val > 0 {
		foundReplayElement()
		alreadyExecutedAsOldest[elemKey]--
		return getNextReplayElement()
	}

	return elem.Routine, elem
}

// Release an element as the oldest element event if it is not the operations turn
//
// Parameter:
//   - key string: key of the element to be released
func releasedElementOldest(key string) {
	alreadyExecutedAsOldest[key]++
}

// foundReplayElement is executed if an operation has been executed.
// It advances the index of the replay trace to the next values, such
// that the next element is returned as the next element to be replayed
func foundReplayElement() {
	lock(&replayLock)
	defer unlock(&replayLock)
	replayIndex++
	if printDebug {
		println("Advance: ", replayIndex)
	}
}

// Set replayForceExit
//
// Parameter:
//
//	force: force exit
func SetForceExit(force bool) {
	replayForceExit = force
}

// Set the expected exit code
//
// Parameters:
//   - code: the expected exit code
func SetExpectedExitCode(code int) {
	expectedExitCode = code
}

// Exit the program with the given code.
//
// Parameter:
//   - code: the exit code
func ExitReplayWithCode(code int) {
	if !hasReturnedExitCode {
		// if !isExitCodeConfOnEndElem(code) && !stuckReplayExecutedSuc {
		// 	return
		// }
		println("\nExit Replay with code ", code, ExitCodeNames[code])
		hasReturnedExitCode = true
	} else {
		println("Exit code already returned")
	}

	if replayForceExit && ExitCodeNames[code] != "" {
		// if !advocateTracingDisabled { // do not exit if recording is enabled
		// 	return
		// }
		println("Forcing exit with code ", code, ExitCodeNames[code])
		exit(int32(code))
	} else {
		println("Exit code not set")
		exit(-1)
	}
}

//	For some exit codes, the replay is seen as confirmed, if the replay end
//	element is reached. This function returns wether the exit code is
//	such a code
//	The codes are
//	   20 - 29: Leak
//
// Parameter:
//   - code int: the exit code
//
// Returns:
//   - bool: true if the code is a leak code
func isExitCodeConfOnEndElem(code int) bool {
	return (code >= 20 && code < 30) || (code >= 40 && code < 50)
}

// Exit the program with the given code if the program panics.
//
// Parameter:
//   - msg: the panic message
func ExitReplayPanic(msg any) {
	SetExitCodeFromPanicMsg(msg)

	if IsAdvocateFuzzingEnabled() {
		finishFuzzingFunc()
	} else if IsTracingEnabled() {
		finishTracingFunc()
	}



	// if !IsReplayEnabled() {
	// 	return
	// }

	// ExitReplayWithCode(advocateExitCode)
}

// AdvocateIgnoreReplay decides if an operation should be ignored for replay.
// Ignored means it is just executed when called without waiting.
// All internal operations are ignored
// Atomic operations are ignored if the corresponding variable is set
//
// Parameter:
//   - operation Operation: the operation to check
//   - file string: the file where the operation is executed
//
// Returns:
//   - bool: true if the operation should be ignored, false otherwise
func AdvocateIgnoreReplay(operation Operation, file string) bool {
	if ignoreAtomicsReplay && getOperationObjectString(operation) == "Atomic" {
		return true
	}

	if containsStr(file, "go/pkg/mod/") {
		return true
	}

	return AdvocateIgnore(file)
}

// Return the replay status
//
// Returns:
//   - int: what was the next tPost, when the manager released an oldest for the first time, if never return 0
//   - int: what was the next tPost, when the manager disables the replay because it was stuck, if not return 0
//   - int: what was the next tPost, when an expected Acknowledgement timed out, if never return 0

func GetReplayStatus() (int, int, int) {
	return tPostWhenFirstTimeout, tPostWhenReplayDisabled, tPostWhenAckFirstTimeout
}
