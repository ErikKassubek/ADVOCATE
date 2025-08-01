// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_partial.go
// Brief: Partial replay
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

// Check if it is time to switch to active replay, and if so, switch
func CheckForPartialReplay(elemReplay ReplayElement) {
	if partialReplay {
		return
	}
	// switch to replay that only looks for active elements
	if printDebug {
		println("Check for partial replay ", elemReplay.Time, startTimeActive)
	}

	if startTimeActive != -1 && elemReplay.Time >= startTimeActive {
		if printDebug {
			println("Switch to active replay")
		}
		partialReplay = true

		lock(&waitingOpsMutex)

		for key, ops := range waitingOps {
			// the operation is never active
			c, ok := active[key]
			if !ok {
				repEl, _ := getSelect(key)
				ops.chWait <- repEl
				if printDebug {
					println("ReleaseNotActive: ", key)
				}
				delete(waitingOps, key)
				continue
			}

			lock(&partialReplayMutex)
			partialReplayCounter[key] += 1
			currentCounter := partialReplayCounter[key]
			unlock(&partialReplayMutex)

			// the operation is sometimes active, but not this time
			if !isInSlice(c, currentCounter) {
				repEl, _ := getSelect(key)

				ops.chWait <- repEl
				if printDebug {
					println("ReleaseNotActive: ", key)
				}
				delete(waitingOps, key)
			}
		}
		unlock(&waitingOpsMutex)
	}
}

func getSelect(key string) (ReplayElement, bool) {
	re := ReplayElement{Blocked: false}
	found := false
	if s, ok := selects[key]; ok && len(s) > 0 {
		re = s[0]
		selects[key] = selects[key][1:]
		found = true
	}

	return re, found
}

func releaseActive(key string) {
	if _, ok := active[key]; !ok {
		return
	}

	active[key] = active[key][1:]
	if len(active[key]) == 0 {
		delete(active, key)
	}
}
