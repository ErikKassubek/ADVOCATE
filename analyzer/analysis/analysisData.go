// Copyright (c) 2024 Erik Kassubek
//
// File: analysisData.go
// Brief: Variables and data for the analysis
//
// Author: Erik Kassubek
// Created: 2024-01-27
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
)

// elemWithVc is a helper element for an element with an additional vector clock
type elemWithVc struct {
	vc   *clock.VectorClock
	elem TraceElement
}

// VectorClockTID2 is a helper to store the relevant elements of a
// trace element without needing to store the element itself
type VectorClockTID2 struct {
	routine  int
	id       int
	vc       *clock.VectorClock
	tID      string
	typeVal  int
	val      int
	buffered bool
	sel      bool
	selID    int
}

// ElemWithVcVal is a helper element for an element with an additional vector clock
// and an additional int val
type ElemWithVcVal struct {
	Elem TraceElement
	Vc   *clock.VectorClock
	Val  int
}

// allSelectCase is a helper element to store individual references to all
// select cases in a trace
type allSelectCase struct {
	sel          *TraceElementSelect // the select
	chanID       int                 // channel id
	elem         elemWithVc          // vector clock and tID
	send         bool                // true: send, false: receive
	buffered     bool                // true: buffered, false: unbuffered
	partnerFound bool                // true: partner found, false: no partner found
	partner      []ElemWithVcVal     // the potential partner
	exec         bool                // true: the case was executed, false: otherwise
	casi         int                 // internal index for the case in the select
}

// ConcurrentEntryType is an enum type used in ConcurrentEntry
type ConcurrentEntryType int

// Possible values for ConcurrentEntryType
const (
	CEOnce ConcurrentEntryType = iota
	CEMutex
	CESend
	CERecv
)

// ConcurrentEntry is a helper element to store elements relevant for
// flow fuzzing
type ConcurrentEntry struct {
	Elem    TraceElement
	Counter int
	Type    ConcurrentEntryType
}

// Values for the possible program exit codes
const (
	ExitCodeNone             = -1
	ExitCodePanic            = 3
	ExitCodeTimeout          = 10
	ExitCodeLeakUnbuf        = 20
	ExitCodeLeakBuf          = 21
	ExitCodeLeakMutex        = 22
	ExitCodeLeakCond         = 23
	ExitCodeLeakWG           = 24
	ExitCodeSendClose        = 30
	ExitCodeRecvClose        = 31
	ExitCodeCloseClose       = 32
	ExitCodeCloseNil         = 33
	ExitCodeNegativeWG       = 34
	ExitCodeUnlockBeforeLock = 35
	ExitCodeCyclic           = 41
)

var (
	// trace data

	// MainTrace is the trace that is created and the trace on which most
	// normal operations and the analysis is performed
	MainTrace Trace

	// current happens before vector clocks
	currentVC = make(map[int]*clock.VectorClock)

	// current must happens before vector clocks
	currentWVC = make(map[int]*clock.VectorClock)

	// channel without partner in main trace
	channelWithoutPartner = make(map[int]map[int]*TraceElementChannel) // id -> opId -> element

	fifo          bool
	modeIsFuzzing bool

	// analysis cases to run
	analysisCases   = make(map[string]bool)
	analysisFuzzing = false

	// vc of close on channel
	closeData = make(map[int]*TraceElementChannel) // id -> vcTID3 val = ch.id

	// last send/receive for each routine and each channel
	lastSendRoutine = make(map[int]map[int]elemWithVc) // routine -> id -> vcTID
	lastRecvRoutine = make(map[int]map[int]elemWithVc) // routine -> id -> vcTID

	// most recent send, used for detection of send on closed
	hasSend        = make(map[int]bool)                  // id -> bool
	mostRecentSend = make(map[int]map[int]ElemWithVcVal) // routine -> id -> vcTID

	// most recent send, used for detection of received on closed
	hasReceived       = make(map[int]bool)                  // id -> bool
	mostRecentReceive = make(map[int]map[int]ElemWithVcVal) // routine -> id -> vcTID3, val = objID

	// vector clock for each buffer place in vector clock
	// the map key is the channel id. The slice is used for the buffer positions
	bufferedVCs = make(map[int]([]bufferedVC))
	// the current buffer position
	bufferedVCsCount = make(map[int]int)
	bufferedVCsSize  = make(map[int]int)

	// add/done on waitGroup
	wgAdd  = make(map[int][]TraceElement) // id  -> []TraceElement
	wgDone = make(map[int][]TraceElement) // id -> []TraceElement
	// wait on waitGroup
	// wgWait = make(map[int]map[int][]VectorClockTID) // id -> routine -> []vcTID

	// lock/unlocks on mutexes
	allLocks   = make(map[int][]TraceElement)
	allUnlocks = make(map[int][]TraceElement) // id -> []TraceElement

	// last acquire on mutex for each routine
	lockSet                = make(map[int]map[int]string)     // routine -> id -> string
	currentlyHoldLock      = make(map[int]*TraceElementMutex) // routine -> lock op
	mostRecentAcquire      = make(map[int]map[int]elemWithVc) // routine -> id -> vcTID
	mostRecentAcquireTotal = make(map[int]ElemWithVcVal)      // id -> vcTID

	// vector clocks for last release times
	relW = make(map[int]*clock.VectorClock) // id -> vc
	relR = make(map[int]*clock.VectorClock) // id -> vc

	// vector clocks for last write times
	lw = make(map[int]*clock.VectorClock)

	// for leak check
	leakingChannels = make(map[int][]VectorClockTID2) // id -> vcTID

	// for check of select without partner
	// store all select cases
	selectCases = make([]allSelectCase, 0)

	// all positions of creations of routines
	allForks = make(map[int]*TraceElementFork) // routineId -> fork

	// currently waiting cond var
	currentlyWaiting = make(map[int][]int) // -> id -> []routine

	// vector clocks for the successful do
	oSuc = make(map[int]*clock.VectorClock)

	// vector clock for each wait group
	lastChangeWG = make(map[int]*clock.VectorClock)

	// exit code info
	exitCode int
	exitPos  string

	// replay timeout info
	replayTimeoutOldest   int
	replayTimeoutDisabled int
	replayTimeoutAck      int

	// for fuzzing flow
	fuzzingFlowOnce  = make([]ConcurrentEntry, 0)
	fuzzingFlowMutex = make([]ConcurrentEntry, 0)
	fuzzingFlowSend  = make([]ConcurrentEntry, 0)
	fuzzingFlowRecv  = make([]ConcurrentEntry, 0)

	executedOnce = make(map[int]*ConcurrentEntry) // id -> elem

	fuzzingCounter = make(map[int]map[string]int) // id -> pos -> counter

	holdSend = make([]holdObj, 0)
	holdRecv = make([]holdObj, 0)

	numberSelectCasesWithPartner int

	durationInSeconds = -1 // the duration of the recording in seconds

	waitingReceive = make([]*TraceElementChannel, 0)
	maxOpID        = make(map[int]int)
)

// SetExitInfo stores the exit code and exit position of a run
//
// Parameter:
//   - code int: the exit code
//   - pos string: the exit position
func SetExitInfo(code int, pos string) {
	exitCode = code
	exitPos = pos
}

// SetReplayTimeoutInfo stores information about wether a run that was guided
// by replay (especially in GoPie fuzzing) had a timeout
//
// Parameter:
//
//   - oldest int: the timer when the the replay released the oldest waiting
//
//     or the current next for the first time, if never it should be 0
//
//   - disabled int: the timer when the the replay was so stuck, that the
//     replay had to be disabled for the first time, if never it should be 0
//
//   - ack int: the timer when the the replay timed out on an acknowledgement,
//     if never it should be 0
func SetReplayTimeoutInfo(oldest, disabled, ack int) {
	replayTimeoutOldest = oldest
	replayTimeoutDisabled = disabled
	replayTimeoutAck = ack
}

// GetTimeoutHappened return if any kind of timeout happened
// A timeout happened if at least one of the three timeout var is not 0
//
// Returns:
//   - - bool: true if a timeout happened, false otherwise
func GetTimeoutHappened() bool {
	return (replayTimeoutOldest + replayTimeoutDisabled + replayTimeoutAck) != 0
}

// SetRuntimeDurationSec is a setter for durationInSeconds
//
// Parameter:
//   - sec int: the runtime duration of a run in second
func SetRuntimeDurationSec(sec int) {
	durationInSeconds = sec
}

// GetRuntimeDurationInSec is a getter for durationInSeconds
//
// Returns:
//   - int: the runtime duration of a run in second
func GetRuntimeDurationInSec() int {
	return durationInSeconds
}

// InitAnalysis initializes the analysis by setting the analysis cases and fuzzing
//
// Parameters:
//   - analysisCasesMap map[string]bool: map with information about which
//     analysis parts should be run
//   - anaFuzzing bool: true if fuzzing, false otherwise
func InitAnalysis(analysisCasesMap map[string]bool, anaFuzzing bool) {
	analysisCases = analysisCasesMap
	analysisFuzzing = anaFuzzing
}

// SetMainTraceToNewTrace sets the main analysis trace to a new, empty trace
func SetMainTraceToNewTrace() {
	MainTrace = NewTrace()
}

// ===========  Helper function for trace operations on the main trace ==========

// AddElementToTrace adds an element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - element TraceElement: The element to add
func AddElementToTrace(element TraceElement) {
	MainTrace.AddElement(element)
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
func GetTraceElementFromTID(tID string) (TraceElement, error) {
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
func GetTraceElementFromBugArg(bugArg string) (TraceElement, error) {
	return MainTrace.GetTraceElementFromBugArg(bugArg)
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
func GetRoutineTrace(id int) []TraceElement {
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

// SetNoRoutines is a setter for the number of routines
//
// Parameter:
//   - n int: The number of routines
func SetNoRoutines(n int) {
	MainTrace.SetNoRoutines(n)
}

// GetNoRoutines is a getter for the number of routines
//
// Returns:
//   - int: The number of routines
func GetNoRoutines() int {
	return MainTrace.GetNoRoutines()
}

// Get the next element from a trace
// Update the current index of the trace
//
// Returns:
//   - (TraceElement) The element in the trace with the smallest TSort that
//   - has not been returned yet
func getNextElement() TraceElement {
	return MainTrace.getNextElement()
}

// Get the last elements in each routine
// Returns
//
//   - []TraceElements: List of elements that are the last element in a routine
func getLastElemPerRout() []TraceElement {
	return MainTrace.getLastElemPerRout()
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

// ShiftConcurrentOrAfterToAfter shifts all elements that are concurrent or
// HB-later than the element such that they are after the element without
// changing the order of these elements
//
// Parameter:
//   - element traceElement: The element
func ShiftConcurrentOrAfterToAfter(element TraceElement) {
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
func ShiftConcurrentOrAfterToAfterStartingFromElement(element TraceElement, start int) {
	MainTrace.ShiftConcurrentOrAfterToAfterStartingFromElement(element, start)
}

// ShiftConcurrentToBefore shifts the element to be after all elements,
// that are concurrent to it
//
// Parameter:
//   - element traceElement: The element
func ShiftConcurrentToBefore(element TraceElement) {
	MainTrace.ShiftConcurrentToBefore(element)
}

// RemoveConcurrent removes all elements that are concurrent to the element
// and have time greater or equal to tMin
//
// Parameter:
//   - element traceElement: The element
//   - tMin int: the minimum time
func RemoveConcurrent(element TraceElement, tMin int) {
	MainTrace.RemoveConcurrent(element, tMin)
}

// RemoveConcurrentOrAfter removes all elements that are concurrent to the
// element or must happen after the element
//
// Parameter:
//   - element traceElement: The element
//   - tMin int: the minimum time
func RemoveConcurrentOrAfter(element TraceElement, tMin int) {
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
func GetConcurrentEarliest(element TraceElement) map[int]TraceElement {
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
func GetPartialTrace(startTime int, endTime int) map[int][]TraceElement {
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
func CopyMainTrace() Trace {
	return MainTrace.Copy()
}

// SetTrace sets the main trace
//
// Parameter:
//   - trace Trace: The trace
func SetTrace(trace Trace) {
	MainTrace = trace
}

// PrintTrace prints the main trace sorted by tPost
func PrintTrace() {
	MainTrace.PrintTrace()
}

// HBWasCalc returns if the hb vector clocks have been calculated for the current trace
//
// Returns:
//   - hbWasCalc of the main trace
func HBWasCalc() bool {
	return MainTrace.hbWasCalc
}

// numberElemsInTrace returns how many elements are in a given routine of the main trace
//
// Parameter:
//   - routine int: routine to check for
//
// Returns:
//   - number of elements in routine
func numberElemsInTrace(routine int) int {
	return MainTrace.numberElemsInTrace[routine]
}
