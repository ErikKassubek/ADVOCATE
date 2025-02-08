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

type VectorClockTID struct {
	Vc      clock.VectorClock
	TID     string
	Routine int
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

type VectorClockTID3 struct {
	Elem TraceElement
	Vc   clock.VectorClock
	Val  int
}

type allSelectCase struct {
	sel          *TraceElementSelect // the select
	chanID       int                 // channel id
	vcTID        VectorClockTID      // vector clock and tID
	send         bool                // true: send, false: receive
	buffered     bool                // true: buffered, false: unbuffered
	partnerFound bool                // true: partner found, false: no partner found
	partner      []VectorClockTID3   // the potential partner
	exec         bool                // true: the case was executed, false: otherwise
	casi         int                 // internal index for the case in the select
}

var (
	// analysis cases to run
	analysisCases = make(map[string]bool)

	// vc of close on channel
	closeData = make(map[int]*TraceElementChannel) // id -> vcTID3 val = ch.id

	// last receive for each routine and each channel
	lastRecvRoutine = make(map[int]map[int]VectorClockTID) // routine -> id -> vcTID

	// most recent send, used for detection of send on closed
	hasSend        = make(map[int]bool)                    // id -> bool
	mostRecentSend = make(map[int]map[int]VectorClockTID3) // routine -> id -> vcTID

	// most recent send, used for detection of received on closed
	hasReceived       = make(map[int]bool)                    // id -> bool
	mostRecentReceive = make(map[int]map[int]VectorClockTID3) // routine -> id -> vcTID3, val = objID

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
	lockSet                = make(map[int]map[int]string)         // routine -> id -> string
	mostRecentAcquire      = make(map[int]map[int]VectorClockTID) // routine -> id -> vcTID
	mostRecentAcquireTotal = make(map[int]VectorClockTID3)        // id -> vcTID

	// vector clocks for last release times
	relW = make(map[int]clock.VectorClock) // id -> vc
	relR = make(map[int]clock.VectorClock) // id -> vc

	// for leak check
	leakingChannels = make(map[int][]VectorClockTID2) // id -> vcTID

	// for check of select without partner
	// store all select cases
	selectCases = make([]allSelectCase, 0)

	// all positions of creations of routines
	allForks = make(map[int]*TraceElementFork) // routineId -> fork

	// set to true if the analyzer has been started, so that it can clean it before reusing it
	DataUsed = false

	// exit code info
	exitCode int
	exitPos  string
)

// InitAnalysis initializes the analysis cases
func InitAnalysis(analysisCasesMap map[string]bool) {
	analysisCases = analysisCasesMap
}

func ClearData() {
	closeData = make(map[int]*TraceElementChannel)
	lastRecvRoutine = make(map[int]map[int]VectorClockTID)
	hasSend = make(map[int]bool)
	mostRecentSend = make(map[int]map[int]VectorClockTID3)
	hasReceived = make(map[int]bool)
	mostRecentReceive = make(map[int]map[int]VectorClockTID3)
	bufferedVCs = make(map[int][]bufferedVC)
	wgAdd = make(map[int][]TraceElement)
	wgDone = make(map[int][]TraceElement)
	allLocks = make(map[int][]TraceElement)
	allUnlocks = make(map[int][]TraceElement)
	lockSet = make(map[int]map[int]string)
	mostRecentAcquire = make(map[int]map[int]VectorClockTID)
	mostRecentAcquireTotal = make(map[int]VectorClockTID3)
	relW = make(map[int]clock.VectorClock)
	relR = make(map[int]clock.VectorClock)
	leakingChannels = make(map[int][]VectorClockTID2)
	selectCases = make([]allSelectCase, 0)
	allForks = make(map[int]*TraceElementFork)
	exitCode = 0
	exitPos = ""
}
