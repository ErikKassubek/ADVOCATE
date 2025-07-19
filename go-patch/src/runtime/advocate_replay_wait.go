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

	routine := GetReplayRoutineID()

	key := BuildReplayKey(routine, file, line)

	// ignore not active operations if partial replay is active
	if partialReplay {
		// the operation is never active
		c, ok := active[key]
		if !ok {
			if printDebug {
				println("ReleaseNonActive: ", key)
			}
			return false, nil, nil
		}

		lock(&partialReplayMutex)
		partialReplayCounter[key] += 1
		currentCounter := partialReplayCounter[key]
		unlock(&partialReplayMutex)

		// the operation is sometimes active, but not this time
		if !isInSlice(c, currentCounter) {
			if printDebug {
				println("ReleaseNonActive: ", key)
			}
			return false, nil, nil
		}
	}

	if printDebug {
		println("Wait: ", key)
	}

	chWait := make(chan ReplayElement, 1)
	chAck := make(chan struct{}, 1)

	replayElem := replayChan{chWait, chAck, counter, waitForResponse, false}

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
