// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_control.go
// Brief: Handle replay control elements
//
// Author: Erik Kassubek
// Created: 2025-04-22
//
// License: BSD-3-Clause

package runtime

// replayControl is run if a replay control element is the next element
// in the trace.
//
// Parameter:
//   - elem ReplayElement: The replay control element
//
// Returns:
//   - bool: if true, the replay manager should be terminated, otherwise
//     it should continue
func replayControl(elem ReplayElement) bool {
	if printDebug {
		println("replayControl")
	}
	defer foundReplayElement(elem.Routine)

	if elem.ControlModeSwitch {
		if printDebug {
			println("Mode switch")
		}
		replayControlModeSwitch(elem)
		return false
	}

	if printDebug {
		println("Exit code")
	}
	replayControlExitCode(elem)
	return true
}

// replayControlModeSwitch changes the replay mode as specified in the replay
// control element.
// For now only the switch from full strict mode to partial strict mode is implemented
// Full strict mode means, that only elements in the trace can be executed
// Partial strict mode means, that for each elem it is checked if it is in the
// remaining mode. If not, it is executed directly
//
// Parameter:
//   - elem ReplayElement: The replay control element
func replayControlModeSwitch(elem ReplayElement) {
	mode := elem.File

	if mode == currentReplayMode {
		return
	}

	if mode == replayModePartial {
		currentReplayMode = replayModePartial

		// release all waiting elements, that are not in the remaining trace
		for key, ops := range waitingOps {
			if !isElementInRemainingTrace(key) {
				if printDebug {
					println("RelNotInRem1: ", key)
				}
				ops.waitAck = false
				releaseElement(ops, replayElemFromKey(key), true, false)
				ops.released = true
			}
		}
	}
}

// replayControlModeSwitch checks for the exit code specified in the replay
// control element and disables replay
//
// Parameter:
//   - elem ReplayElement: The replay control element
func replayControlExitCode(elem ReplayElement) {
	println("Found ReplayEnd Marker with exit code", elem.Line)
	// wait long enough, that all operations that have been released but not
	// finished executing can execute
	if elem.Line == ExitCodeCyclic {
		lock(&waitDeadlockDetectLock)
		waitDeadlockDetect = true
		unlock(&waitDeadlockDetectLock)
	}
	sleep(0.5)

	DisableReplay()
	// foundReplayElement(routine)
	sleep(0.1)

	// Check if a deadlock has been reached
	if elem.Line == ExitCodeCyclic {
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
			ExitReplayWithCode(elem.Line)
		}

		lock(&waitDeadlockDetectLock)
		waitDeadlockDetect = false
		unlock(&waitDeadlockDetectLock)
	} else if isExitCodeConfOnEndElem(elem.Line) {
		ExitReplayWithCode(elem.Line)
	}
}
