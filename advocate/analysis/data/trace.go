// Copyright (c) 2025 Erik Kassubek
//
// File: vc.go
// Brief: Data required for the trace
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package data

import (
	"advocate/analysis/clock"
	"advocate/trace"
)

var (
	// MainTrace is the trace that is created and the trace on which most
	// normal operations and the analysis is performed
	MainTrace     trace.Trace
	MainTraceIter trace.Iterator

	HoldSend = make([]HoldObj, 0)
	HoldRecv = make([]HoldObj, 0)

	WaitingReceive = make([]*trace.ElementChannel, 0)
	MaxOpID        = make(map[int]int)

	// most recent send, used for detection of send on closed
	HasSend        = make(map[int]bool)                  // id -> bool
	MostRecentSend = make(map[int]map[int]ElemWithVcVal) // routine -> id -> vcTID

	// most recent send, used for detection of received on closed
	HasReceived       = make(map[int]bool)                  // id -> bool
	MostRecentReceive = make(map[int]map[int]ElemWithVcVal) // routine -> id -> vcTID3, val = objID

	// vc of close on channel
	CloseData = make(map[int]*trace.ElementChannel) // id -> vcTID3 val = ch.id

	// currently waiting cond var
	CurrentlyWaiting = make(map[int][]int) // -> id -> []routine

	// all positions of creations of routines
	ForkOperations = make(map[int]*trace.ElementFork) // routineId -> fork

	// currently hold locks
	CurrentlyHoldLock = make(map[int]*trace.ElementMutex) // routine -> lock op

	// vector clocks for the successful do
	OSuc = make(map[int]*clock.VectorClock)

	// last send/receive for each routine and each channel
	LastSendRoutine = make(map[int]map[int]ElemWithVc) // routine -> id -> vcTID
	LastRecvRoutine = make(map[int]map[int]ElemWithVc) // routine -> id -> vcTID

	ExecutedOnce = make(map[int]*ConcurrentEntry) // id -> elem

	// for leak check
	LeakingChannels = make(map[int][]VectorClockTID2) // id -> vcTID

	// for check of select without partner
	// store all select cases
	SelectCases                  = make([]AllSelectCase, 0)
	NumberSelectCasesWithPartner int

	// last acquire on mutex for each routine
	LockSet                = make(map[int]map[int]string)     // routine -> id -> string
	MostRecentAcquire      = make(map[int]map[int]ElemWithVc) // routine -> id -> vcTID
	MostRecentAcquireTotal = make(map[int]ElemWithVcVal)      // id -> vcTID

	// lock/unlocks on mutexes
	AllLocks   = make(map[int][]trace.Element)
	AllUnlocks = make(map[int][]trace.Element) // id -> []TraceElement

	// add/done on waitGroup
	WgAdd  = make(map[int][]trace.Element) // id  -> []TraceElement
	WgDone = make(map[int][]trace.Element) // id -> []TraceElement

	// last analyzed element per routine
	LastAnalyzedElementPerRoutine = make(map[int]trace.Element) // routine -> elem

	// state for resource deadlock
	CurrentState State
)

// ClearTrace sets the main analysis trace to a new, empty trace
func ClearTrace() {
	MainTrace = trace.NewTrace()
	MainTraceIter = MainTrace.AsIterator()
}

// HBWasCalc returns if the hb vector clocks have been calculated for the current trace
//
// Returns:
//   - hbWasCalc of the main trace
func HBWasCalc() bool {
	return MainTrace.GetHBWasCalc()
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

// GetLastElemPerRout returns the last elements in each routine
// Returns
//
//   - []TraceElements: List of elements that are the last element in a routine
func GetLastElemPerRout() []trace.Element {
	return MainTrace.GetLastElemPerRout()
}

// GetNrAddDoneBeforeTime returns the number of add and done operations that were
// executed before a given time for a given wait group id,
//
// Parameter:
//   - wgID int: The id of the wait group
//   - waitTime int: The time to check
//
// Returns:
//   - int: The number of add operations
//   - int: The number of done operations
func GetNrAddDoneBeforeTime(wgID int, waitTime int) (int, int) {
	return MainTrace.GetNrAddDoneBeforeTime(wgID, waitTime)
}

// ShiftTrace shifts all elements with time greater or equal to startTSort by shift
// Only shift forward
//
// Parameter:
//   - startTPre int: The time to start shifting
//   - shift int: The shift
func ShiftTrace(startTPre int, shift int) bool {
	return MainTrace.ShiftTrace(startTPre, shift)
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

// ShiftConcurrentOrAfterToAfter shifts all elements that are concurrent or
// HB-later than the element such that they are after the element without
// changing the order of these elements
//
// Parameter:
//   - element traceElement: The element
func ShiftConcurrentOrAfterToAfter(element trace.Element) {
	MainTrace.ShiftConcurrentOrAfterToAfter(element)
}

// ShiftConcurrentOrAfterToAfterStartingFromElement shifts all elements that
// are concurrent or HB-later than the element such
// that they are after the element without changing the order of these elements
// Only shift elements that are after start
//
// Parameter:
//   - element traceElement: The element
//   - start traceElement: The time to start shifting (not including)
func ShiftConcurrentOrAfterToAfterStartingFromElement(element trace.Element, start int) {
	MainTrace.ShiftConcurrentOrAfterToAfterStartingFromElement(element, start)
}

// ShiftConcurrentToBefore shifts the element to be after all elements,
// that are concurrent to it
//
// Parameter:
//   - element traceElement: The element
func ShiftConcurrentToBefore(element trace.Element) {
	MainTrace.ShiftConcurrentToBefore(element)
}

// RemoveConcurrent removes all elements that are concurrent to the element
// and have time greater or equal to tMin
//
// Parameter:
//   - element traceElement: The element
//   - tMin int: the minimum time
func RemoveConcurrent(element trace.Element, tMin int) {
	MainTrace.RemoveConcurrent(element, tMin)
}

// RemoveConcurrentOrAfter removes all elements that are concurrent to the
// element or must happen after the element
//
// Parameter:
//   - element traceElement: The element
//   - tMin int: the minimum time
func RemoveConcurrentOrAfter(element trace.Element, tMin int) {
	MainTrace.RemoveConcurrentOrAfter(element, tMin)
}

// GetConcurrentEarliest returns for each routine the earliest element that
// is concurrent to the parameter element
//
// Parameter:
//   - element traceElement: The element
//
// Returns:
//   - map[int]traceElement: The earliest concurrent element for each routine
func GetConcurrentEarliest(element trace.Element) map[int]trace.Element {
	return MainTrace.GetConcurrentEarliest(element)
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
	return MainTrace.Copy()
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

// numberElemsInTrace returns how many elements are in a given routine of the main trace
//
// Parameter:
//   - routine int: routine to check for
//
// Returns:
//   - number of elements in routine
func numberElemsInTrace(routine int) int {
	return MainTrace.NumberElemInTrace(routine)
}
