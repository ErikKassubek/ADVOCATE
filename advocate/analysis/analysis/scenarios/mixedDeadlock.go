// advocate/analysis/analysis/scenarios/mixedDeadlock.go

// Copyright (c) 2025 Erik Kassubek
//
// File: analysisMixedDeadlock.go
// Brief: Trace analysis for mixed (lock + channel) deadlocks.
//
// Author: ChatGPT Integration Draft based on Advocate Architecture
// Created: 2025-10-23
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

// Data Structure: mixedCandidate
// represents potential mixed deadlock detected between
// two routines that share a lock and communicate via a channel
// Tuple: (e, f, acq(x)_e, acq(x)_f)
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

// Resets all temp data for mixed deadlock detection
func ResetMixedDeadlockState() {
	mixedCandidates = make([]mixedCandidate, 0)
}

/*
Lock Acquire Notation:
----------------------------------------------------------------------
Let LS(t) denote the current lockset of thread t (all locks held by t).
Let Acq(t, x) denote the *most recent acquire event* of lock x by thread t,
annotated with its vector clock V_acq(t,x).

Then, for each Lock() event e = acq(x)_t with vector clock V_e,
we update:
	LS(t) 		← LS(t) ∪ {x}
	Acq(t, x) 	← (t, e, V_e)

These structures are later used to form mixed-deadlock candidates
(e, f, acq(x)_e, acq(x)_f), where acq(x)_e ∈ Acq(t_e, x).
*/

// LockSetAddLock adds a lock to the lockSet of a routine
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
	routine := mu.GetRoutine() // t
	id := mu.GetID()           // x

	// Ensure per thread lockset LS(t) exists
	// LS(t): Set of all locks currently heled by thread t
	if _, ok := baseA.LockSet[routine]; !ok {
		baseA.LockSet[routine] = make(map[int]string)
	}

	// Ensure per thread most recent acquire Acq(t, x) exists
	// Acq(t, x): Most recent acquire of lock x by thread t
	if _, ok := baseA.MostRecentAcquire[routine]; !ok {
		baseA.MostRecentAcquire[routine] = make(map[int]baseA.ElemWithVc)
	}

	// Record that thread t currently holds lock x:
	// LS(t) ← LS(t) ∪ {x}
	baseA.LockSet[routine][id] = mu.GetTID()

	// Store the latest acquire event acq(x)_t with its vector clock:
	// Acq(t, x) ← (t, mu, V_mu)
	// V_mu: Represents partial-order timestamp at acquisition
	baseA.MostRecentAcquire[routine][id] = baseA.ElemWithVc{
		Vc:   vc, // vector clock 		V_e = V_acq(t,x)
		Elem: mu, // event reference 	  e = acq(x)_t
	}
}

/*
Lock Release Notation:
----------------------------------------------------------------------
When a lock x is released by thread t, we update:
    LS(t) ← LS(t) \ {x}

Removing x ensures that subsequent channel sends/receives are no longer considered as
being executed under that lock.

The mapping Acq(t, x) (the last acquisition event) is *not* deleted here
It is needed for predictive reasoning in later mixed-deadlock checks (MD2-2, MD2-3)
where the thread has already released x before the partner executes
*/

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

	// Verify that lock x is indeed recorded in LS(t)
	// i.e., that t currently holds x
	if _, ok := baseA.LockSet[routine][lock]; !ok {
		errorMsg := "Lock " + strconv.Itoa(lock) +
			" not in lockSet for routine " + strconv.Itoa(routine)
		log.Error(errorMsg)
		return
	}

	// Remove lock x from LS(t)
	// LS(t) ← LS(t) \ {x}
	delete(baseA.LockSet[routine], lock)
}

/*
Mixed Deadlock Detection
----------------------------------------------------------------------
In Two-Cycle Mixed Deadlock (MDS-2) model, each mixed-deadlock candidate
is defined as a 4-tuple:

    MD = ( e, f, acq(x)_e, acq(x)_f )

where
  • e, f         are channel operations (send/recv or close/recv)
  • x            is a lock ID held or recently acquired by both threads
  • acq(x)_e     is the most recent acquire of x by the sender’s thread
  • acq(x)_f     is the most recent acquire of x by the receiver’s thread

The algorithm searches for pairs (routineSend, routineRecv)
that communicate over a channel and have dependent lock acquisitions
that are concurrent in vector-clock happens-before relation.

Each concurrent pair forms a potential mixed-deadlock candidate (predictive warning)

Handled MDS-2 cases:
  • MD2-1: 		Both sender and receiver inside the same critical section
           		(x ∈ LS(send) ∩ LS(recv))
  • MD2-2: 		Sender inside, receiver after (x ∈ LS(send))
  • MD2-3: 	 	Sender after, receiver inside (x ∈ LS(recv))
  • MD-Close: 	close/recv interaction
*/

// CheckForMixedDeadlock analyzes two routines for potential mixed deadlock
//
// Parameter:
//   - routineSend int: The sending routine ID
//   - routineRecv int: The receiving routine ID
func CheckForMixedDeadlock(routineSend int, routineRecv int) {
	log.Info(fmt.Sprintf("[MixedDeadlock] Checking routines %d ↔ %d", routineSend, routineRecv))

	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	// (1) Collect Lock Sets LS(t) for both routine LS(send) and LS(recv)
	lsSend := baseA.LockSet[routineSend]
	lsRecv := baseA.LockSet[routineRecv]
	if lsSend == nil || lsRecv == nil {
		return
	}

	log.Info(fmt.Sprintf("[MixedDeadlock] lsSend=%v lsRecv=%v", lsSend, lsRecv))

	// (2) Consider union of LS(send) ∪ LS(recv)
	//     Ensures asymmetric MD2-2 (sender-in-CS, receiver-after)
	//     				  and MD2-3 (sender-after, receiver-in-CS) cases are included
	seen := make(map[int]struct{})

	for lockID := range lsSend {
		seen[lockID] = struct{}{}
		// Try to build candidate for each lock x ∈ LS(send)
		addMixedCandidate(routineSend, routineRecv, lockID, false)
	}

	for lockID := range lsRecv {
		if _, done := seen[lockID]; !done {
			// Add remaining locks x ∈ LS(recv) \ LS(send)
			addMixedCandidate(routineSend, routineRecv, lockID, false)
		}
	}

	// (3) Handle close–recv variant:
	//     For each recorded close(c) by routineSend,
	//     create candidates with receiver’s LS(recv) locks
	for _, closeElem := range baseA.CloseData {
		if closeElem.GetRoutine() == routineSend {
			for lockID := range lsRecv {
				addMixedCandidate(routineSend, routineRecv, lockID, true)
			}
		}
	}
}

/*
Mixed Deadlock Candidate Creation
------------------------------------------------------
Given two threads t_s and t_r and a lock x:

    if  Acq(t_s, x) = (t_s, e_s, V_s)
        Acq(t_r, x) = (t_r, e_r, V_r)

    and V_s  ||  V_r   (concurrent vector clocks)

    record mixed-deadlock candidate
	    MD = ( e, f, acq(x)_e, acq(x)_f )

 Otherwise, discard (ordered acquisitions cannot yield re-orderable cycles)

*/

// addMixedCandidate attempts to create a mixed-deadlock candidate
// from the given send and receive routines and lock ID.
// // Parameter:
//   - sendTid int: The sending routine ID
//   - recvTid int: The receiving routine ID
//   - lockID int:  The lock ID to consider
//   - isClose bool: Whether the send is a close operation
func addMixedCandidate(sendTid, recvTid, lockID int, isClose bool) {
	// Lookup last acquire events:
	//     acq_s = Acq(t_s, x)
	//     acq_r = Acq(t_r, x)
	acqS, okS := baseA.MostRecentAcquire[sendTid][lockID]
	acqR, okR := baseA.MostRecentAcquire[recvTid][lockID]
	if !okS || !okR {
		// At least one thread never acquired x → no dependency
		return
	}

	// HB-Check: Vector clocks must be concurrent for acquisitions to be potentially re-orderable
	// 		hb(V_s, V_r) = concurrent
	if clock.GetHappensBefore(acqS.Vc, acqR.Vc) != hb.Concurrent {
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
		for chID, chElem := range baseA.CloseData { // retrieve Channel ID
			if chElem.GetRoutine() == sendTid {
				cand.ChanID = chID
				cand.SendEvent = chElem
				break
			}
		}
	}

	// Store in global candidate set:
	// MD_Candidates ← MD_Candidates ∪ {cand}
	mixedCandidates = append(mixedCandidates, cand)

	// Report
	reportMixedDeadlock(cand)
}

/*
Reporting
------------------------------------------------------
*/

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

	results.Result(results.WARNING, helper.PMixedDeadlock, msg, []results.ResultElem{stuck}, "context", context)
	log.Info(msg)
}
