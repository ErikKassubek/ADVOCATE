// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/analysis/analysis/scenarios/mixedDeadlock.go
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
)

/*
******************************************************
// Two-Cycle Mixed Deadlock (MDS-2) model
******************************************************

MD-candidate defined as a 4-tuple: MD = ( e, f, acq(x)_e, acq(x)_f )
- e, f         Channel operations (send/recv, close/recv)
- x            Common lock held / recently acquired by both routines
- acq(x)_e     Senders MostRecentAcquire of x
- acq(x)_f     Receivers MostRecentAcquire of x
*/

type mixedCandidate struct {
	LockID      int
	ChanID      int
	SendRoutine int
	RecvRoutine int
	SendEvent   *trace.ElementChannel
	RecvEvent   *trace.ElementChannel
	AcqSend     baseA.ElemWithVc
	AcqRecv     baseA.ElemWithVc
	Case        string // MD2-1 | MD2-2 | MD2-3 | MD2-XClose
	Buffered    bool
}

var mixedCandidates []mixedCandidate
var lastLockType = make(map[int]map[int]string)

func ResetMixedDeadlockState() {
	mixedCandidates = make([]mixedCandidate, 0)
	lastLockType = make(map[int]map[int]string)
}

/*
************************************************
// LOCKSET TRACKING
************************************************
*/

// Ensure maps per thread lockset LS(t), MostRecentAcquire Acq(t,x) and LockType
func ensureLockTracking(routine int) {
	if _, ok := baseA.LockSet[routine]; !ok {
		baseA.LockSet[routine] = make(map[int]string)
	}
	if _, ok := baseA.MostRecentAcquire[routine]; !ok {
		baseA.MostRecentAcquire[routine] = make(map[int]baseA.ElemWithVc)
	}
	if _, ok := baseA.MostRecentRelease[routine]; !ok {
		baseA.MostRecentRelease[routine] = make(map[int]baseA.ElemWithVc)
	}
	if _, ok := lastLockType[routine]; !ok {
		lastLockType[routine] = make(map[int]string)
	}
}

// Lock type READ/WRITE
func getLockTypeFromAcquireElem(e trace.Element) string {
	if mu, ok := e.(*trace.ElementMutex); ok {
		switch mu.GetType(true) {
		case trace.MutexRLock, trace.MutexTryRLock:
			return "READ"
		default:
			return "WRITE"
		}
	}
	return "UNKNOWN"
}

// Adds lock to routine's LS() and saves VC of Acq
func LockSetAddLock(mu *trace.ElementMutex, vc *clock.VectorClock) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	// Get routine that executed lock mu
	routine := mu.GetRoutine()
	id := mu.GetID()

	// Init LockTracking maps
	ensureLockTracking(routine)

	// Determine lock type
	mode := "WRITE"
	switch mu.GetType(true) {
	case trace.MutexRLock, trace.MutexTryRLock:
		mode = "READ"
	}

	// Add lock to Lockset LS(t) <- LS(t) ∪ {x} with mode
	baseA.LockSet[routine][id] = mu.GetTID()
	lastLockType[routine][id] = mode

	// Store latest acq(x)_t event with VC: Acq(t, x) <- (t, mu, V_mu)
	baseA.MostRecentAcquire[routine][id] = baseA.ElemWithVc{
		Vc:   vc.Copy(),
		Elem: mu,
	}
}

// Removes lock from routine lockSet
func LockSetRemoveLock(mu *trace.ElementMutex, vc *clock.VectorClock) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	routine := mu.GetRoutine()
	id := mu.GetID()

	ensureLockTracking(routine)

	if _, ok := baseA.LockSet[routine][id]; !ok {
		log.Error(fmt.Sprintf("Lock %d not in lockSet for routine %d", id, routine))
		return
	}

	baseA.MostRecentRelease[routine][id] = baseA.ElemWithVc{
		Vc:   vc.Copy(),
		Elem: mu,
	}

	// LS(t) <- LS(t) \ {x}
	delete(baseA.LockSet[routine], id)

	//delete(lastLockType[routine], id)
}

/*
************************************************
// CHECK MD-SCENARIO
************************************************
MDS-2 cases:
- MD2-1: both sender and receiver inside CS:  x ∈ LS(send) ∩ LS(recv)
- MD2-2: sender inside CS, receiver with PCS: x ∈ LS(send), x ∉ LS(recv)
- MD2-3: sender with PCS, receiver inside CS: x ∈ LS(recv), x ∉ LS(send)
*/

// Analyzes MD-scenario for given routines called by elements/channel.go
func CheckForMixedDeadlock(routineSend int, routineRecv int) {
	log.Info(fmt.Sprintf("[MixedDeadlock] Checking routines %d ↔ %d", routineSend, routineRecv))

	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	// Union: LS(send) ∪ LS(recv) (seen locks to avoid duplicates)
	seen := make(map[int]struct{})
	if ma := baseA.MostRecentAcquire[routineSend]; ma != nil {
		for lockID := range ma {
			seen[lockID] = struct{}{}
		}
	}
	if ma := baseA.MostRecentAcquire[routineRecv]; ma != nil {
		for lockID := range ma {
			if _, ok := seen[lockID]; !ok {
				seen[lockID] = struct{}{}
			}
		}
	}

	for lockID := range seen {
		addMixedCandidate(routineSend, routineRecv, lockID)
	}
}

// Attempts candidate creation
func addMixedCandidate(sendTid, recvTid, lockID int) {
	// Get latest Acq(x,t)
	acqMapSend := baseA.MostRecentAcquire[sendTid]
	acqMapRecv := baseA.MostRecentAcquire[recvTid]
	if acqMapSend == nil || acqMapRecv == nil {
		return
	}
	acqS, okS := acqMapSend[lockID]
	acqR, okR := acqMapRecv[lockID]
	if !okS || !okR {
		log.Info(fmt.Sprintf("[MD] skip lock %d: missing Acq S=%v R=%v", lockID, okS, okR))
		return
	}

	// Get LockType to skip READ/READ
	modeS := getLockTypeFromAcquireElem(acqS.Elem)
	modeR := getLockTypeFromAcquireElem(acqR.Elem)

	if (modeS == "READ" || modeS == "UNKNOWN") && (modeR == "READ" || modeR == "UNKNOWN") {
		log.Info(fmt.Sprintf("[MD] skip lock %d: READ/READ (%s/%s)", lockID, modeS, modeR))
		return
	}

	// Get HB-relation between the two acquires
	rel := clock.GetHappensBefore(acqS.Vc, acqR.Vc)
	if rel == hb.None {
		log.Info(fmt.Sprintf("[MD] skip lock %d: Acq S,R not related", lockID))
		return
	}

	// Filter unreorderable pairs with WMHB
	if mustOrderByWVC(acqS.Elem, acqR.Elem) || mustOrderByWVC(acqR.Elem, acqS.Elem) {
		log.Info(fmt.Sprintf("[MD] skip lock %d: WVC must-order blocks reorder", lockID))
		return
	}

	// Candidate construction per direction
	buildMixedCandidate(sendTid, recvTid, lockID, acqS, acqR)
}

// Determines sendElem/closeElem/recvElem using MostRecenet channel events
func getChannelPair(sendTid, recvTid int) (*trace.ElementChannel, *trace.ElementChannel, int, bool, bool) {
	recvMap := baseA.MostRecentReceive[recvTid]
	sendMap := baseA.MostRecentSend[sendTid]

	// Case 1: Send-Recv pairing
	for chID, recvVal := range recvMap {
		if sendVal, ok := sendMap[chID]; ok {
			recvElem, _ := recvVal.Elem.(*trace.ElementChannel)
			sendElem, _ := sendVal.Elem.(*trace.ElementChannel)
			if recvElem != nil && sendElem != nil {
				return sendElem, recvElem, chID, recvElem.IsBuffered(), true
			}
		}
	}

	// Case 2: Close-Recv pairing
	for chID, closeElem := range baseA.CloseData {
		if closeElem.GetRoutine() == sendTid {
			var recvElem *trace.ElementChannel = nil
			if rmap := baseA.MostRecentReceive[recvTid]; rmap != nil {
				if rv, ok := rmap[chID]; ok {
					recvElem, _ = rv.Elem.(*trace.ElementChannel)
				}
			}
			return closeElem, recvElem, chID, closeElem.IsBuffered(), true
		}
	}

	return nil, nil, 0, false, false
}

/*
************************************************
// (PRECEEDING) CRITICAL SECTION
************************************************
*/

// hbLeq(a,b): a <= b as (a HB-before b) OR (a concurrent with b)
func hbLeq(a, b *clock.VectorClock) bool {
	if a == nil || b == nil {
		return false
	}
	rel := clock.GetHappensBefore(a, b)
	return rel == hb.Before || rel == hb.Concurrent
}

// hbLt(a,b): strict a < b (a HB-before b only)
func hbLt(a, b *clock.VectorClock) bool {
	if a == nil || b == nil {
		return false
	}
	return clock.GetHappensBefore(a, b) == hb.Before
}

// Helper lastAcquire
func lastAcquire(routine, lockID int) (baseA.ElemWithVc, bool) {
	m := baseA.MostRecentAcquire[routine]
	if m == nil {
		return baseA.ElemWithVc{}, false
	}
	acq, ok := m[lockID]
	return acq, ok && acq.Vc != nil
}

// Helper lastRelaease
func lastRelease(routine, lockID int) (baseA.ElemWithVc, bool) {
	m := baseA.MostRecentRelease[routine]
	if m == nil {
		return baseA.ElemWithVc{}, false
	}
	rel, ok := m[lockID]
	return rel, ok && rel.Vc != nil
}

// Determine if channel event inside CS
func inCSAtEvent(routine, lockID int, eventVc *clock.VectorClock) bool {
	if eventVc == nil {
		return false
	}

	acq, okA := lastAcquire(routine, lockID)
	if !okA {
		return false
	}

	// Acq <= e ?
	if !hbLeq(acq.Vc, eventVc) {
		return false
	}

	// No Release yet -> still in CS
	rel, okR := lastRelease(routine, lockID)
	if !okR {
		return true
	}

	// e < Rel ? (strict; concurrent does NOT count as after)
	return hbLt(eventVc, rel.Vc)
}

// Determine if channel event has PCS
func hasPCSBeforeEvent(routine, lockID int, eventVc *clock.VectorClock) bool {
	if eventVc == nil {
		return false
	}

	acq, okA := lastAcquire(routine, lockID)
	rel, okR := lastRelease(routine, lockID)
	if !okA || !okR {
		return false
	}

	// Acq <= Rel <= e  (<= as before or concurrent)
	return hbLeq(acq.Vc, rel.Vc) && hbLeq(rel.Vc, eventVc)
}

// Helper to classify MD2 type for reporting
func classifyMDCase(cand *mixedCandidate) bool {
	if cand.SendEvent == nil || cand.RecvEvent == nil {
		return false
	}
	sendTid, recvTid, lockID := cand.SendRoutine, cand.RecvRoutine, cand.LockID
	sendVc, recvVc := cand.SendEvent.GetVC(), cand.RecvEvent.GetVC()

	// Without VC no classification
	if sendVc == nil || recvVc == nil {
		return false
	}

	// Determine CS/PCS at time of channel operation (not current lockset)
	sendInCS := inCSAtEvent(sendTid, lockID, sendVc)
	recvInCS := inCSAtEvent(recvTid, lockID, recvVc)
	sendPCS := hasPCSBeforeEvent(sendTid, lockID, sendVc)
	recvPCS := hasPCSBeforeEvent(recvTid, lockID, recvVc)

	// MD2-2: Sender in CS, Receiver PCS
	if sendInCS && recvPCS {
		cand.Case = "MD2-2"
		return true
	}

	// MD2-3: Sender PCS, Receiver in CS
	if sendPCS && recvInCS {
		cand.Case = "MD2-3"
		return true
	}

	// MD2-1: Sender & Receiver in CS (Buffered)
	if sendInCS && recvInCS && cand.Buffered && !sendPCS && !recvPCS {
		cand.Case = "MD2-1"
		return true
	}

	return false
}

// Constructs candidate & triggers reporting
func buildMixedCandidate(sendTid, recvTid, lockID int,
	acqS, acqR baseA.ElemWithVc) {

	sendElem, recvElem, chID, buffered, ok := getChannelPair(sendTid, recvTid)
	if !ok {
		return
	}

	cand := mixedCandidate{
		LockID:      lockID,
		ChanID:      chID,
		SendRoutine: sendTid,
		RecvRoutine: recvTid,
		SendEvent:   sendElem,
		RecvEvent:   recvElem,
		AcqSend:     acqS,
		AcqRecv:     acqR,
		Buffered:    buffered,
	}

	if !classifyMDCase(&cand) {
		return
	}

	mixedCandidates = append(mixedCandidates, cand)
	reportMixedDeadlock(cand)
}

/*
************************************************
// REPORTING
************************************************
*/

// Headline: Kommunikationstyp
func getOpType(md mixedCandidate) string {
	if md.SendEvent != nil && md.SendEvent.GetType(true) == trace.ChannelClose {
		return "Close–Recv"
	}
	return "Send–Recv"
}

// Headline: Buffering
func getBufferType(md mixedCandidate) string {
	if md.Buffered {
		return "Buffered"
	}
	return "Unbuffered"
}

// Channel-Op trace.ObjectType (explanation.objectTypes)
func chanOpCode(e *trace.ElementChannel) trace.ObjectType {
	switch e.GetType(true) {
	case trace.ChannelSend:
		return trace.ObjectType("CS") // Channel: Send
	case trace.ChannelRecv:
		return trace.ObjectType("CR") // Channel: Receive
	case trace.ChannelClose:
		return trace.ObjectType("CC") // Channel: Close
	default:
		return trace.ObjectType("XX")
	}
}

// Message-String
func makeMDMessage(md mixedCandidate) string {
	opType := getOpType(md)
	bufStr := getBufferType(md)
	modeS := getLockTypeFromAcquireElem(md.AcqSend.Elem)
	modeR := getLockTypeFromAcquireElem(md.AcqRecv.Elem)

	return fmt.Sprintf(
		"Potential Mixed Deadlock (%s, %s, %s) on lock %d [%s/%s] with channel %d between routines %d -> %d",
		md.Case, bufStr, opType, md.LockID, modeS, modeR, md.ChanID, md.SendRoutine, md.RecvRoutine,
	)
}

// Stuck-Element (pointing to locking side: recv)
func makeStuckElem(md mixedCandidate) results.TraceElementResult {
	fileR, lineR, tPreR, _ := trace.InfoFromTID(md.AcqRecv.Elem.GetTID())
	return results.TraceElementResult{
		RoutineID: md.RecvRoutine,
		ObjID:     md.LockID,
		ObjType:   trace.ObjectType("DH"),
		File:      fileR,
		Line:      lineR,
		TPre:      tPreR, // int
	}
}

// Context elements (actual channel events)
func makeContextElems(md mixedCandidate) []results.ResultElem {
	// Sender/Closer
	var sendFile string
	var sendLine int
	var sendTpre int
	sendCode := trace.ObjectType("XX")
	if md.SendEvent != nil {
		sendCode = chanOpCode(md.SendEvent)
		f, l, tp, _ := trace.InfoFromTID(md.SendEvent.GetTID()) // CS / CC
		sendFile, sendLine, sendTpre = f, l, tp
	}

	// Receiver
	var recvFile string
	var recvLine int
	var recvTpre int
	recvCode := trace.ObjectType("CR")
	if md.RecvEvent != nil {
		recvCode = chanOpCode(md.RecvEvent) // CR
		f, l, tp, _ := trace.InfoFromTID(md.RecvEvent.GetTID())
		recvFile, recvLine, recvTpre = f, l, tp
	}

	return []results.ResultElem{
		results.TraceElementResult{
			RoutineID: md.SendRoutine,
			ObjID:     md.ChanID,
			ObjType:   sendCode, // CS / CC
			File:      sendFile,
			Line:      sendLine,
			TPre:      sendTpre,
		},
		results.TraceElementResult{
			RoutineID: md.RecvRoutine,
			ObjID:     md.ChanID,
			ObjType:   recvCode, // CR
			File:      recvFile,
			Line:      recvLine,
			TPre:      recvTpre,
		},
	}
}

// Final Report
func reportMixedDeadlock(md mixedCandidate) {
	msg := makeMDMessage(md)
	stuck := makeStuckElem(md)
	context := makeContextElems(md)

	results.Result(
		results.CRITICAL,
		helper.PMixedDeadlock,
		msg,
		[]results.ResultElem{stuck},
		"context",
		context,
	)

	log.Info(msg)
}

// Determining "impossible" reorderings if two events have "Must-Happens-Before" relation.
// (WMHB = HB-underapproximation)
func mustOrderByWVC(e1, e2 trace.Element) bool {
	if e1 == nil || e2 == nil {
		return false
	}
	v1 := e1.GetWVC()
	v2 := e2.GetWVC()
	if v1 == nil || v2 == nil {
		return false
	}
	rel := clock.GetHappensBefore(v1, v2)
	return rel == hb.Before || rel == hb.After
}
