// Copyright (c) 2025 Erik Kassubek
//
// File: types.go
// Brief: Types used in analysis data
//
// Author: Erik Kassubek
// Created: 2025-07-01
//
// License: BSD-3-Clause

package data

import (
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"strconv"
	"strings"
)

// ElemWithVc is a helper element for an element with an additional vector clock
type ElemWithVc struct {
	Vc   *clock.VectorClock
	Elem trace.Element
}

// VectorClockTID2 is a helper to store the relevant elements of a
// trace element without needing to store the element itself
type VectorClockTID2 struct {
	Routine  int
	ID       int
	Vc       *clock.VectorClock
	TID      string
	TypeVal  int
	Val      int
	Buffered bool
	Sel      bool
	SelID    int
}

// ElemWithVcVal is a helper element for an element with an additional vector clock
// and an additional int val
type ElemWithVcVal struct {
	Elem trace.Element
	Vc   *clock.VectorClock
	Val  int
}

// AllSelectCase is a helper element to store individual references to all
// select cases in a trace
type AllSelectCase struct {
	Sel          *trace.ElementSelect // the select
	ChanID       int                  // channel id
	Elem         ElemWithVc           // vector clock and tID
	Send         bool                 // true: send, false: receive
	Buffered     bool                 // true: buffered, false: unbuffered
	PartnerFound bool                 // true: partner found, false: no partner found
	Partner      []ElemWithVcVal      // the potential partner
	Exec         bool                 // true: the case was executed, false: otherwise
	Casi         int                  // internal index for the case in the select
}

// ConcurrentEntryType is an enum type used in ConcurrentEntry
type ConcurrentEntryType int

// ConcurrentEntry is a helper element to store elements relevant for
// flow fuzzing
type ConcurrentEntry struct {
	Elem    trace.Element
	Counter int
	Type    ConcurrentEntryType
}

// Possible values for ConcurrentEntryType
const (
	CEOnce ConcurrentEntryType = iota
	CEMutex
	CESend
	CERecv
)

// elements for buffered channel internal vector clock
type BufferedVC struct {
	Occupied bool
	Send     *trace.ElementChannel
}

// holdObj can temporarily hold an channel operations with additional information
// it is used in the case that for a synchronous communication, the recv is
// recorded before the send
type HoldObj struct {
	Ch   *trace.ElementChannel
	Vc   map[int]*clock.VectorClock
	WVc  map[int]*clock.VectorClock
	Fifo bool
}

// State for ressource deadlock detection
type State struct {
	Threads map[ThreadID]Thread // Recording lock dependencies in phase 1
	Cycles  []Cycle             // Computing cycles in phase 2
	Failed  bool                // Analysis failed (encountered unsupported lock action)
}

// Lock dependencies for resource deadlock detection
type LockDependency struct {
	Thread   ThreadID
	Lock     LockID
	Lockset  Lockset
	Requests []LockEvent
}

type Cycle []LockDependency

// Lock dependencies are computed thread-local. We make use of the following structures.
type Thread struct {
	LockDependencies map[LockID][]Dependency
	CurrentLockset   Lockset        // The thread's current lockset.
	ReaderCounter    map[LockID]int // Store how many readers a readlock has
}

// Unfortunately, we can't use double-indexed map of the following form in Go.
// type Deps map[Lock]map[Lockset][]Event
// Hence, we introduce some intermediate structure.
type Dependency struct {
	Lockset  Lockset
	Requests []LockEvent
}

// Representation of vector clocks, events, threads, lock and lockset.

type LockEvent struct {
	ThreadID    ThreadID
	TraceID     string
	LockID      int
	VectorClock *clock.VectorClock
}

type ThreadID int
type LockID struct {
	ID       int
	ReadLock bool
}
type Lockset map[LockID]struct{}

// Lock Dependency methods.

func (l LockDependency) Clone() LockDependency {
	reqs := make([]LockEvent, len(l.Requests))
	for i, r := range l.Requests {
		reqs[i] = r.Clone()
	}
	return LockDependency{
		Thread:   l.Thread,
		Lock:     l.Lock,
		Lockset:  l.Lockset.Clone(),
		Requests: reqs,
	}
}

// Event methods.

func (e LockEvent) Clone() LockEvent {
	return LockEvent{
		ThreadID:    e.ThreadID,
		TraceID:     e.TraceID,
		LockID:      e.LockID,
		VectorClock: e.VectorClock.Copy(),
	}
}

// Lock methods.

func (l LockID) IsRead() bool {
	return l.ReadLock
}

func (l LockID) IsWrite() bool {
	return !l.ReadLock
}

func (l LockID) AddReader(s Thread) {
	s.ReaderCounter[l]++
}

func (l LockID) RemoveReader(s Thread) {
	if !l.HasReaders(s) {
		return
	}
	s.ReaderCounter[l]--
	if s.ReaderCounter[l] <= 0 {
		delete(s.ReaderCounter, l)
	}
}

func (l LockID) HasReaders(s Thread) bool {
	if _, exists := s.ReaderCounter[l]; !exists {
		return false
	}
	return s.ReaderCounter[l] > 0
}

// Check if two locks are equal ignoring whether they are read or write locks.
func (l LockID) EqualsIgnoreRW(other LockID) bool {
	return l.ID == other.ID
}

// Check if two locks are the same and at least one of them is a write lock.
func (l LockID) EqualsCouldBlock(other LockID) bool {
	if !l.EqualsIgnoreRW(other) {
		return false
	}
	return l.IsWrite() || other.IsWrite()
}

// Lockset methods.

func (ls Lockset) Empty() bool {
	return len(ls) == 0

}

func (ls Lockset) Add(x LockID) {
	ls[x] = struct{}{}
}

func (ls Lockset) Remove(x LockID) bool {
	if _, contains := ls[x]; !contains {
		return false
	}
	delete(ls, x)
	return true
}

func (ls Lockset) Clone() Lockset {
	clone := make(Lockset, 0)
	for l := range ls {
		clone[l] = ls[l]
	}
	return clone
}

func (ls Lockset) String() string {
	b := strings.Builder{}
	b.WriteString("Lockset{")
	for l := range ls {
		b.WriteString(strconv.Itoa(int(l.ID)))
	}
	b.WriteString("}")
	return b.String()
}

func (ls Lockset) Equal(ls2 Lockset) bool {
	if len(ls) != len(ls2) {
		return false
	}

	for l := range ls {
		if _, contains := ls2[l]; !contains {
			return false
		}
	}
	return true
}

func (ls Lockset) Disjoint(ls2 Lockset) bool {
	for l := range ls {
		if _, contains := ls2[l]; contains {
			return false
		}
	}
	return true
}

func (ls Lockset) DisjointCouldBlock(ls2 Lockset) bool {
	for l := range ls {
		for l2 := range ls2 {
			if l.EqualsCouldBlock(l2) {
				return false
			}
		}
	}
	return true
}
