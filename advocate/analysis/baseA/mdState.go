// advocate/analysis/baseA/mdState.go

// Copyright (c) 2024 Erik Kassubek
//
// File: mdState.go
// Brief: Data structures for the two-phase mixed deadlock analysis.
//
// Author: Ilian Kohl
//
// License: BSD-3-Clause

package baseA

import (
	"advocate/analysis/hb/clock"
	"advocate/trace"
)

// MDLockHistEntry records one acquire or release event
type MDLockHistEntry struct {
	Routine int
	LockID  int
	TraceID string
	VC      *clock.VectorClock
	WVC     *clock.VectorClock
	IsRead  bool
	IsAcq   bool
}

// MDCommHistEntry records one completed channel event
type MDCommHistEntry struct {
	Routine  int
	ChanID   int
	Op       trace.OperationType
	OID      int
	TraceID  string
	VC       *clock.VectorClock
	WVC      *clock.VectorClock
	Elem     *trace.ElementChannel
	Buffered bool
}

// MDState holds all online-recorded evidence for mixed deadlock detection
type MDState struct {
	AcqHist  map[int]map[int][]MDLockHistEntry
	RelHist  map[int]map[int][]MDLockHistEntry
	CommHist map[int]map[int][]MDCommHistEntry
}

// Reset by ResetMixedDeadlockState() before each analysis run
var CurrentMDState MDState

// RecordMDAcquire appends an acquire entry to AcqHist
func RecordMDAcquire(e MDLockHistEntry) {
	if CurrentMDState.AcqHist == nil {
		CurrentMDState.AcqHist = make(map[int]map[int][]MDLockHistEntry)
	}
	if CurrentMDState.AcqHist[e.Routine] == nil {
		CurrentMDState.AcqHist[e.Routine] = make(map[int][]MDLockHistEntry)
	}
	CurrentMDState.AcqHist[e.Routine][e.LockID] = append(
		CurrentMDState.AcqHist[e.Routine][e.LockID], e)
}

// RecordMDRelease appends a release entry to RelHist
func RecordMDRelease(e MDLockHistEntry) {
	if CurrentMDState.RelHist == nil {
		CurrentMDState.RelHist = make(map[int]map[int][]MDLockHistEntry)
	}
	if CurrentMDState.RelHist[e.Routine] == nil {
		CurrentMDState.RelHist[e.Routine] = make(map[int][]MDLockHistEntry)
	}
	CurrentMDState.RelHist[e.Routine][e.LockID] = append(
		CurrentMDState.RelHist[e.Routine][e.LockID], e)
}

// RecordMDCommEvent appends a communication event entry to CommHist
func RecordMDCommEvent(e MDCommHistEntry) {
	if CurrentMDState.CommHist == nil {
		CurrentMDState.CommHist = make(map[int]map[int][]MDCommHistEntry)
	}
	if CurrentMDState.CommHist[e.Routine] == nil {
		CurrentMDState.CommHist[e.Routine] = make(map[int][]MDCommHistEntry)
	}
	CurrentMDState.CommHist[e.Routine][e.ChanID] = append(
		CurrentMDState.CommHist[e.Routine][e.ChanID], e)
}
