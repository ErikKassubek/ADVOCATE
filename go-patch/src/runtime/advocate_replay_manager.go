// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_manager.go
// Brief: Replay manager
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

type replayChan struct {
	chWait   chan ReplayElement
	chAck    chan struct{}
	counter  int
	waitAck  bool
	released bool
}

type released struct {
	replayC    replayChan
	replayE    ReplayElement
	waitForAck bool
}

// Function to run in the background and to release the waiting operations
func ReplayManager() {
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

		if routine == -1 && !partialReplay {
			println("BREAK1")
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
					// SetForceExit(true)
					ExitReplayWithCode(replayElem.Line, "")
				}

				lock(&waitDeadlockDetectLock)
				waitDeadlockDetect = false
				unlock(&waitDeadlockDetectLock)
			} else if isExitCodeConfOnEndElem(replayElem.Line) {
				ExitReplayWithCode(replayElem.Line, "")
			}
			return
		}

		key := replayElem.Key()

		if key == lastKey {
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
					var oldest = replayChan{nil, nil, -1, false, false}
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

				continue
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
			CheckForPartialReplay(replayElem)
			lastKey = key
			if printDebug {
				println("\n\n===================\n")
				println("Next: ", key)
				if i, ok := active[key]; ok {
					print("AC: ")
					for _, a := range i {
						print(a, ", ")
					}
					println("")
				} else {
					print("AC: XX")
				}
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

/*
 * Wait until all operations in the trace are executed.
 * This function should be called after the main routine is finished, to prevent
 * the program to terminate before the trace is finished.
 */
func WaitForReplayFinish(exit bool) {
	if printDebug {
		println("Wait for replay finish")
	}

	startTime := currentTime()

	if IsReplayEnabled() {
		for {
			if replayIndex >= numberElementsInTrace {
				break
			}

			if !replayEnabled {
				break
			}

			if hasTimePast(startTime, 10) {
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
		ExitReplayWithCode(expectedExitCode, "")
	}
}
