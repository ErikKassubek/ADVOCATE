// advocate/analysis/analysis/scenarios/mixedDeadlock.go

// Copyright (c) 2024 Erik Kassubek
//
// File: mixedDeadlock.go
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

// ---------------------------------------------------------------------------
// Data structures
// ---------------------------------------------------------------------------

// mdLockRef associates a lock ID and its CS/PCS status with the concrete RD
type mdLockRef struct {
	LockID baseA.LockID
	IsCS   bool
	RD     *mdRDNode
}

// mdRDNode represents a single lock-acquire event in a goroutine
type mdRDNode struct {
	Thread   baseA.ThreadID
	Lock     baseA.LockID
	Requests []baseA.LockEvent
	Elem     *trace.ElementMutex
}

// mdCDNode represents a channel operation that has at least one lock context
type mdCDNode struct {
	Thread   baseA.ThreadID
	ChanID   int
	OpType   trace.OperationType // ChannelSend | ChannelRecv | ChannelClose
	Buffered bool
	Event    baseA.LockEvent       // channel-op VC
	AssocRDs []mdLockRef           // CS and PCS lock contexts at time of op
	Elem     *trace.ElementChannel // concrete trace element
}

// mdThreadState holds per-goroutine online recording state
type mdThreadState struct {
	CurrentLockset baseA.Lockset
	ActiveRDs      map[baseA.LockID]*mdRDNode // open RD per currently-held lock
	MostRecentRD   map[baseA.LockID]*mdRDNode // last completed RD per lock (PCS)
}

// mdState as global analysis state for mixed-deadlock detection
type mdState struct {
	Threads map[baseA.ThreadID]*mdThreadState
	AllCDs  []*mdCDNode
}

var currentMDState mdState

// ---------------------------------------------------------------------------
// Phase 1: Online event recording
// ---------------------------------------------------------------------------

// ResetMixedDeadlockState resets all state before a new analysis run.
func ResetMixedDeadlockState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)
	currentMDState = mdState{
		Threads: make(map[baseA.ThreadID]*mdThreadState),
	}
}

// getOrCreateMDThread returns the per-goroutine state, creating it if absent.
func getOrCreateMDThread(tid baseA.ThreadID) *mdThreadState {
	if t, ok := currentMDState.Threads[tid]; ok {
		return t
	}
	t := &mdThreadState{
		CurrentLockset: make(baseA.Lockset),
		ActiveRDs:      make(map[baseA.LockID]*mdRDNode),
		MostRecentRD:   make(map[baseA.LockID]*mdRDNode),
	}
	currentMDState.Threads[tid] = t
	return t
}

// HandleMutexEventForMixedDeadlock processes one mutex trace event
//
// On acquire: create RD node and add lock to CurrentLockset
// On release: move RD to MostRecentRD (PCS4) and remove from lockset
//
// Cross-goroutine unlocks (Go semaphore semantics) are silently ignored when
// the lock is not in the goroutine's own ActiveRDs map (like resourceDeadlock)
func HandleMutexEventForMixedDeadlock(element *trace.ElementMutex) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	tid := baseA.ThreadID(element.GetRoutine())
	t := getOrCreateMDThread(tid)

	readLock := false
	switch element.GetType(true) {
	case trace.MutexRLock, trace.MutexTryRLock:
		readLock = true
	}
	lockID := baseA.LockID{ID: element.GetObjId(), ReadLock: readLock}

	event := baseA.LockEvent{
		ThreadID:    tid,
		TraceID:     element.GetTID(),
		LockID:      element.GetObjId(),
		VectorClock: element.GetWVC().Copy(),
	}

	switch element.GetType(true) {
	case trace.MutexLock, trace.MutexTryLock, trace.MutexRLock, trace.MutexTryRLock:
		mdInsertRD(t, tid, lockID, event, element)
		t.CurrentLockset.Add(lockID)
		log.Debug(fmt.Sprintf("MD phase1: T%d acq(lock=%d) LS=%v -> RD recorded",
			tid, element.GetObjId(), t.CurrentLockset))

	case trace.MutexUnlock, trace.MutexRUnlock:
		// Determine if this is unlocking a read lock or write lock
		isReadUnlock := false
		switch element.GetType(true) {
		case trace.MutexRUnlock:
			isReadUnlock = true
		}
		lockID := baseA.LockID{ID: element.GetObjId(), ReadLock: isReadUnlock}

		if rd, ok := t.ActiveRDs[lockID]; ok {
			t.MostRecentRD[lockID] = rd
			delete(t.ActiveRDs, lockID)
			log.Debug(fmt.Sprintf("MD phase1: T%d rel(lock=%d, read=%v) -> moved to MostRecentRD",
				tid, element.GetObjId(), isReadUnlock))
		}
		t.CurrentLockset.Remove(lockID)
	}
}

// mdInsertRD creates RD node for the given lock-acquire event
//
// Always replaces ActiveRDs[lockID] so that loop re-acquires produce a new
// node pointing to the most recent *trace.ElementMutex
// Old nodes remain alive as long as any CDNode captured them in its AssocRDs
func mdInsertRD(
	t *mdThreadState,
	tid baseA.ThreadID,
	lockID baseA.LockID,
	event baseA.LockEvent,
	element *trace.ElementMutex,
) {
	rd := &mdRDNode{
		Thread:   tid,
		Lock:     lockID,
		Requests: []baseA.LockEvent{event.Clone()},
		Elem:     element,
	}
	t.ActiveRDs[lockID] = rd
}

// HandleChannelEventForMixedDeadlock processes one channel trace event
//
// Collects all CS and PCS lock contexts active at the time of the channel op
// and records a CDNode if any lock context is present
// Operations with no lock context cannot contribute to a mixed deadlock
func HandleChannelEventForMixedDeadlock(element *trace.ElementChannel) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	opType := element.GetType(true)
	switch opType {
	case trace.ChannelSend, trace.ChannelRecv, trace.ChannelClose:
		// proceed
	default:
		return
	}

	tid := baseA.ThreadID(element.GetRoutine())
	t := getOrCreateMDThread(tid)

	event := baseA.LockEvent{
		ThreadID:    tid,
		TraceID:     element.GetTID(),
		LockID:      element.GetObjId(),
		VectorClock: element.GetWVC().Copy(),
	}

	var assocRDs []mdLockRef

	// CS locks: currently held during the channel op
	for lockID := range t.CurrentLockset {
		if rd, ok := t.ActiveRDs[lockID]; ok {
			assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: true, RD: rd})
		}
	}

	// PCS locks: released before the channel op (most recent completed CS)
	// Skip any lock that is currently held, those are already CS above
	for lockID, rd := range t.MostRecentRD {
		if _, held := t.CurrentLockset[lockID]; held {
			continue
		}
		assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: false, RD: rd})
	}

	if len(assocRDs) == 0 {
		log.Debug(fmt.Sprintf("MD phase1: T%d chan(op=%s, ch=%d) — no lock context, skipping",
			tid, opType, element.GetObjId()))
		return
	}

	cd := &mdCDNode{
		Thread:   tid,
		ChanID:   element.GetObjId(),
		OpType:   opType,
		Buffered: element.IsBuffered(),
		Event:    event,
		AssocRDs: assocRDs,
		Elem:     element,
	}
	currentMDState.AllCDs = append(currentMDState.AllCDs, cd)

	csCount, pcsCount := 0, 0
	for _, r := range assocRDs {
		if r.IsCS {
			csCount++
		} else {
			pcsCount++
		}
	}
	log.Debug(fmt.Sprintf(
		"MD phase1: T%d chan(op=%s, ch=%d, buf=%v) — CD recorded assocRDs=%d CS=%d PCS=%d",
		tid, opType, element.GetObjId(), element.IsBuffered(),
		len(assocRDs), csCount, pcsCount))
}

// ---------------------------------------------------------------------------
// Phase 2: Offline CD-CD partner matching
// ---------------------------------------------------------------------------

// CheckForMixedDeadlock as entry point
//
// Iterates over AllCDs looking for Recv events (and unbuffered Send events)
// that have a recorded communication partner
//
// # For each pair it checks MD conditions and reports if they pass
//
// Buffered-Send events are skipped here, because their partner Recv drives the
// check, which avoids processing the same pair twice
//
// Close events are matched from the receiver side via mdFindCloseCDNode
func CheckForMixedDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	log.Debug(fmt.Sprintf("MD phase2: start partner matching, AllCDs=%d",
		len(currentMDState.AllCDs)))

	// Build per-channel index for efficient partner lookup
	cdByChan := make(map[int][]*mdCDNode, len(currentMDState.AllCDs))
	for _, cd := range currentMDState.AllCDs {
		cdByChan[cd.ChanID] = append(cdByChan[cd.ChanID], cd)
	}

	// Deduplication: each (elemA, elemB) pair is reported at most once
	// Key as sorted pointer pair so (A,B) and (B,A) map to the same entry
	reported := make(map[[2]*trace.ElementChannel]bool)

	for _, cd := range currentMDState.AllCDs {
		switch cd.OpType {

		case trace.ChannelRecv:
			// Check for send partner first
			partner := mdFindPartnerForRecv(cd, cdByChan)
			if partner != nil {
				mdCheckAndReport(cd, partner, reported)
			}

			// Also check for close partner (even if GetClosed() is false)
			closePartner := mdFindCloseCDNode(cd, cdByChan)
			if closePartner != nil {
				mdCheckAndReport(cd, closePartner, reported)
			}

		case trace.ChannelSend:
			// Only process unbuffered sends; buffered sends are handled by
			// their matching Recv event to avoid duplicate work
			if cd.Buffered {
				continue
			}
			partner := mdFindPartnerCDByElem(cd.Elem.GetPartner(), cdByChan[cd.ChanID])
			if partner == nil {
				continue
			}
			mdCheckAndReport(cd, partner, reported)

		default:
			// ChannelClose: handled from the receiver side
		}
	}
}

// ---------------------------------------------------------------------------
// Partner lookup helpers
// ---------------------------------------------------------------------------

// mdFindPartnerForRecv finds the CDNode that is the communication partner of
// the given Recv CDNode
//
//	Recv on closed channel to find a Close CDNode on the same channel
//	Normal recv uses GetPartner() on the underlying element
func mdFindPartnerForRecv(recvCD *mdCDNode, cdByChan map[int][]*mdCDNode) *mdCDNode {
	elem := recvCD.Elem

	if elem.GetClosed() {
		// Receive-on-close: partner is the Close CDNode
		return mdFindCloseCDNode(recvCD, cdByChan)
	}

	// Normal send/recv: GetPartner() is set by the HB computation pass
	return mdFindPartnerCDByElem(elem.GetPartner(), cdByChan[recvCD.ChanID])
}

// mdFindPartnerCDByElem returns CDNode with underlying element pointer
// equals partnerElem, searching within candidates
// Returns nil if partnerElem is nil or not found
func mdFindPartnerCDByElem(partnerElem *trace.ElementChannel, candidates []*mdCDNode) *mdCDNode {
	if partnerElem == nil {
		return nil
	}
	for _, c := range candidates {
		if c.Elem == partnerElem {
			return c
		}
	}
	return nil
}

// mdFindCloseCDNode finds the Close CDNode on the same channel as recvCD
//
// Prefers a Close that is concurrent with the Recv (strongest WMHB evidence)
// Falls back to any Close that happened-before the Recv
// Ignores Close events in the same goroutine as the Recv
// mdFindCloseCDNode should find the close CDNode regardless of order
func mdFindCloseCDNode(recvCD *mdCDNode, cdByChan map[int][]*mdCDNode) *mdCDNode {
	var best *mdCDNode

	for _, c := range cdByChan[recvCD.ChanID] {
		if c.OpType != trace.ChannelClose {
			continue
		}
		if c.Thread == recvCD.Thread {
			continue
		}

		// Accept any close on the same channel (order doesn't matter for detection)
		// We'll use the concurrent check to filter infeasible reorderings
		if best == nil {
			best = c
		}
	}
	return best
}

// ---------------------------------------------------------------------------
// MD condition checks and reporting
// ---------------------------------------------------------------------------

// mdCheckAndReport applies all MD conditions to the pair (cdA, cdB) and
// reports the bug if they all pass
//
// cdA and cdB must be on the same channel with complementary op types
// The reported map prevents the same pair from being reported more than once
func mdCheckAndReport(cdA, cdB *mdCDNode, reported map[[2]*trace.ElementChannel]bool) {
	key := mdPairKey(cdA.Elem, cdB.Elem)
	if reported[key] {
		return
	}

	// MD2: find a shared lock where at least one side is CS
	for _, refA := range cdA.AssocRDs {
		for _, refB := range cdB.AssocRDs {

			// Locks must be the same and at least one must be a write lock.
			if !refA.LockID.EqualsCouldBlock(refB.LockID) {
				continue
			}

			// Both PCS: no goroutine holds the lock during the channel op,
			// so neither can block the other on lock acquisition.
			if !refA.IsCS && !refB.IsCS {
				log.Debug(fmt.Sprintf(
					"MD phase2: T%d/T%d ch=%d lock=%d — both PCS, skip",
					cdA.Thread, cdB.Thread, cdA.ChanID, refA.LockID.ID))
				continue
			}

			// Both CS: MD2-1 Buffered
			// Unbuffered MD2-1 deadlocks deterministically
			if refA.IsCS && refB.IsCS && !cdA.Buffered {
				// Only skip for send/recv pairs, not for close/recv
				if cdA.OpType != trace.ChannelClose && cdB.OpType != trace.ChannelClose {
					log.Debug("MD phase2: both CS on unbuffered send/recv - skip")
					continue
				}
			}

			// WMHB feasibility: lock-acquire VCs must be concurrent.
			// If one HB-precedes the other, swapping their order would
			// violate the causal structure (e.g. goroutine forked after).
			if !mdLockAcqAreConcurrent(refA.RD, refB.RD) {
				log.Debug(fmt.Sprintf(
					"MD phase2: T%d/T%d ch=%d lock=%d — lock acquires not concurrent, skip",
					cdA.Thread, cdB.Thread, cdA.ChanID, refA.LockID.ID))
				continue
			}

			// When all conditions passed, determine roles and report
			holderCD, holderRef, waiterCD, waiterRef :=
				mdDetermineRoles(cdA, refA, cdB, refB)

			log.Debug(fmt.Sprintf(
				"MD phase2: FOUND MD | ch=%d lock=%d | holder=T%d(CS=%v) waiter=T%d(CS=%v)",
				cdA.ChanID, refA.LockID.ID,
				holderCD.Thread, holderRef.IsCS,
				waiterCD.Thread, waiterRef.IsCS))

			mdReportCandidate(holderCD, holderRef, waiterCD, waiterRef)
			reported[key] = true
			return
		}
	}
}

// mdLockAcqAreConcurrent returns true when any pair of lock-acquire VCs from
// the two RD nodes is concurrent under the weak HB relation
func mdLockAcqAreConcurrent(rdA, rdB *mdRDNode) bool {
	if rdA == nil || rdB == nil {
		return false
	}
	for _, reqA := range rdA.Requests {
		if reqA.VectorClock == nil {
			continue
		}
		for _, reqB := range rdB.Requests {
			if reqB.VectorClock == nil {
				continue
			}
			if clock.GetHappensBefore(reqA.VectorClock, reqB.VectorClock) == hb.Concurrent {
				return true
			}
		}
	}
	return false
}

// mdDetermineRoles assigns holder and waiter roles to the two goroutines
//
// Holder: goroutine that will hold the lock while blocking on its channel op
// Waiter: goroutine whose lock acquire will block (because the holder holds it)
//
// Role assignment:
//
//	Asymmetric (MD2-2 or MD2-3):
//	  - CS side is always holder
//	  - PCS always waiter,  goroutine must re-acquire the lock to make progress
//
//	Symmetric (MD2-1, both CS):
//	  - Goroutine whose lock acquire comes LATER in the working trace becomes holder
//	  - In rewrite causing holder to block on its channel op while the waiter blocks trying to acquire
func mdDetermineRoles(
	cdA *mdCDNode, refA mdLockRef,
	cdB *mdCDNode, refB mdLockRef,
) (holderCD *mdCDNode, holderRef mdLockRef, waiterCD *mdCDNode, waiterRef mdLockRef) {

	// Asymmetric: CS side is the holder
	if refA.IsCS && !refB.IsCS {
		return cdA, refA, cdB, refB
	}
	if refB.IsCS && !refA.IsCS {
		return cdB, refB, cdA, refA
	}

	// Both CS (MD2-1)
	if len(refA.RD.Requests) > 0 && len(refB.RD.Requests) > 0 {
		vcA := refA.RD.Requests[0].VectorClock
		vcB := refB.RD.Requests[0].VectorClock
		if vcA != nil && vcB != nil {
			switch clock.GetHappensBefore(vcA, vcB) {
			case hb.Before:
				// A acquired first, B acquired later, B is the holder
				return cdB, refB, cdA, refA
			case hb.After:
				// B acquired first, A acquired later, A is the holder
				return cdA, refA, cdB, refB
			}
			// Concurrent: fall through to tPre tie-break
		}
	}

	// Tie-break: larger tPre on the lock-acquire element = acquired later = holder
	tPreA, tPreB := 0, 0
	if refA.RD.Elem != nil {
		tPreA = refA.RD.Elem.GetTPre()
	}
	if refB.RD.Elem != nil {
		tPreB = refB.RD.Elem.GetTPre()
	}
	if tPreA >= tPreB {
		return cdA, refA, cdB, refB
	}
	return cdB, refB, cdA, refA
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

// mdReportCandidate reports a validated MD2 pair as a PMixedDeadlock bug
//
// TraceElement2 layout (4 elements, consumed by rewriteMixedDeadlock):
//
//	[0]  cdHolder.Elem    *trace.ElementChannel   holder's channel op
//	[1]  lockHolder.Elem  *trace.ElementMutex     holder's lock acquire
//	[2]  cdWaiter.Elem    *trace.ElementChannel   waiter's channel op
//	[3]  lockWaiter.Elem  *trace.ElementMutex     waiter's lock acquire
//
// TraceElement1 (stuck element): the waiter's lock acquire, mirroring how
// resourceDeadlock.go uses the last cycle element with ObjType="DH"
func mdReportCandidate(
	holderCD *mdCDNode, holderRef mdLockRef,
	waiterCD *mdCDNode, waiterRef mdLockRef,
) {
	if holderCD.Elem == nil || holderRef.RD == nil || holderRef.RD.Elem == nil ||
		waiterCD.Elem == nil || waiterRef.RD == nil || waiterRef.RD.Elem == nil {
		log.Error("MD report: nil element pointer in candidate — skipping")
		return
	}

	// --- holder's channel element ---
	holderChanFile, holderChanLine, holderChanTPre, err :=
		trace.InfoFromTID(holderCD.Event.TraceID)
	if err != nil {
		log.Error("MD report: InfoFromTID for holder CD: ", err.Error())
		return
	}
	holderChanRes := results.TraceElementResult{
		RoutineID: int(holderCD.Thread),
		ObjID:     holderCD.ChanID,
		TPre:      holderChanTPre,
		ObjType:   holderCD.OpType,
		File:      holderChanFile,
		Line:      holderChanLine,
	}

	// --- holder's lock acquire element ---
	holderLockReq := holderRef.RD.Requests[0]
	holderLockFile, holderLockLine, holderLockTPre, err :=
		trace.InfoFromTID(holderLockReq.TraceID)
	if err != nil {
		log.Error("MD report: InfoFromTID for holder RD: ", err.Error())
		return
	}
	holderLockRes := results.TraceElementResult{
		RoutineID: int(holderRef.RD.Thread),
		ObjID:     holderRef.LockID.ID,
		TPre:      holderLockTPre,
		ObjType:   "DC",
		File:      holderLockFile,
		Line:      holderLockLine,
	}

	// --- waiter's channel element ---
	waiterChanFile, waiterChanLine, waiterChanTPre, err :=
		trace.InfoFromTID(waiterCD.Event.TraceID)
	if err != nil {
		log.Error("MD report: InfoFromTID for waiter CD: ", err.Error())
		return
	}
	waiterChanRes := results.TraceElementResult{
		RoutineID: int(waiterCD.Thread),
		ObjID:     waiterCD.ChanID,
		TPre:      waiterChanTPre,
		ObjType:   waiterCD.OpType,
		File:      waiterChanFile,
		Line:      waiterChanLine,
	}

	// --- waiter's lock acquire element (also stuck element) ---
	waiterLockReq := waiterRef.RD.Requests[0]
	waiterLockFile, waiterLockLine, waiterLockTPre, err :=
		trace.InfoFromTID(waiterLockReq.TraceID)
	if err != nil {
		log.Error("MD report: InfoFromTID for waiter RD: ", err.Error())
		return
	}
	waiterLockRes := results.TraceElementResult{
		RoutineID: int(waiterRef.RD.Thread),
		ObjID:     waiterRef.LockID.ID,
		TPre:      waiterLockTPre,
		ObjType:   "DC",
		File:      waiterLockFile,
		Line:      waiterLockLine,
	}

	// Stuck element is waiter's lock acquire and never completes
	stuckElement := waiterLockRes
	stuckElement.ObjType = "DH"

	cycleElements := []results.ResultElem{
		holderChanRes, // [0] holder's channel op
		holderLockRes, // [1] holder's lock acquire
		waiterChanRes, // [2] waiter's channel op
		waiterLockRes, // [3] waiter's lock acquire
	}

	log.Debug(fmt.Sprintf(
		"MD report: PMixedDeadlock holder=T%d(ch=%d CS=%v) waiter=T%d(ch=%d CS=%v) lock=%d",
		holderCD.Thread, holderCD.ChanID, holderRef.IsCS,
		waiterCD.Thread, waiterCD.ChanID, waiterRef.IsCS,
		holderRef.LockID.ID))

	results.Result(
		results.CRITICAL,
		helper.PMixedDeadlock,
		"stuck", []results.ResultElem{stuckElement},
		"cycle", cycleElements,
	)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mdPairKey returns a canonical key for a channel-element pair so that
// (A, B) and (B, A) map to the same entry in the reported map
// The element with the smaller tPre is placed first
func mdPairKey(a, b *trace.ElementChannel) [2]*trace.ElementChannel {
	if a.GetTPre() <= b.GetTPre() {
		return [2]*trace.ElementChannel{a, b}
	}
	return [2]*trace.ElementChannel{b, a}
}
