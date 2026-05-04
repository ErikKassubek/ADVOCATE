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
	Thread     baseA.ThreadID
	ChanID     int
	OpType     trace.OperationType // ChannelSend | ChannelRecv | ChannelClose
	Buffered   bool
	Event      baseA.LockEvent       // channel-op VC
	AssocRDs   []mdLockRef           // CS and PCS lock contexts at time of op
	Elem       *trace.ElementChannel // concrete trace element
	Depth      int                   // total lock stack depth at channel op
	ReadDepth  int                   // read lock depth at channel op (for RWMutex)
	WriteDepth int                   // write lock depth at channel op
}

// mdThreadState holds per-goroutine online recording state
type mdThreadState struct {
	CurrentLockset baseA.Lockset
	ActiveRDs      map[baseA.LockID][]*mdRDNode // Stack for nested locks
	MostRecentRD   map[baseA.LockID]*mdRDNode
	ReadLockCount  map[baseA.LockID]int
	LockDepth      int // Total locks currently held
	ReadDepth      int // Read locks currently held (for RWMutex)
	WriteDepth     int // Write locks currently held
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
		ActiveRDs:      make(map[baseA.LockID][]*mdRDNode),
		MostRecentRD:   make(map[baseA.LockID]*mdRDNode),
		ReadLockCount:  make(map[baseA.LockID]int),
		LockDepth:      0,
		ReadDepth:      0,
		WriteDepth:     0,
	}
	currentMDState.Threads[tid] = t
	return t
}

// HandleMutexEventForMixedDeadlock processes one mutex trace event
func HandleMutexEventForMixedDeadlock(element *trace.ElementMutex) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	tid := baseA.ThreadID(element.GetRoutine())
	t := getOrCreateMDThread(tid)

	isReadLock := false
	isReadUnlock := false
	switch element.GetType(true) {
	case trace.MutexRLock, trace.MutexTryRLock:
		isReadLock = true
	case trace.MutexRUnlock:
		isReadUnlock = true
	}

	lockID := baseA.LockID{ID: element.GetObjId(), ReadLock: isReadLock || isReadUnlock}

	event := baseA.LockEvent{
		ThreadID:    tid,
		TraceID:     element.GetTID(),
		LockID:      element.GetObjId(),
		VectorClock: element.GetWVC().Copy(),
	}

	switch element.GetType(true) {

	// --------- WRITE LOCK ---------
	case trace.MutexLock, trace.MutexTryLock:
		mdPushRD(t, tid, lockID, event, element)
		t.CurrentLockset.Add(lockID)
		t.LockDepth++
		t.WriteDepth++
		log.Debug(fmt.Sprintf("MD phase1: T%d acq(write lock=%d) depth=%d writeDepth=%d",
			tid, element.GetObjId(), t.LockDepth, t.WriteDepth))

	// --------- READ LOCK (RWMutex RLock) ---------
	case trace.MutexRLock, trace.MutexTryRLock:
		if _, ok := t.ReadLockCount[lockID]; !ok {
			t.ReadLockCount[lockID] = 0
		}
		t.ReadLockCount[lockID]++

		if t.ReadLockCount[lockID] == 1 {
			mdPushRD(t, tid, lockID, event, element)
			t.CurrentLockset.Add(lockID)
			t.LockDepth++
		}
		t.ReadDepth++
		log.Debug(fmt.Sprintf("MD phase1: T%d acq(read lock=%d) readDepth=%d lockDepth=%d count=%d",
			tid, element.GetObjId(), t.ReadDepth, t.LockDepth, t.ReadLockCount[lockID]))

	// --------- WRITE UNLOCK ---------
	case trace.MutexUnlock:
		if stack, ok := t.ActiveRDs[lockID]; ok && len(stack) > 0 {
			rd := stack[len(stack)-1]
			t.ActiveRDs[lockID] = stack[:len(stack)-1]
			t.MostRecentRD[lockID] = rd
			if len(t.ActiveRDs[lockID]) == 0 {
				delete(t.ActiveRDs, lockID)
			}
			t.LockDepth--
			t.WriteDepth--
			log.Debug(fmt.Sprintf("MD phase1: T%d rel(write lock=%d) depth=%d writeDepth=%d",
				tid, element.GetObjId(), t.LockDepth, t.WriteDepth))
		}
		t.CurrentLockset.Remove(lockID)

	// --------- READ UNLOCK (RWMutex RUnlock) ---------
	case trace.MutexRUnlock:
		if _, ok := t.ReadLockCount[lockID]; !ok {
			log.Debug(fmt.Sprintf("MD phase1: T%d rel(read lock=%d) - no counter, ignoring",
				tid, element.GetObjId()))
			break
		}

		t.ReadLockCount[lockID]--
		t.ReadDepth--

		if t.ReadLockCount[lockID] == 0 {
			if stack, ok := t.ActiveRDs[lockID]; ok && len(stack) > 0 {
				rd := stack[len(stack)-1]
				t.ActiveRDs[lockID] = stack[:len(stack)-1]
				t.MostRecentRD[lockID] = rd
				if len(t.ActiveRDs[lockID]) == 0 {
					delete(t.ActiveRDs, lockID)
				}
				t.LockDepth--
			}
			t.CurrentLockset.Remove(lockID)
			delete(t.ReadLockCount, lockID)
		}
		log.Debug(fmt.Sprintf("MD phase1: T%d rel(read lock=%d) readDepth=%d lockDepth=%d count=%d",
			tid, element.GetObjId(), t.ReadDepth, t.LockDepth, t.ReadLockCount[lockID]))

	default:
		log.Error(fmt.Sprintf("MD phase1: unknown mutex operation: %s", element.ToString()))
	}
}

// mdPushRD pushes a new RD node onto the stack for the given lock
func mdPushRD(
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
	t.ActiveRDs[lockID] = append(t.ActiveRDs[lockID], rd)
}

// HandleChannelEventForMixedDeadlock processes one channel trace event
func HandleChannelEventForMixedDeadlock(element *trace.ElementChannel) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	opType := element.GetType(true)
	switch opType {
	case trace.ChannelSend, trace.ChannelRecv, trace.ChannelClose:
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

	// In HandleChannelEventForMixedDeadlock, when collecting CS locks, add debug:
	for lockID := range t.CurrentLockset {
		if stack, ok := t.ActiveRDs[lockID]; ok && len(stack) > 0 {
			topRD := stack[len(stack)-1]
			fmt.Printf("DEBUG: CS lock for T%d chan: lock=%d, tPre=%d\n",
				tid, lockID.ID, topRD.Elem.GetTPre())
			assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: true, RD: topRD})
		}
	}

	// PCS locks: released before the channel op
	for lockID, rd := range t.MostRecentRD {
		if _, held := t.CurrentLockset[lockID]; held {
			continue
		}
		assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: false, RD: rd})
	}

	if len(assocRDs) == 0 {
		return
	}

	cd := &mdCDNode{
		Thread:     tid,
		ChanID:     element.GetObjId(),
		OpType:     opType,
		Buffered:   element.IsBuffered(),
		Event:      event,
		AssocRDs:   assocRDs,
		Elem:       element,
		Depth:      t.LockDepth,
		ReadDepth:  t.ReadDepth,
		WriteDepth: t.WriteDepth,
	}
	currentMDState.AllCDs = append(currentMDState.AllCDs, cd)

	log.Debug(fmt.Sprintf("MD phase1: T%d chan(%s) depth=%d readDepth=%d writeDepth=%d assocRDs=%d",
		tid, opType, t.LockDepth, t.ReadDepth, t.WriteDepth, len(assocRDs)))
}

// ---------------------------------------------------------------------------
// Phase 2: Offline CD-CD partner matching
// ---------------------------------------------------------------------------

// CheckForMixedDeadlock as entry point
func CheckForMixedDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	log.Debug(fmt.Sprintf("MD phase2: start partner matching, AllCDs=%d",
		len(currentMDState.AllCDs)))

	cdByChan := make(map[int][]*mdCDNode, len(currentMDState.AllCDs))
	for _, cd := range currentMDState.AllCDs {
		cdByChan[cd.ChanID] = append(cdByChan[cd.ChanID], cd)
	}

	reported := make(map[[2]*trace.ElementChannel]bool)

	for _, cd := range currentMDState.AllCDs {
		switch cd.OpType {

		case trace.ChannelRecv:
			partner := mdFindPartnerForRecv(cd, cdByChan)
			if partner != nil {
				mdCheckAndReport(cd, partner, reported)
			}
			closePartner := mdFindCloseCDNode(cd, cdByChan)
			if closePartner != nil {
				mdCheckAndReport(cd, closePartner, reported)
			}

		case trace.ChannelSend:
			if cd.Buffered {
				continue
			}
			partner := mdFindPartnerCDByElem(cd.Elem.GetPartner(), cdByChan[cd.ChanID])
			if partner == nil {
				continue
			}
			mdCheckAndReport(cd, partner, reported)

		default:
		}
	}
}

// ---------------------------------------------------------------------------
// Partner lookup helpers
// ---------------------------------------------------------------------------

func mdFindPartnerForRecv(recvCD *mdCDNode, cdByChan map[int][]*mdCDNode) *mdCDNode {
	elem := recvCD.Elem
	if elem.GetClosed() {
		return mdFindCloseCDNode(recvCD, cdByChan)
	}
	return mdFindPartnerCDByElem(elem.GetPartner(), cdByChan[recvCD.ChanID])
}

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

func mdFindCloseCDNode(recvCD *mdCDNode, cdByChan map[int][]*mdCDNode) *mdCDNode {
	for _, c := range cdByChan[recvCD.ChanID] {
		if c.OpType != trace.ChannelClose {
			continue
		}
		if c.Thread == recvCD.Thread {
			continue
		}
		return c
	}
	return nil
}

// ---------------------------------------------------------------------------
// MD condition checks and reporting
// ---------------------------------------------------------------------------

func mdCheckAndReport(cdA, cdB *mdCDNode, reported map[[2]*trace.ElementChannel]bool) {
	key := mdPairKey(cdA.Elem, cdB.Elem)
	if reported[key] {
		return
	}

	for _, refA := range cdA.AssocRDs {
		for _, refB := range cdB.AssocRDs {

			if !refA.LockID.EqualsCouldBlock(refB.LockID) {
				continue
			}

			if !refA.IsCS && !refB.IsCS {
				log.Debug(fmt.Sprintf("MD phase2: T%d/T%d ch=%d — both PCS, skip",
					cdA.Thread, cdB.Thread, cdA.ChanID))
				continue
			}

			if refA.IsCS && refB.IsCS && !cdA.Buffered {
				if cdA.OpType != trace.ChannelClose && cdB.OpType != trace.ChannelClose {
					log.Debug("MD phase2: both CS on unbuffered send/recv - skip")
					continue
				}
			}

			if !mdLockAcqAreConcurrent(refA.RD, refB.RD) {
				log.Debug(fmt.Sprintf("MD phase2: T%d/T%d ch=%d — lock acquires not concurrent",
					cdA.Thread, cdB.Thread, cdA.ChanID))
				continue
			}

			holderCD, holderRef, waiterCD, waiterRef := mdDetermineRoles(cdA, refA, cdB, refB)

			log.Debug(fmt.Sprintf("MD phase2: FOUND MD | ch=%d | holder=T%d(CS=%v depth=%d readDepth=%d) waiter=T%d(CS=%v depth=%d readDepth=%d)",
				cdA.ChanID, holderCD.Thread, holderRef.IsCS, holderCD.Depth, holderCD.ReadDepth,
				waiterCD.Thread, waiterRef.IsCS, waiterCD.Depth, waiterCD.ReadDepth))

			mdReportCandidate(holderCD, holderRef, waiterCD, waiterRef)
			reported[key] = true
			return
		}
	}
}

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

// mdDetermineRoles assigns holder and waiter roles using depth for deterministic selection
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
	if refA.IsCS && refB.IsCS {
		// For RWMutex: reader with read lock should be holder
		if cdA.ReadDepth > 0 && cdB.ReadDepth == 0 {
			return cdA, refA, cdB, refB
		}
		if cdB.ReadDepth > 0 && cdA.ReadDepth == 0 {
			return cdB, refB, cdA, refA
		}
		// Otherwise use total depth (deeper stack = inner CS = holder)
		if cdA.Depth > cdB.Depth {
			return cdA, refA, cdB, refB
		}
		if cdB.Depth > cdA.Depth {
			return cdB, refB, cdA, refA
		}
	}

	// Both CS with same depth - use VC order
	if len(refA.RD.Requests) > 0 && len(refB.RD.Requests) > 0 {
		vcA := refA.RD.Requests[0].VectorClock
		vcB := refB.RD.Requests[0].VectorClock
		if vcA != nil && vcB != nil {
			switch clock.GetHappensBefore(vcA, vcB) {
			case hb.Before:
				return cdB, refB, cdA, refA
			case hb.After:
				return cdA, refA, cdB, refB
			}
		}
	}

	// Tie-break: larger tPre = acquired later = holder
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

func mdReportCandidate(
	holderCD *mdCDNode, holderRef mdLockRef,
	waiterCD *mdCDNode, waiterRef mdLockRef,
) {
	if holderCD.Elem == nil || holderRef.RD == nil || holderRef.RD.Elem == nil ||
		waiterCD.Elem == nil || waiterRef.RD == nil || waiterRef.RD.Elem == nil {
		log.Error("MD report: nil element pointer in candidate — skipping")
		return
	}

	// holder's channel element
	holderChanFile, holderChanLine, holderChanTPre, err := trace.InfoFromTID(holderCD.Event.TraceID)
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

	// holder's lock acquire element
	holderLockReq := holderRef.RD.Requests[0]
	holderLockFile, holderLockLine, holderLockTPre, err := trace.InfoFromTID(holderLockReq.TraceID)
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

	// waiter's channel element
	waiterChanFile, waiterChanLine, waiterChanTPre, err := trace.InfoFromTID(waiterCD.Event.TraceID)
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

	// waiter's lock acquire element (stuck element)
	waiterLockReq := waiterRef.RD.Requests[0]
	waiterLockFile, waiterLockLine, waiterLockTPre, err := trace.InfoFromTID(waiterLockReq.TraceID)
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

	stuckElement := waiterLockRes
	stuckElement.ObjType = "DH"

	cycleElements := []results.ResultElem{
		holderChanRes,
		holderLockRes,
		waiterChanRes,
		waiterLockRes,
	}

	log.Debug(fmt.Sprintf("MD report: PMixedDeadlock holder=T%d(ch=%d) waiter=T%d(ch=%d) lock=%d",
		holderCD.Thread, holderCD.ChanID, waiterCD.Thread, waiterCD.ChanID, holderRef.LockID.ID))

	results.Result(
		results.CRITICAL,
		helper.PMixedDeadlock,
		"stuck", []results.ResultElem{stuckElement},
		"cycle", cycleElements,
	)
}

// mdPairKey returns a canonical key for a channel-element pair
func mdPairKey(a, b *trace.ElementChannel) [2]*trace.ElementChannel {
	if a.GetTPre() <= b.GetTPre() {
		return [2]*trace.ElementChannel{a, b}
	}
	return [2]*trace.ElementChannel{b, a}
}
