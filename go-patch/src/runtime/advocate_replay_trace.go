// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_replay_trace.go
// Brief: Replay trace
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package runtime

type AdvocateReplayTrace []ReplayElement

// Add a routine local replay trace to the replay data.
//
// Parameters:
//   - trace trace: the replay trace
//   - map[int][]int: for each routine replay id store the replay ids of all spawns
//   - map[string][]ReplayElement: for each routPath with a select, store the replay elements
func GetReplayTrace() (*AdvocateReplayTrace, *map[int][]int, *map[string][]ReplayElement) {
	return &replayData, &spawns, &selects
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
//   - startTime int: switch to active replay if the element with time startTime
//     has been replayed. If 0, start with active from the beginning,
//     if -1 never switch to active replay
//   - active map[string][int]: the map of active operations where the map
//     key is equal to the replay element key (buildReplayKey) and value is the
//     list of occurrences when the replay should be active for the element,
//     e.g. if the value is [3, 4], the operation in the key is scheduled by
//     the replay if it is executed the 3rd and 4th time, but not for the
//     1st and 2nd time.
//   - numActive int: number of active elements
func AddActiveTrace(startTime int, activeMap map[string][]int, numActive int) {
	active = activeMap
	startTimeActive = startTime
	if printDebug {
		println("Add active with start time ", startTimeActive, " and ", len(active), " active elements")
	}
	if startTime == 0 {
		PartialReplay = true
	}
	NumberActive = numActive
}

// Get the next element to be executed from the replay trace
//
// Returns:
//   - uint64: the routine of the next replay element or -1 if the trace is empty
//   - ReplayElement: the next replay element
func getNextReplayElement() (int, ReplayElement) {
	lock(&replayLock)
	defer unlock(&replayLock)

	if replayIndex >= numberElementsInTrace {
		return -1, ReplayElement{}
	}

	elem := replayData[replayIndex]

	// if the elem was already executed as an oldest before, do not get again
	elemKey := elem.Key()
	if val, ok := alreadyExecutedAsOldest[elemKey]; ok && val > 0 {
		foundReplayElement()
		alreadyExecutedAsOldest[elemKey]--
		return getNextReplayElement()
	}

	return elem.Routine, elem
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

// foundReplayElement is executed if an operation has been executed.
// It advances the index of the replay trace to the next values, such
// that the next element is returned as the next element to be replayed
func foundReplayElement() {
	lock(&replayLock)
	defer unlock(&replayLock)
	replayIndex++
}
