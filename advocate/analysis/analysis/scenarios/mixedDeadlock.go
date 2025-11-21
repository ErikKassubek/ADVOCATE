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
	"fmt"
)

/*
------------------------------------------------
// Two-Cycle Mixed Deadlock (MDS-2) model
------------------------------------------------
// MDS-2 cases:
// - MD2-1: both sender and receiver inside CS:  x ∈ LS(send) ∩ LS(recv)
// - MD2-2: sender inside CS, receiver with PCS: x ∈ LS(send), x ∉ LS(recv)
// - MD2-3: sender with PCS, receiver inside CS: x ∈ LS(recv), x ∉ LS(send)
*/

// ------------------------------------------------
// Constants & Types
// ------------------------------------------------

const (
	LockTypeRead    = "READ"
	LockTypeWrite   = "WRITE"
	LockTypeUnknown = "UNKNOWN"
)

type mixedCandidate struct {
	LockID      int
	ChanID      int
	SendRoutine int
	RecvRoutine int
	SendEvent   *trace.ElementChannel
	RecvEvent   *trace.ElementChannel
	AcqSend     baseA.ElemWithVc
	AcqRecv     baseA.ElemWithVc
	Case        string // MD2-1 | MD2-2 | MD2-3
	Buffered    bool
}

// ------------------------------------------------
// Global State & Reset
// ------------------------------------------------

var mixedCandidates []mixedCandidate

func ResetMixedDeadlockState() {
	mixedCandidates = make([]mixedCandidate, 0)
}

// ------------------------------------------------
// Entry Point (from channel.go)
// ------------------------------------------------

// Analyzes MD-scenario using event snapshots
func CheckForMixedDeadlock(sendElem trace.Element, recvElem trace.Element) {
	sCh, _ := sendElem.(*trace.ElementChannel)
	rCh, _ := recvElem.(*trace.ElementChannel)
	if sCh == nil || rCh == nil {
		return
	}

	sSnap, sOK := baseA.CSSnaps[eventKeyFor(sCh)]
	rSnap, rOK := baseA.CSSnaps[eventKeyFor(rCh)]
	if !sOK || !rOK {
		return
	}

	addMixedCandidate(sCh, rCh, sSnap, rSnap)
}

// ------------------------------------------------
// Create Candidate
// ------------------------------------------------

// Iterates over union and filters for possible candidates
func addMixedCandidate(
	sCh, rCh *trace.ElementChannel,
	sSnap, rSnap baseA.CSSnapshot,
) {
	sendTid := sCh.GetRoutine()
	recvTid := rCh.GetRoutine()
	chID := sCh.GetID()
	buffered := sCh.IsBuffered() || rCh.IsBuffered()

	seen := collectLockUnion(sSnap.ByLock, rSnap.ByLock)
	for lockID := range seen {
		sFlag, hasS := sSnap.ByLock[lockID]
		rFlag, hasR := rSnap.ByLock[lockID]
		if !hasS && !hasR {
			continue
		}
		if shouldSkipByMode(hasS, sFlag, hasR, rFlag) {
			continue
		}
		if !passesWMHB(hasS, sFlag, hasR, rFlag) {
			continue
		}

		buildMixedCandidate(
			lockID, chID, buffered,
			sendTid, recvTid,
			sCh, rCh,
			sFlag, rFlag,
		)
	}
}

// Get union of locks
func collectLockUnion(a, b map[int]baseA.CSFlag) map[int]struct{} {
	seen := make(map[int]struct{}, len(a)+len(b))
	for lid := range a {
		seen[lid] = struct{}{}
	}
	for lid := range b {
		seen[lid] = struct{}{}
	}
	return seen
}

// Get LockType Read/Write
func getLockType(e trace.Element) string {
	if e == nil {
		return LockTypeUnknown
	}
	if mu, ok := e.(*trace.ElementMutex); ok {
		switch mu.GetType(true) {
		case trace.MutexRLock, trace.MutexTryRLock:
			return LockTypeRead
		default:
			return LockTypeWrite
		}
	}
	return LockTypeUnknown
}

// Skip Read/Read
func shouldSkipByMode(hasS bool, s baseA.CSFlag, hasR bool, r baseA.CSFlag) bool {
	modeS := LockTypeUnknown
	if hasS {
		modeS = getLockType(s.Acq.Elem)
	}
	modeR := LockTypeUnknown
	if hasR {
		modeR = getLockType(r.Acq.Elem)
	}
	return (modeS == LockTypeRead || modeS == LockTypeUnknown) &&
		(modeR == LockTypeRead || modeR == LockTypeUnknown)
}

// HB-underapproximation (WMHB): Two events Must-HB
func mustOrderByWVC(e1, e2 trace.Element) bool {
	if e1 == nil || e2 == nil {
		return false
	}
	v1, v2 := e1.GetWVC(), e2.GetWVC()
	if v1 == nil || v2 == nil {
		return false
	}
	return !clock.IsConcurrent(v1, v2)
}

// Filter impossible reorderings (WMHB)
func passesWMHB(hasS bool, s baseA.CSFlag, hasR bool, r baseA.CSFlag) bool {
	if !hasS || !hasR || s.Acq.Elem == nil || r.Acq.Elem == nil {
		return true
	}
	if mustOrderByWVC(s.Acq.Elem, r.Acq.Elem) ||
		mustOrderByWVC(r.Acq.Elem, s.Acq.Elem) {
		return false
	}
	return true
}

// Build MD Candidate
func buildMixedCandidate(
	lockID, chID int,
	buffered bool,
	sendTid, recvTid int,
	sCh, rCh *trace.ElementChannel,
	sFlag, rFlag baseA.CSFlag,
) {
	cand := mixedCandidate{
		LockID:      lockID,
		ChanID:      chID,
		SendRoutine: sendTid,
		RecvRoutine: recvTid,
		SendEvent:   sCh,
		RecvEvent:   rCh,
		AcqSend:     sFlag.Acq,
		AcqRecv:     rFlag.Acq,
		Buffered:    buffered,
	}
	if classifyMDCase(&cand) {
		mixedCandidates = append(mixedCandidates, cand)
		reportMixedDeadlock(cand)
	}
}

// ------------------------------------------------
// Get CS/PCS
// ------------------------------------------------

// Get EventKey
func eventKeyFor(ch *trace.ElementChannel) baseA.EventKey {
	return baseA.EventKey{
		Routine: ch.GetRoutine(),
		ChanID:  ch.GetID(),
		OID:     ch.GetOID(),
	}
}

// Determine if CS with existing snapshots
func inCSAt(ch *trace.ElementChannel, lockID int) bool {
	if ch == nil {
		return false
	}
	snap, ok := baseA.CSSnaps[eventKeyFor(ch)]
	if !ok {
		return false
	}
	if f, ok := snap.ByLock[lockID]; ok {
		return f.InCS
	}
	return false
}

// Determine if PCS with existing snapshots
func hasPCSBefore(ch *trace.ElementChannel, lockID int) bool {
	if ch == nil {
		return false
	}
	snap, ok := baseA.CSSnaps[eventKeyFor(ch)]
	if !ok {
		return false
	}
	if f, ok := snap.ByLock[lockID]; ok {
		return f.PCS
	}
	return false
}

// ------------------------------------------------
// CS/PCS snapshot & helpers
// ------------------------------------------------

// HB helpers
func hbLE(a, b *clock.VectorClock) bool { // a <=HB b
	if a == nil || b == nil {
		return false
	}
	rel := clock.GetHappensBefore(a, b)
	return rel == hb.Before || rel == hb.Concurrent
}
func hbLT(a, b *clock.VectorClock) bool { // a <HB b
	if a == nil || b == nil {
		return false
	}
	return clock.GetHappensBefore(a, b) == hb.Before
}
func decideCSFlags(acq baseA.ElemWithVc, rel baseA.ElemWithVc, hasRel bool,
	eventVc *clock.VectorClock, lockMode string, rCount int) (inCS, pcs bool) {

	if lockMode == LockTypeRead && rCount > 0 {
		return true, false
	}
	inCS, pcs = true, false // no matching rel -> still in CS & no PCS

	if hasRel {
		inCS = hbLT(eventVc, rel.Vc)                        // Event <HB Rel ?
		pcs = hbLE(acq.Vc, rel.Vc) && hbLE(rel.Vc, eventVc) // Acq <= Rel <= Event
	}
	return
}

// Finds Acq-Rel pair for correct CS (else nil)
func matchingReleaseFor(acq baseA.ElemWithVc, relMap map[int]baseA.ElemWithVc, lockID int) (baseA.ElemWithVc, bool) {
	if relMap == nil {
		return baseA.ElemWithVc{}, false
	}
	rel, ok := relMap[lockID]
	if !ok || rel.Vc == nil {
		return baseA.ElemWithVc{}, false
	}
	if hbLE(acq.Vc, rel.Vc) { // same CS
		return rel, true
	}
	return baseA.ElemWithVc{}, false // release from earlier CS
}

// Determines PCS and CS using Acq, Rel, Event-VC
func DecideCSForEvent(routine int, eventVc *clock.VectorClock) map[int]baseA.CSFlag {
	flags := make(map[int]baseA.CSFlag)
	if eventVc == nil {
		return flags
	}

	acqMap := baseA.MostRecentAcquire[routine]
	if acqMap == nil {
		return flags
	}
	relMap := baseA.MostRecentRelease[routine]

	for lockID, acq := range acqMap {
		if acq.Vc == nil {
			continue
		}
		if !hbLE(acq.Vc, eventVc) {
			continue
		} // Acq ≤ Event ?

		rel, hasRel := matchingReleaseFor(acq, relMap, lockID)

		lockMode := getLockType(acq.Elem)
		rCount := 0
		if rcMap, ok := baseA.RLockCount[routine]; ok {
			rCount = rcMap[lockID]
		}

		inCS, pcs := decideCSFlags(acq, rel, hasRel, eventVc, lockMode, rCount)
		if inCS || pcs {
			flags[lockID] = baseA.CSFlag{
				InCS: inCS,
				PCS:  pcs,
				Acq:  acq,
				Rel:  rel,
			}
		}
	}
	return flags
}

// ------------------------------------------------
// Classification
// ------------------------------------------------

// Helper to classify MD2 type for reporting
func classifyMDCase(cand *mixedCandidate) bool {
	if cand.SendEvent == nil || cand.RecvEvent == nil {
		return false
	}

	// Determine CS/PCS at time of channel operation (not current lockset)
	sendInCS := inCSAt(cand.SendEvent, cand.LockID)
	recvInCS := inCSAt(cand.RecvEvent, cand.LockID)
	sendPCS := hasPCSBefore(cand.SendEvent, cand.LockID)
	recvPCS := hasPCSBefore(cand.RecvEvent, cand.LockID)

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

// ------------------------------------------------
// Reporting
// ------------------------------------------------

// Communication Type
func getOpType(md mixedCandidate) string {
	if md.SendEvent != nil && md.SendEvent.GetType(true) == trace.ChannelClose {
		return "Close–Recv"
	}
	return "Send–Recv"
}

// Channel Type
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

// Message
func makeMDMessage(md mixedCandidate) string {
	opType := getOpType(md)
	bufStr := getBufferType(md)
	modeS := getLockType(md.AcqSend.Elem)
	modeR := getLockType(md.AcqRecv.Elem)

	return fmt.Sprintf(
		"Potential Mixed Deadlock (%s, %s, %s) on lock %d [%s/%s] with channel %d between routines %d -> %d",
		md.Case, bufStr, opType, md.LockID, modeS, modeR, md.ChanID, md.SendRoutine, md.RecvRoutine,
	)
}

// Stuck-Element
func makeStuckElem(md mixedCandidate) results.TraceElementResult {
	fileR, lineR, tPreR, _ := trace.InfoFromTID(md.AcqRecv.Elem.GetTID())
	return results.TraceElementResult{
		RoutineID: md.RecvRoutine,
		ObjID:     md.LockID,
		ObjType:   trace.ObjectType("DH"),
		File:      fileR,
		Line:      lineR,
		TPre:      tPreR,
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
