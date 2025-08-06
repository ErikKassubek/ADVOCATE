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

var NumberActive = 0
var NumberActiveReleased = 0

const timeoutPartialSec = 4

// Check if it is time to switch to active replay, and if so, switch
func CheckForPartialReplay(elemReplay ReplayElement) {
	if PartialReplay {
		return
	}
	// switch to replay that only looks for active element

	if startTimeActive != -1 && elemReplay.Time >= startTimeActive {
		if printDebug {
			println("Switch to active replay")
		}
		PartialReplay = true

		lock(&waitingOpsMutex)

		for key, ops := range waitingOps {
			// the operation is never active
			c, ok := active[key]
			if !ok {
				repEl, _ := getSelect(key)
				ops.chWait <- repEl
				if printDebug {
					println("ReleaseNotActiveN: ", key)
				}
				delete(waitingOps, key)
				continue
			}

			lock(&partialReplayMutex)
			currentCounter := partialReplayCounter[key]
			unlock(&partialReplayMutex)

			// the operation is sometimes active, but not this time
			if !isInSlice(c, currentCounter) {
				repEl, _ := getSelect(key)

				println(currentCounter)

				ops.chWait <- repEl
				if printDebug {
					println("ReleaseNotActiveT: ", key)
				}
				delete(waitingOps, key)
			}
		}
		unlock(&waitingOpsMutex)
	}
}

// For a select, get the ReplayElement with teh preferred case
// Used for non active elements in partial replay
//
// Parameter:
//   - key string: the select element key
//
// Returns:
//   - ReplayElement: the select element
//   - bool: t
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

// Call if an element is released. It will check if this element is in active
// and remove if necessary
//
// Parameter:
//   - key: the key of the element
func releaseActive(key string) {
	if !PartialReplay {
		return
	}

	if _, ok := active[key]; !ok {
		return
	}
	NumberActiveReleased++
	println("RELEASE ACTIVE: ", key)
}
