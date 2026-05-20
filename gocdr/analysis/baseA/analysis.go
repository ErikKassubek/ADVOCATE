// gocdr/analysis/baseA/analysis.go

// Copyright (c) 2025 Erik Kassubek
//
// File: vc.go
// Brief: Data required for the trace
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package baseA

import "gocdr/trace"

// MaxNumberElements is the maximum number of elements fow which a HB analysis is run
const MaxNumberElements = 10000000

// MainTrace is the trace that is created and the trace on which most
// normal operations and the analysis is performed
var (
	MainTrace     trace.Trace
	MainTraceIter trace.Iterator
)

// Data required for the analysis
var (
	numberOpsPerID = make(map[int]int) // id -> number of ops on this elem

	HoldSend = make([]HoldObj, 0)
	HoldRecv = make([]HoldObj, 0)

	WaitingReceive = make([]*trace.ElementChannel, 0)
	MaxOpID        = make(map[int]int)

	// vc of close on channel
	CloseData = make(map[int]*trace.ElementChannel) // id -> vcTID3 val = ch.id

	// currently waiting cond var
	CurrentlyWaiting = make(map[int][]*trace.ElementCond) // -> id -> []*trace.ElementCond

	// all positions of creations of routines
	ForkOperations = make(map[int]*trace.ElementFork) // routineId -> fork

	// last change for wait group counters
	LastChangeWG = make(map[int]*trace.ElementWait)

	// currently hold locks
	CurrentlyHoldLock = make(map[int]*trace.ElementMutex) // routine -> lock op

	// vector clocks for the successful do
	OSuc = make(map[int]*trace.ElementOnce)

	ExecutedOnce = make(map[int]*ConcurrentEntry) // id -> elem

	// for check of select without partner
	// store all select cases
	SelectCases                  = make([]AllSelectCase, 0)
	NumberSelectCasesWithPartner int

	// last acquire on mutex for each routine
	LockSet    = make(map[int]map[int]string) // routine -> id -> string
	RLockCount = make(map[int]map[int]int)    // routine -> lockID -> count

	// lock/unlocks on mutexes
	AllLocks   = make(map[int][]trace.Element)
	AllUnlocks = make(map[int][]trace.Element) // id -> []TraceElement

	// add/done on waitGroup
	WGAddData  = make(map[int][]trace.Element) // id  -> []TraceElement
	WgDoneData = make(map[int][]trace.Element) // id -> []TraceElement

	// state for resource deadlock
	CurrentState State

	// last atomic writer for each atomic variable

	// vector clocks for last write times
	LastAtomicWriter = make(map[int]*trace.ElementAtomic)

	// channel creation position
	NewChan = make(map[int]string) // id -> pos
)

// ClearTrace sets the main analysis trace to a new, empty trace
func ClearTrace() {
	MainTrace = trace.NewTrace()
	MainTraceIter = MainTrace.AsIterator()
}

// SetMainTrace sets the main trace to a given trace
//
// Parameter:
//   - t *trace.Trace: the new trace
func SetMainTrace(t *trace.Trace) {
	MainTrace = *t
	MainTraceIter = MainTrace.AsIterator()
}

// ShortenTrace shortens the trace by removing all elements after the given time
//
// Parameter:
//   - time int: The time to shorten the trace to
//   - incl bool: True if an element with the same time should stay included in the trace
func ShortenTrace(time int, incl bool) {
	MainTrace.ShortenTrace(time, incl)
}

// RemoveElementFromTrace removes the element with the given tID from the trace
//
// Parameter:
//   - tID string: The tID of the element to remove
func RemoveElementFromTrace(tID string) {
	MainTrace.RemoveElementFromTrace(tID)
}

// ShortenRoutine shortens the trace of the given routine by removing all
// elements after and equal the given time
//
// Parameter:
//   - routine int: The routine to shorten
//   - time int: The time to shorten the trace to
func ShortenRoutine(routine int, time int) {
	MainTrace.ShortenRoutine(routine, time)
}

// GetRoutineTrace returns the trace of the given routine
//
// Parameter:
//   - id int: The id of the routine
//
// Returns:
//   - []traceElement: The trace of the routine
func GetRoutineTrace(id int) []trace.Element {
	return MainTrace.GetRoutineTrace(id)
}

// ShortenRoutineIndex a given a routine to index
//
// Parameter:
//   - routine int: the routine to shorten
//   - index int: the index to which it should be shortened
//   - incl bool: if true, the value a index will remain in the routine, otherwise it will be removed
func ShortenRoutineIndex(routine, index int, incl bool) {
	MainTrace.ShortenRoutineIndex(routine, index, incl)
}

// GetNoRoutines is a getter for the number of routines
//
// Returns:
//   - int: The number of routines
func GetNoRoutines() int {
	return MainTrace.GetNoRoutines()
}

// GetTraceLength returns the number of elements in the trace of a given routine
//
// Parameter:
//   - int: the routine id
//
// Returns:
//   - []int: number of elements in routines.
func GetTraceLength(routine int) int {
	return MainTrace.GetTraceLength(routine)
}

// GetTraceLengths returns a slice containing the number of elements in the
// routines
//
// Returns:
//   - []int: number of elements in routines.
func GetTraceLengths() []int {
	return MainTrace.GetTraceLengths()
}

// GetLastElemPerRout returns the last elements in each routine
// Returns
//
//   - []TraceElements: List of elements that are the last element in a routine
func GetLastElemPerRout() []trace.Element {
	return MainTrace.GetLastElemPerRout()
}

// GetTraceElementFromTID returns the routine and index of the element
// in trace, given the tID
//
// Parameter:
//   - tID string: The tID of the element
//
// Returns:
//   - TraceElement: The element
//   - error: An error if the element does not exist
func GetTraceElementFromTID(tID string) (trace.Element, error) {
	return MainTrace.GetTraceElementFromTID(tID)
}

// GetTraceElementFromBugArg return the element in the trace, that correspond
// to the element in a bug argument.
//
// Parameter:
//   - bugArg string: The bug info from the machine readable result file
//
// Returns:
//   - *TraceElement: The element
//   - error: An error if the element does not exist
func GetTraceElementFromBugArg(bugArg string) (trace.Element, error) {
	return MainTrace.GetTraceElementFromBugArg(bugArg)
}

// RemoveLater removes all elements that have a later tPost that the given tPost
//
// Parameter:
//   - tPost int: Remove elements after tPost
func RemoveLater(tPost int) {
	MainTrace.RemoveLater(tPost)
}

// ShiftRoutine shifts all elements with time greater or equal to startTSort by shift
// Only shift back
//
// Parameter:
//   - routine int: The routine to shift
//   - startTSort int: The time to start shifting
//   - shift int: The shift
//
// Returns:
//   - bool: True if the shift was successful, false otherwise (shift <= 0)
func ShiftRoutine(routine int, startTSort int, shift int) bool {
	return MainTrace.ShiftRoutine(routine, startTSort, shift)
}

// GetPartialTrace returns the partial trace of all element between startTime
// and endTime inclusive.
//
// Parameter:
//   - startTime int: The start time
//   - endTime int: The end time
//
// Returns:
//   - map[int][]TraceElement: The partial trace
func GetPartialTrace(startTime int, endTime int) map[int][]trace.Element {
	return MainTrace.GetPartialTrace(startTime, endTime)
}

// SortTrace sorts each routine of the trace by tPost
func SortTrace() {
	MainTrace.Sort()
}

// CopyMainTrace returns a copy of the current main trace
//
// Returns:
//   - Trace: The copy of the trace
//   - error
func CopyMainTrace() (trace.Trace, error) {
	return MainTrace.Copy(true)
}

// SetTrace sets the main trace
//
// Parameter:
//   - trace Trace: The trace
func SetTrace(trace trace.Trace) {
	MainTrace = trace
}

// PrintTrace prints the main trace sorted by tPost
func PrintTrace() {
	MainTrace.PrintTrace()
}

// AddOpsPerID increases the counter for numberOpsPerId by one for a given element id
// Do not call for forks
//
// Parameter:
//   - id int: the id of the element
func AddOpsPerID(id int) {
	numberOpsPerID[id]++
}

// GetOpsPerID how often an operations has been performed on a given element
//
// Parameter:
//   - id int: the id of the element
//
// Returns:
//   - int: how often an operations has been executed on the element. Return 0
//     if id does not exists
//   - bool: true if operation on id exists, false otherwise
func GetOpsPerID(id int) (int, bool) {
	if count, ok := numberOpsPerID[id]; ok {
		return count, true
	}
	return 0, false
}
