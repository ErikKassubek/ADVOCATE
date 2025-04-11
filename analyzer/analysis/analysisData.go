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

type elemWithVc struct {
	vc   *clock.VectorClock
	elem TraceElement
}

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

type ElemWithVcVal struct {
	Elem TraceElement
	Vc   *clock.VectorClock
	Val  int
}

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

type ConcurrentEntryType int

const (
	CEOnce ConcurrentEntryType = iota
	CEMutex
	CESend
	CERecv
)

type ConcurrentEntry struct {
	Elem    TraceElement
	Counter int
	Type    ConcurrentEntryType
}

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

	// The MainTrace
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

	// add/dones on waitGroup
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

func SetExitInfo(code int, pos string) {
	exitCode = code
	exitPos = pos
}

func SetReplayTimeoutInfo(oldest, disabled, ack int) {
	replayTimeoutOldest = oldest
	replayTimeoutDisabled = disabled
	replayTimeoutAck = ack
}

/*
 * Return if a timeout happened
 * A timeout happened if at least one of the two timeout var is not 0
 */
func GetTimeoutHappened() bool {
	return (replayTimeoutOldest + replayTimeoutDisabled + replayTimeoutAck) != 0
}

func SetRuntimeDurationSec(sec int) {
	durationInSeconds = sec
}

func GetRuntimeDurationInSec() int {
	return durationInSeconds
}

func InitAnalysis(analysisCasesMap map[string]bool, anaFuzzing bool) {
	analysisCases = analysisCasesMap
	analysisFuzzing = anaFuzzing
}

/*
 * Set the global trace variable to a new, empty trace
 */
func InitTrace() {
	MainTrace = NewTrace()
}

// ===========  Helper function for trace operations on the main trace ==========

/*
* Add an element to the main trace
* Args:
*   routine (int): The routine id
*   element (TraceElement): The element to add
 */
func AddElementToTrace(element TraceElement) {
	MainTrace.AddElement(element)
}

/*
 * Given the file and line info, return the routine and index of the element
 * in trace.
 * Args:
 * 	tID (string): The tID of the element
 * Returns:
 * 	*TraceElement: The element
 * 	error: An error if the element does not exist
 */
func GetTraceElementFromTID(tID string) (TraceElement, error) {
	return MainTrace.GetTraceElementFromTID(tID)
}

/*
 * Given the bug info from the machine readable result file, return the element
 * in the trace.
 * Args:
 * 	bugArg (string): The bug info from the machine readable result file
 * Returns:
 * 	*TraceElement: The element
 * 	error: An error if the element does not exist
 */
func GetTraceElementFromBugArg(bugArg string) (TraceElement, error) {
	return MainTrace.GetTraceElementFromBugArg(bugArg)
}

/*
 * Shorten the trace by removing all elements after the given time
 * Args:
 * 	time (int): The time to shorten the trace to
 * 	incl (bool): True if an element with the same time should stay included in the trace
 */
func ShortenTrace(time int, incl bool) {
	MainTrace.ShortenTrace(time, incl)
}

/*
 * Remove the element with the given tID from the trace
 * Args:
 * 	tID (string): The tID of the element to remove
 */
func RemoveElementFromTrace(tID string) {
	MainTrace.RemoveElementFromTrace(tID)
}

/*
 * Shorten the trace of the given routine by removing all elements after and equal the given time
 * Args:
 * 	routine (int): The routine to shorten
 * 	time (int): The time to shorten the trace to
 */
func ShortenRoutine(routine int, time int) {
	MainTrace.ShortenRoutine(routine, time)
}

/*
 * Get the trace of the given routine
 * Args:
 * 	id (int): The id of the routine
 * Returns:
 * 	[]traceElement: The trace of the routine
 */
func GetRoutineTrace(id int) []TraceElement {
	return MainTrace.GetRoutineTrace(id)
}

/*
 * Shorten a given a routine to index
 * Args:
 * 	routine (int): the routine to shorten
 * 	index (int): the index to which it should be shortened
 * 	incl (bool): if true, the value a index will remain in the routine, otherwise it will be removed
 */
func ShortenRoutineIndex(routine, index int, incl bool) {
	MainTrace.ShortenRoutineIndex(routine, index, incl)
}

/*
 * Set the number of routines
 * Args:
 * 	n (int): The number of routines
 */
func SetNoRoutines(n int) {
	MainTrace.SetNoRoutines(n)
}

/*
 * Get the number of routines
 * Return:
 * 	(int): The number of routines
 */
func GetNoRoutines() int {
	return MainTrace.GetNoRoutines()
}

/*
 * Get the next element from a trace
 * Update the current index of the trace
 * Returns:
 * 	(TraceElement) The element in the trace with the smallest TSort that
 * 		has not been returned yet
 */
func getNextElement() TraceElement {
	return MainTrace.getNextElement()
}

/*
 * Get the last elements in each routine
 * Returns
 * 	[]TraceElements: List of elements that are the last element in a routine
 */
func getLastElemPerRout() []TraceElement {
	return MainTrace.getLastElemPerRout()
}

/*
 * For a given waitgroup id, get the number of add and done operations that were
 * executed before a given time.
 * Args:
 * 	wgID (int): The id of the waitgroup
 * 	waitTime (int): The time to check
 * Returns:
 * 	int: The number of add operations
 * 	int: The number of done operations
 */
func GetNrAddDoneBeforeTime(wgID int, waitTime int) (int, int) {
	return MainTrace.GetNrAddDoneBeforeTime(wgID, waitTime)
}

/*
 * Shift all elements with time greater or equal to startTSort by shift
 * Only shift forward
 * Args:
 * 	startTPre (int): The time to start shifting
 * 	shift (int): The shift
 */
func ShiftTrace(startTPre int, shift int) bool {
	return MainTrace.ShiftTrace(startTPre, shift)
}

/*
 * Shift all elements that are concurrent or HB-later than the element such
 * that they are after the element without changing the order of these elements
 * Args:
 * 	element (traceElement): The element
 */
func ShiftConcurrentOrAfterToAfter(element TraceElement) {
	MainTrace.ShiftConcurrentOrAfterToAfter(element)
}

/*
 * Shift all elements that are concurrent or HB-later than the element such
 * that they are after the element without changeing the order of these elements
 * Only shift elements that are after start
 * Args:
 * 	element (traceElement): The element
 * 	start (traceElement): The time to start shifting (not including)
 */
func ShiftConcurrentOrAfterToAfterStartingFromElement(element TraceElement, start int) {
	MainTrace.ShiftConcurrentOrAfterToAfterStartingFromElement(element, start)
}

/*
 * Shift the element to be after all elements, that are concurrent to it
 * Args:
 * 	element (traceElement): The element
 */
func ShiftConcurrentToBefore(element TraceElement) {
	MainTrace.ShiftConcurrentToBefore(element)
}

/*
 * Remove all elements that are concurrent to the element and have time greater or equal to tmin
 * Args:
 * 	element (traceElement): The element
 */
func RemoveConcurrent(element TraceElement, tmin int) {
	MainTrace.RemoveConcurrent(element, tmin)
}

/*
 * Remove all elements that are concurrent to the element or must happen after the element
 * Args:
 * 	element (traceElement): The element
 */
func RemoveConcurrentOrAfter(element TraceElement, tmin int) {
	MainTrace.RemoveConcurrentOrAfter(element, tmin)
}

/*
 * For each routine, get the earliest element that is concurrent to the element
 * Args:
 * 	element (traceElement): The element
 * Returns:
 * 	map[int]traceElement: The earliest concurrent element for each routine
 */
func GetConcurrentEarliest(element TraceElement) map[int]TraceElement {
	return MainTrace.GetConcurrentEarliest(element)
}

/*
 * Remove all elements that have a later tPost that the given tPost
 * Args:
 * 	tPost (int): Remove elements after tPost
 */
func RemoveLater(tPost int) {
	MainTrace.RemoveLater(tPost)
}

/*
 * Shift all elements with time greater or equal to startTSort by shift
 * Only shift back
 * Args:
 * 	routine (int): The routine to shift
 * 	startTSort (int): The time to start shifting
 * 	shift (int): The shift
 * Returns:
 * 	bool: True if the shift was successful, false otherwise (shift <= 0)
 */
func ShiftRoutine(routine int, startTSort int, shift int) bool {
	return MainTrace.ShiftRoutine(routine, startTSort, shift)
}

/*
 * Get the partial trace of all element between startTime and endTime incluseve.
 * Args:
 *  startTime (int): The start time
 *  endTime (int): The end time
 * Returns:
 *  map[int][]TraceElement: The partial trace
 */
func GetPartialTrace(startTime int, endTime int) map[int][]TraceElement {
	return MainTrace.GetPartialTrace(startTime, endTime)
}

/*
 * Sort each routine of the trace by tpost
 */
func SortTrace() {
	MainTrace.Sort()
}

/*
 * Copy the current main trace
 * Returns:
 * 	Trace: The copy of the trace
 */
func CopyMainTrace() Trace {
	return MainTrace.Copy()
}

/*
 * Set the main trace
 * Args:
 * 	trace (Trace): The trace
 */
func SetTrace(trace Trace) {
	MainTrace = trace
}

/*
* Print the main trace sorted by tPost
 */
func PrintTrace() {
	MainTrace.PrintTrace()
}

/*
 * Return if the hb vector clocks have been calculated for the current trace
 * Returns:
 * 	hbWasCalc of the main trace
 */
func HBWasCalc() bool {
	return MainTrace.hbWasCalc
}

/*
 * Return how many elements are in a given routine of the main trace
 * Args:
 * 	routine (int): routine to check for
 * Returns:
 * 	number of elements in routine
 */
func numberElemsInTrace(routine int) int {
	return MainTrace.numberElemsInTrace[routine]
}
