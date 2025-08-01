// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_wait.go
// Brief: Wait for replay
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

var mgcRoutine uint64 = 0

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
//   - bool: true if released because not active, false otherwise
func WaitForReplay(op Operation, skip int, waitForResponse bool) (bool, chan ReplayElement, chan struct{}, bool) {
	// without this the runtime build runs into an deadlock message, no idea why
	if !replayEnabled {
		return false, nil, nil, false
	}

	// if Caller is run in the garbage collector, the execution stops
	// this should prevent this
	if currentGoRoutineInfo() != nil && mgcRoutine != 0 && currentGoRoutineInfo().id == mgcRoutine {
		return false, nil, nil, false
	}

	_, file, line, _ := Caller(skip)

	if currentGoRoutineInfo() != nil && containsStr(file, "/src/runtime/mgc.go") {
		mgcRoutine = currentGoRoutineInfo().id
	}

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
//   - bool: true if released because not active, false otherwise
func WaitForReplayPath(op Operation, file string, line int, waitForResponse bool) (bool, chan ReplayElement, chan struct{}, bool) {
	if !replayEnabled {
		return false, nil, nil, false
	}

	if AdvocateIgnoreReplay(op, file) {
		return false, nil, nil, false
	}

	routine := GetReplayRoutineID()

	key := BuildReplayKey(routine, file, line)

	lock(&partialReplayMutex)
	partialReplayCounter[key] += 1
	currentCounter := partialReplayCounter[key]
	unlock(&partialReplayMutex)

	chWait := make(chan ReplayElement, 1)
	chAck := make(chan struct{}, 1)

	// ignore not active operations if partial replay is active
	if partialReplay {
		// the operation is never active
		c, ok := active[key]
		if !ok {
			repEl, found := getSelect(key)

			chWait <- repEl

			if printDebug {
				println("ReleaseNonActiveN: ", key)
			}
			lastKey = ""
			return found, chWait, nil, true
		}

		// the operation is sometimes active, but not this time
		if !isInSlice(c, currentCounter) {
			if printDebug {
				print("ReleaseNonActiveT: ", key)
				for _, d := range c {
					print(", ", d, ", ")
				}
				println(currentCounter)
			}
			repEl, found := getSelect(key)

			chWait <- repEl

			lastKey = ""
			return found, chWait, nil, true
		}

	}

	if printDebug {
		if partialReplay {
			for _, c := range active[key] {
				print(", ", c, ", ")
			}
			println("PRC: ", partialReplayCounter[key])
		}
	}

	replayElem := replayChan{chWait, chAck, counter, waitForResponse, false}

	_, nextElem := getNextReplayElement()
	nextElemKey := nextElem.Key()
	if key == nextElemKey && !waitForAck.waitForAck {
		_, _ = getSelect(key)
		// if it is the next element, release directly and add elems to waitForAck
		chWait <- nextElem
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

		CheckForPartialReplay(nextElem)
	}

	return true, chWait, chAck, false
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
			println("Already released: ", elemReplay.Key())
		}
		return false
	}

	if rel {
		key := elemReplay.Key()
		_, _ = getSelect(key)
		elem.chWait <- elemReplay
		elem.released = true
		if printDebug {
			println("Release: ", elemReplay.Key())
		}
	}

	if elem.waitAck {
		if printDebug {
			println("Wait Ack: ", elemReplay.Key())
		}
		select {
		case <-elem.chAck:
			if printDebug {
				println("Ack: ", elemReplay.Key())
			}
		case <-after(sToNs(acknowledgementMaxWaitSec)):
			if printDebug {
				println("AckTimeout: ", elemReplay.Key())
			}
			if tPostWhenAckFirstTimeout == 0 {
				tPostWhenAckFirstTimeout = elemReplay.Time
			}
		}
	}

	if printDebug {
		println("Complete: ", elemReplay.Key())
	}

	lastTime = currentTime()
	if next {
		lastTimeWithoutOldest = currentTime()
		releaseOldestWait = releaseOldestWaitLastMax
		foundReplayElement()
	} else {
		releasedElementOldest(elemReplay.Key())
	}

	return true
}
