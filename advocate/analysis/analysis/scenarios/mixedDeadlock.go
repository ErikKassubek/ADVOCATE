// Copyright (c) 2024 Erik Kassubek
//
// File: analysisMixedDeadlock.go
// Brief: Trace analysis for mixed deadlocks. Currently not used.
//
// Author: Erik Kassubek
// Created: 2024-01-27
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"fmt"
	"strconv"
)

// In Two-Cycle Mixed Deadlock (MDS-2) model, each mixed-deadlock candidate
// is defined as a 4-tuple:
//
//  • MD = ( e, f, acq(x)_e, acq(x)_f )
//
// where:
//  • e, f         are channel operations (send/recv or close/recv)
//  • x            is a lock ID held or recently acquired by both threads
//  • acq(x)_e     is the most recent acquire of x by the sender’s thread
//  • acq(x)_f     is the most recent acquire of x by the receiver’s thread

type mixedCandidate struct {
	LockID      int
	ChanID      int
	SendRoutine int
	RecvRoutine int
	SendEvent   *trace.ElementChannel
	RecvEvent   *trace.ElementChannel
	AcqSend     baseA.ElemWithVc
	AcqRecv     baseA.ElemWithVc
	IsCloseRecv bool
}

// Internal state (reset per analysis)
var mixedCandidates []mixedCandidate

func ResetMixedDeadlockState() {
	mixedCandidates = make([]mixedCandidate, 0)
}

// LockSetAddLock adds a lock to the lockSet of a routine
// and saves the vector clock of the acquire event
//
// Parameter:
//   - mu *trace.ElementMutex: The mutex element representing the lock event
//   - vc *clock.VectorClock:  The vector clock at the time of lock acquisition
func LockSetAddLock(mu *trace.ElementMutex, vc *clock.VectorClock) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	// Identify current routine t and lock identifier x
	routine := mu.GetRoutine()
	id := mu.GetID()

	// Ensure per thread lockset LS(t) exists
	if _, ok := baseA.LockSet[routine]; !ok {
		baseA.LockSet[routine] = make(map[int]string)
	}

	// Ensure per thread most recent acquire Acq(t, x) exists
	if _, ok := baseA.MostRecentAcquire[routine]; !ok {
		baseA.MostRecentAcquire[routine] = make(map[int]baseA.ElemWithVc)
	}
	// Add lock x to lock set: LS(t) ← LS(t) ∪ {x}
	baseA.LockSet[routine][id] = mu.GetTID()

	// Store latest acquire event acq(x)_t with vector clock: Acq(t, x) ← (t, mu, V_mu)
	baseA.MostRecentAcquire[routine][id] = baseA.ElemWithVc{
		Vc:   vc,
		Elem: mu,
	}
}

// LockSetRemoveLock removes a lock from the lockSet of a routine
//
// Parameter:
//   - routine int: The routine ID
//   - lock int:    The lock ID to remove
func LockSetRemoveLock(routine int, lock int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	// Verify lock is in lockSet
	if _, ok := baseA.LockSet[routine][lock]; !ok {
		errorMsg := "Lock " + strconv.Itoa(lock) +
			" not in lockSet for routine " + strconv.Itoa(routine)
		log.Error(errorMsg)
		return
	}

	// Remove lock x: LS(t) ← LS(t) \ {x}
	delete(baseA.LockSet[routine], lock)
}

// CheckForMixedDeadlock analyzes a potential mixed deadlock scenario for the given routines
//
// Both routines can hold multiple locks or have acquired them recently.
// The analysis considers all locks held by either routine.
//
// MDS-2 cases:
//   - MD2-1: both sender and receiver are inside critical sections
//            → x ∈ LS(send) ∩ LS(recv)
//   - MD2-2: sender inside, receiver after its last critical section
//            → x ∈ LS(send), x ∉ LS(recv)
//   - MD2-3: sender after, receiver inside
//            → x ∈ LS(recv), x ∉ LS(send)
//
// Parameter:
//   - routineSend int: The sending routine ID
//   - routineRecv int: The receiving routine ID

func CheckForMixedDeadlock(routineSend int, routineRecv int) {
	log.Info(fmt.Sprintf("[MixedDeadlock] Checking routines %d ↔ %d", routineSend, routineRecv))

	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	// Collect LS(t) for both routines: LS(send) and LS(recv)
	lsSend := baseA.LockSet[routineSend]
	lsRecv := baseA.LockSet[routineRecv]
	if lsSend == nil || lsRecv == nil {
		return
	}

	log.Info(fmt.Sprintf("[MixedDeadlock] lsSend=%v lsRecv=%v", lsSend, lsRecv))

	// Consider union of LS(send) ∪ LS(recv)
	seen := make(map[int]struct{})

	// Iterate over locks in LS(send) (save seen locks to avoid duplicates)
	for lockID := range lsSend {
		seen[lockID] = struct{}{}
		// MD2-1 and MD2-2
		addMixedCandidate(routineSend, routineRecv, lockID, false)
	}

	// Iterate over locks in LS(recv) (skip seen locks))
	for lockID := range lsRecv {
		if _, done := seen[lockID]; !done {
			// MD2-3
			addMixedCandidate(routineSend, routineRecv, lockID, false)
		}
	}

	// Iterate over close events by routineSend and consider all locks in LS(recv)
	for _, closeElem := range baseA.CloseData {
		if closeElem.GetRoutine() == routineSend {
			for lockID := range lsRecv {
				addMixedCandidate(routineSend, routineRecv, lockID, true)
			}
		}
	}
}

// addMixedCandidate attempts to create a mixed-deadlock candidate
// from communication partners sendTid and recvTid (MD1) for the given lockID
//
// Parameter:
//   - sendTid int: The sending routine ID
//   - recvTid int: The receiving routine ID
//   - lockID int:  The lock ID to consider
//   - isClose bool: Whether the send is a close operation
func addMixedCandidate(sendTid, recvTid, lockID int, isClose bool) {
	// Lookup last acquire events
	acqS, okS := baseA.MostRecentAcquire[sendTid][lockID]
	acqR, okR := baseA.MostRecentAcquire[recvTid][lockID]
	if !okS || !okR {
		// At least one routine never acquired the lock in question so no dependency
		return
	}

	// HB Check:
	// - Symmetric Deadlocks (MD2-1): hb.Concurrent
	// - Asymmetric Deadlocks (MD2-2/MD2-3): hb.Before or hb.After
	rel := clock.GetHappensBefore(acqS.Vc, acqR.Vc)

	// Accepty any HB-relation betwween the two lock acquisitions (skip only invalid cases)
	if rel == hb.None {
		return
	}

	// Construct candidate
	cand := mixedCandidate{
		LockID:      lockID,
		SendRoutine: sendTid,
		RecvRoutine: recvTid,
		AcqSend:     acqS,
		AcqRecv:     acqR,
		IsCloseRecv: isClose,
	}

	// For close–recv pattern, attach channel reference
	if isClose {
		for chID, chElem := range baseA.CloseData {
			if chElem.GetRoutine() == sendTid {
				cand.ChanID = chID
				cand.SendEvent = chElem
				break
			}
		}
	}

	mixedCandidates = append(mixedCandidates, cand)
	reportMixedDeadlock(cand)
}

// reportMixedDeadlock generates a warning report for the given mixed deadlock candidate
//
// Parameter:
//   - md mixedCandidate: The mixed deadlock candidate to report
func reportMixedDeadlock(md mixedCandidate) {
	fileS, lineS, tPreS, _ := trace.InfoFromTID(md.AcqSend.Elem.GetTID())
	fileR, lineR, tPreR, _ := trace.InfoFromTID(md.AcqRecv.Elem.GetTID())

	lockLabel := fmt.Sprintf("lock %d", md.LockID)
	var chanLabel string
	if md.ChanID != 0 {
		chanLabel = fmt.Sprintf("channel %d", md.ChanID)
	} else {
		chanLabel = "unknown channel"
	}

	msg := fmt.Sprintf("Potential mixed deadlock on %s with %s", lockLabel, chanLabel)

	stuck := results.TraceElementResult{
		RoutineID: md.RecvRoutine,
		ObjID:     md.LockID,
		ObjType:   "DH",
		File:      fileR,
		Line:      lineR,
		TPre:      tPreR,
	}

	context := []results.ResultElem{
		results.TraceElementResult{
			RoutineID: md.SendRoutine,
			ObjID:     md.LockID,
			ObjType:   "Send/Close",
			File:      fileS,
			Line:      lineS,
			TPre:      tPreS,
		},
		results.TraceElementResult{
			RoutineID: md.RecvRoutine,
			ObjID:     md.LockID,
			ObjType:   "Recv",
			File:      fileR,
			Line:      lineR,
			TPre:      tPreR,
		},
	}

	// Report Potential Mixed Deadlock P06
	results.Result(results.WARNING, helper.PMixedDeadlock, msg, []results.ResultElem{stuck}, "context", context)
	log.Info(msg)
}

// ===============================
// TODO / Future Work:
// ===============================
//
// 1) RWMutex Support (RLock / RUnlock vs Lock / Unlock)
//   - Extend LockSetAddLock() to classify and store lock mode (READ or WRITE)
//     ElementMutex type or name ("RLock", "RUnlock").
//   - Maintain `lastLockMode[routine][lockID]` map as implemented in draft.
//   - In addMixedCandidate(), skip (READ, READ) pairs (no exclusion)
//   - (WRITE, WRITE) and (WRITE, READ) / (READ, WRITE) as potential MDs.
//
// 2) Non-standard Critical Sections (Cross-Goroutine Unlock)
//   - Count / flag "non-standard" CS where LockSetRemoveLock() for statistics
//     logs "Lock X not in lockSet for routine Y".
//
// 3) WMHB (Weak Must-Happen-Before) Integration
//   - Implement WMHB relation for impossible reorder pruning.
//   - `mustHappenBefore(e1,e2)` function using fork/start and atomic relations.
//   - In addMixedCandidate(): skip reporting if WMHB(e1,e2) == true.
//   - Purpose: prevent false positives from thread-creation or atomic constraints.//
//
// 4) Channel Partnering Logic (Buffered / Unbuffered / Close)
//   - Ensure channel analysis layer provides correct routine pairs:
//       • Unbuffered → (sender, receiver)
//       • Buffered   → (true sender, dequeuing receiver)
//       • Close/Recv → (closer, receiver)
//
// 5) Confirm Replay Integration (A10 Detection)
//   - rewriteMixedDeadlock(tr, bug) in rewriter/ to generate trace reorder.
//   - enforce reversed HB order of acq(x)_e and acq(x)_f
//     so that both threads reach cyclic lock–channel wait (confirm A10).
//
// 6) Other
//   - Implement grouping of redundant candidates (MD clustering by lock/channel).
//   - Track vector clock distances for prioritizing replays.
//
// ===============================
