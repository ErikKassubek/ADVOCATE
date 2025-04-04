// Copyrigth (c) 2024 Erik Kassubek
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
	vc   clock.VectorClock
	elem TraceElement
}

type VectorClockTID2 struct {
	routine  int
	id       int
	vc       clock.VectorClock
	tID      string
	typeVal  int
	val      int
	buffered bool
	sel      bool
	selID    int
}

type ElemWithVcVal struct {
	Elem TraceElement
	Vc   clock.VectorClock
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

var (
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
	relW = make(map[int]clock.VectorClock) // id -> vc
	relR = make(map[int]clock.VectorClock) // id -> vc

	// vector clocks for last write times
	lw = make(map[int]clock.VectorClock)

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
	oSuc = make(map[int]clock.VectorClock)

	// vector clock for each wait group
	lastChangeWG = make(map[int]clock.VectorClock)

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

// InitAnalysis initializes the analysis cases
func InitAnalysis(analysisCasesMap map[string]bool, anaFuzzing bool) {
	analysisCases = analysisCasesMap
	analysisFuzzing = anaFuzzing
}
