// File: goCR_replay_stuck.go
// Brief: Stuck replay
//
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

var alreadyExecutedAsOldest = make(map[string]int)

// Returns all routines for which the wait reason has not changed within checkStuckTime seconds.
// Used to detect if replay of deadlock was successful
//
// Parameters:
//   - checkStuckTime float64: find routines that have been waiting for at least this many seconds
//   - checkStuckIterations int: iterations to check
func checkForStuckRoutines(checkStuckTime float64, checkStuckIterations int) map[uint64]waitReason {
	stuckRoutines := make(map[uint64]waitReason)

	lock(&GoCRRoutinesLock)
	for id, routine := range GoCRRoutines {
		stuckRoutines[id] = routine.G.waitreason
	}
	unlock(&GoCRRoutinesLock)

	// Repeatedly check if wait reason has changed
	for i := 0; i < checkStuckIterations; i++ {
		sleep(checkStuckTime / float64(checkStuckIterations))
		lock(&GoCRRoutinesLock)
		for id, routine := range GoCRRoutines {
			if _, ok := stuckRoutines[id]; ok && routine.G.waitreason != stuckRoutines[id] {
				delete(stuckRoutines, id)
			}
		}
		unlock(&GoCRRoutinesLock)
	}
	return stuckRoutines
}

// Release an element as the oldest element event if it is not the operations turn
//
// Parameter:
//   - key string: key of the element to be released
func releasedElementOldest(key string) {
	alreadyExecutedAsOldest[key]++
}
