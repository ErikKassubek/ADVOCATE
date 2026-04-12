// advocate/analysis/analysis/scenarios/mixedDeadlock.go

// Copyright (c) 2024 Erik Kassubek
//
// File: mixedDeadlock.go
// Brief: Analysis for mixed deadlocks involving both mutexes and channels.
//
// Author: Ilian Kohl
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

type mdLockRef struct {
	LockID baseA.LockID
	IsCS   bool
	RD     *mdRDNode
}

type mdRDNode struct {
	Thread   baseA.ThreadID
	Lock     baseA.LockID
	Lockset  baseA.Lockset
	Requests []baseA.LockEvent
}

type mdCDNode struct {
	Thread   baseA.ThreadID
	ChanID   int
	OpType   trace.OperationType
	Buffered bool
	Event    baseA.LockEvent
	AssocRDs []mdLockRef
}

type mdThreadState struct {
	CurrentLockset baseA.Lockset
	ActiveRDs      map[baseA.LockID]*mdRDNode
	MostRecentRD   map[baseA.LockID]*mdRDNode
}

type mdState struct {
	Threads map[baseA.ThreadID]*mdThreadState
	AllRDs  []*mdRDNode
	AllCDs  []*mdCDNode
}

var currentMDState mdState

// ---------------------------------------------------------------------------
// Phase 1: Online Record Dependencies
// ---------------------------------------------------------------------------

// Reset data structures for a new analysis run (analysis.go)
func ResetMixedDeadlockState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)
	currentMDState = mdState{
		Threads: make(map[baseA.ThreadID]*mdThreadState),
	}
}

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

// Record resource requests
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
		mdInsertRD(t, tid, lockID, event)
		t.CurrentLockset.Add(lockID)
		log.Debug(fmt.Sprintf("MD phase1: T%d acq(lock=%d) LS=%v -> RD recorded",
			tid, element.GetObjId(), t.CurrentLockset))

	case trace.MutexUnlock, trace.MutexRUnlock:
		if rd, ok := t.ActiveRDs[lockID]; ok {
			t.MostRecentRD[lockID] = rd
			delete(t.ActiveRDs, lockID)
			log.Debug(fmt.Sprintf("MD phase1: T%d rel(lock=%d) -> moved to MostRecentRD",
				tid, element.GetObjId()))
		}
		t.CurrentLockset.Remove(lockID)
	}
}

// Insert RD nodes for requested lock acquires
func mdInsertRD(t *mdThreadState, tid baseA.ThreadID, lockID baseA.LockID, event baseA.LockEvent) {
	ls := t.CurrentLockset.Clone()
	if existing, ok := t.ActiveRDs[lockID]; ok {
		if existing.Lockset.Equal(ls) {
			existing.Requests = append(existing.Requests, event.Clone())
			return
		}
	}
	rd := &mdRDNode{
		Thread:   tid,
		Lock:     lockID,
		Lockset:  ls,
		Requests: []baseA.LockEvent{event.Clone()},
	}
	t.ActiveRDs[lockID] = rd
	currentMDState.AllRDs = append(currentMDState.AllRDs, rd)
}

// Record channel requests
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

	for lockID := range t.CurrentLockset {
		if rd, ok := t.ActiveRDs[lockID]; ok {
			assocRDs = append(assocRDs, mdLockRef{LockID: lockID, IsCS: true, RD: rd})
		}
	}
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
	}

	currentMDState.AllCDs = append(currentMDState.AllCDs, cd)
	log.Debug(fmt.Sprintf("MD phase1: T%d chan(op=%s, ch=%d) — CD recorded with %d assocRDs (CS=%d PCS=%d)",
		tid, opType, element.GetObjId(), len(assocRDs),
		func() int {
			n := 0
			for _, r := range assocRDs {
				if r.IsCS {
					n++
				}
			}
			return n
		}(),
		func() int {
			n := 0
			for _, r := range assocRDs {
				if !r.IsCS {
					n++
				}
			}
			return n
		}(),
	))
}

// ---------------------------------------------------------------------------
// Phase 2: Offline Cycle Detection
// ---------------------------------------------------------------------------

type mdCycleNode struct {
	RD *mdRDNode
	CD *mdCDNode
}

func (n mdCycleNode) isCD() bool { return n.CD != nil }
func (n mdCycleNode) isRD() bool { return n.RD != nil }

func (n mdCycleNode) thread() baseA.ThreadID {
	if n.isCD() {
		return n.CD.Thread
	}
	return n.RD.Thread
}

// CheckForMixedDeadlock performs DFS to find cyclces in MD graph
// applies feasability filtering, and reports detected cycles (analysis.go)
func CheckForMixedDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	log.Debug(fmt.Sprintf("MD phase2: starting cycle detection. AllRDs=%d AllCDs=%d",
		len(currentMDState.AllRDs), len(currentMDState.AllCDs)))

	visitedCDRoots := make(map[*mdCDNode]bool)
	var stack []mdCycleNode

	for _, cd := range currentMDState.AllCDs {
		if visitedCDRoots[cd] {
			continue
		}
		visitedCDRoots[cd] = true

		log.Debug(fmt.Sprintf("MD phase2: DFS root CD T%d chan=%d op=%s assocRDs=%d",
			cd.Thread, cd.ChanID, cd.OpType, len(cd.AssocRDs)))

		stack = append(stack[:0], mdCycleNode{CD: cd})
		for _, ref := range cd.AssocRDs {
			stack = append(stack[:1], mdCycleNode{RD: ref.RD})
			mdDFS(&stack, cd)
		}
	}
}

// mDFS performs depth-first search to find cycles in MD graph,
// starting from a root CD node, alternates between CD and RD nodes
func mdDFS(stack *[]mdCycleNode, rootCD *mdCDNode) {
	top := (*stack)[len(*stack)-1]
	if !top.isRD() {
		log.Error("mixedDeadlock: mdDFS invariant violated: top of stack is not an RD")
		return
	}
	rdTop := top.RD
	stackThreads := mdStackThreadSet(stack)

	log.Debug(fmt.Sprintf("MD phase2: mdDFS top=RD(T%d lock=%d ls_size=%d) stackLen=%d",
		rdTop.Thread, rdTop.Lock.ID, len(rdTop.Lockset), len(*stack)))

	// RD -> CD (inter-thread)
	for _, cd := range currentMDState.AllCDs {
		if cd.Thread == rdTop.Thread {
			continue
		}
		if stackThreads[cd.Thread] && cd != rootCD {
			continue
		}

		for _, ref := range cd.AssocRDs {
			if !rdTop.Lock.EqualsCouldBlock(ref.LockID) {
				continue
			}

			log.Debug(fmt.Sprintf("MD phase2:   RD->CD edge: RD(T%d lock=%d) -> CD(T%d ch=%d isCS=%v)",
				rdTop.Thread, rdTop.Lock.ID, cd.Thread, cd.ChanID, ref.IsCS))

			if cd == rootCD {
				candidate := make([]mdCycleNode, len(*stack))
				copy(candidate, *stack)
				log.Debug(fmt.Sprintf("MD phase2:   cycle candidate of length %d", len(candidate)))
				if mdIsCycleRoot(rootCD, &candidate) && mdCheckFeasibility(&candidate) {
					log.Debug("MD phase2:   feasibility PASSED -> reporting cycle")
					mdReportCycle(&candidate)
				} else if !mdIsCycleRoot(rootCD, &candidate) {
					log.Debug("MD phase2:   skipping cycle (not canonical root — duplicate rotation)")
				} else {
					log.Debug("MD phase2:   feasibility FAILED -> discarding")
				}
				continue
			}

			*stack = append(*stack, mdCycleNode{CD: cd})
			for _, nextRef := range cd.AssocRDs {
				if mdRDOnStack(stack, nextRef.RD) {
					continue
				}
				*stack = append(*stack, mdCycleNode{RD: nextRef.RD})
				mdDFS(stack, rootCD)
				*stack = (*stack)[:len(*stack)-1]
			}
			*stack = (*stack)[:len(*stack)-1]
		}
	}

	// RD -> RD (inter-thread)
	for _, rdNext := range currentMDState.AllRDs {
		if rdNext.Thread == rdTop.Thread {
			continue
		}
		if stackThreads[rdNext.Thread] {
			continue
		}
		if !mdRDChainCondition(rdTop, rdNext) {
			continue
		}
		if mdRDOnStack(stack, rdNext) {
			continue
		}

		log.Debug(fmt.Sprintf("MD phase2:   RD->RD edge: RD(T%d lock=%d) -> RD(T%d lock=%d)",
			rdTop.Thread, rdTop.Lock.ID, rdNext.Thread, rdNext.Lock.ID))

		*stack = append(*stack, mdCycleNode{RD: rdNext})
		mdDFS(stack, rootCD)
		*stack = (*stack)[:len(*stack)-1]
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mdIsCycleRoot returns true if rootCD has the smallest tPre among all CDs
// in the cycle. Ensures each unique cycle is reported exactly once,
// eliminating duplicate rotations found by rooting from different CD nodes.
func mdIsCycleRoot(rootCD *mdCDNode, cycle *[]mdCycleNode) bool {
	_, _, rootTPre, err := trace.InfoFromTID(rootCD.Event.TraceID)
	if err != nil {
		return true // cannot determine, allow
	}
	for _, node := range *cycle {
		if !node.isCD() {
			continue
		}
		if node.CD == rootCD {
			continue
		}
		_, _, tPre, err := trace.InfoFromTID(node.CD.Event.TraceID)
		if err != nil {
			continue
		}
		if tPre < rootTPre {
			return false // another CD in the cycle has smaller tPre -> not canonical root
		}
	}
	return true
}

func mdRDChainCondition(rdFrom, rdTo *mdRDNode) bool {
	if rdTo.Lockset.Empty() {
		return false
	}
	if !rdFrom.Lockset.DisjointCouldBlock(rdTo.Lockset) {
		return false
	}
	for l := range rdTo.Lockset {
		if rdFrom.Lock.EqualsCouldBlock(l) {
			return true
		}
	}
	return false
}

func mdStackThreadSet(stack *[]mdCycleNode) map[baseA.ThreadID]bool {
	m := make(map[baseA.ThreadID]bool, len(*stack))
	for _, n := range *stack {
		m[n.thread()] = true
	}
	return m
}

func mdRDOnStack(stack *[]mdCycleNode, rd *mdRDNode) bool {
	for _, n := range *stack {
		if n.isRD() && n.RD == rd {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Feasibility filter
// ---------------------------------------------------------------------------

// mdCheckFeasibility checks that each inter-thread RD->CD or RD->RD edge in
// the cycle is feasible under the weak happens-before relation.
//
//	For a RD->CD inter-thread edge, do NOT use the CD's channel-op VC.
//	The channel op VC for a PCS node is AFTER the lock release, which is
//	always happens-after other goroutines' lock acquires, so the check
//	would always fail for PCS dependencies.
//
//	Instead, use the CD's AssocRD's lock-acquire VC (the VC at the time
//	the goroutine acquired the lock for this CS/PCS). This VC represents
//	the actual point of contention and may be concurrent with the RD's
//	lock-acquire VC.
//
//	For each RD node, find the AssocRD of the NEXT CD in the cycle that
//	corresponds to the RD->CD edge (matched by lock id), and use that
//	AssocRD's earliest request VC for the feasibility check.
//
// For RD->RD edges: use request VCs of both RDs
func mdCheckFeasibility(cycle *[]mdCycleNode) bool {
	n := len(*cycle)
	for i, node := range *cycle {
		if !node.isRD() {
			continue
		}
		rd := node.RD
		nextIdx := (i + 1) % n
		nextNode := (*cycle)[nextIdx]

		if nextNode.isCD() {
			// Inter-thread RD->CD edge.
			// Find the AssocRD of nextCD that matches this RD's lock.
			nextCD := nextNode.CD
			var assocRDVCs []*clock.VectorClock
			for _, ref := range nextCD.AssocRDs {
				if rd.Lock.EqualsCouldBlock(ref.LockID) {
					// Use the lock-acquire VCs from the AssocRD, not the channel-op VC.
					for _, req := range ref.RD.Requests {
						if req.VectorClock != nil {
							assocRDVCs = append(assocRDVCs, req.VectorClock)
						}
					}
					break
				}
			}

			if len(assocRDVCs) == 0 {
				// Cannot find matching AssocRD VC — use channel-op VC as fallback.
				if nextCD.Event.VectorClock == nil {
					log.Debug(fmt.Sprintf("MD feasibility: RD node %d (RD->CD) nil VC — skipping", i))
					continue
				}
				assocRDVCs = []*clock.VectorClock{nextCD.Event.VectorClock}
			}

			// Check: at least one request VC of RD is concurrent with at least
			// one lock-acquire VC of the next CD's AssocRD.
			found := false
		outerCS:
			for _, rdVC := range mdNodeVCs(node) {
				if rdVC == nil {
					continue
				}
				for _, cdAcqVC := range assocRDVCs {
					if clock.GetHappensBefore(rdVC, cdAcqVC) == hb.Concurrent {
						found = true
						break outerCS
					}
				}
			}
			if !found {
				log.Debug(fmt.Sprintf("MD feasibility: RD node %d (RD->CD T%d ch=%d isCS=%v) failed concurrency check",
					i, nextCD.Thread, nextCD.ChanID,
					func() bool {
						for _, ref := range nextCD.AssocRDs {
							if rd.Lock.EqualsCouldBlock(ref.LockID) {
								return ref.IsCS
							}
						}
						return false
					}()))
				return false
			}

		} else {
			// Inter-thread RD->RD edge. Use request VCs of both.
			nextRD := nextNode.RD
			prevIdx := (n + i - 1) % n
			prevNode := (*cycle)[prevIdx]

			prevVCs := mdNodeVCs(prevNode)
			nextVCs := mdNodeVCs(nextNode)

			if mdAnyNilVC(prevVCs) || mdAnyNilVC(nextVCs) {
				log.Debug(fmt.Sprintf("MD feasibility: RD node %d (RD->RD) has nil VC — skipping", i))
				continue
			}

			found := false
		outerRR:
			for _, vc := range mdNodeVCs(node) {
				if vc == nil {
					continue
				}
				for _, pvc := range prevVCs {
					if clock.GetHappensBefore(vc, pvc) != hb.Concurrent {
						continue
					}
					for _, nvc := range nextVCs {
						if clock.GetHappensBefore(vc, nvc) == hb.Concurrent {
							found = true
							break outerRR
						}
					}
				}
			}
			if !found {
				log.Debug(fmt.Sprintf("MD feasibility: RD node %d (RD->RD T%d lock=%d) failed concurrency check",
					i, nextRD.Thread, nextRD.Lock.ID))
				return false
			}
		}
	}
	return true
}

func mdAnyNilVC(vcs []*clock.VectorClock) bool {
	for _, vc := range vcs {
		if vc == nil {
			return true
		}
	}
	return len(vcs) == 0
}

func mdNodeVCs(n mdCycleNode) []*clock.VectorClock {
	if n.isCD() {
		return []*clock.VectorClock{n.CD.Event.VectorClock}
	}
	vcs := make([]*clock.VectorClock, len(n.RD.Requests))
	for i, r := range n.RD.Requests {
		vcs[i] = r.VectorClock
	}
	return vcs
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

func mdReportCycle(cycle *[]mdCycleNode) {
	var lockElems []results.ResultElem
	var chanElems []results.ResultElem

	for idx, node := range *cycle {
		if node.isRD() {
			rd := node.RD
			req := mdFindEarliestRequest(rd)
			file, line, tPre, err := trace.InfoFromTID(req.TraceID)
			if err != nil {
				log.Error("mixedDeadlock: InfoFromTID for RD: ", err.Error())
				return
			}
			log.Debug(fmt.Sprintf("MD report: cycle[%d] RD T%d lock=%d tPre=%d %s:%d",
				idx, rd.Thread, rd.Lock.ID, tPre, file, line))
			lockElems = append(lockElems, results.TraceElementResult{
				RoutineID: int(rd.Thread),
				ObjID:     req.LockID,
				TPre:      tPre,
				ObjType:   "DC",
				File:      file,
				Line:      line,
			})
		} else {
			cd := node.CD
			file, line, tPre, err := trace.InfoFromTID(cd.Event.TraceID)
			if err != nil {
				log.Error("mixedDeadlock: InfoFromTID for CD: ", err.Error())
				return
			}
			log.Debug(fmt.Sprintf("MD report: cycle[%d] CD T%d ch=%d op=%s tPre=%d %s:%d",
				idx, cd.Thread, cd.ChanID, cd.OpType, tPre, file, line))
			chanElems = append(chanElems, results.TraceElementResult{
				RoutineID: int(cd.Thread),
				ObjID:     cd.ChanID,
				TPre:      tPre,
				ObjType:   cd.OpType,
				File:      file,
				Line:      line,
			})
		}
	}

	if len(lockElems) == 0 || len(chanElems) == 0 {
		log.Error("mixedDeadlock: cycle has no RD or no CD elements — skipping report")
		return
	}

	log.Debug(fmt.Sprintf("MD report: PMixedDeadlock with %d lock elems and %d chan elems",
		len(lockElems), len(chanElems)))

	results.Result(
		results.CRITICAL,
		helper.PMixedDeadlock,
		"stuck", lockElems,
		"cycle", chanElems,
	)
}

func mdFindEarliestRequest(rd *mdRDNode) baseA.LockEvent {
	if len(rd.Requests) == 0 {
		return baseA.LockEvent{}
	}
	earliest := rd.Requests[0]
	_, _, earliestTime, err := trace.InfoFromTID(earliest.TraceID)
	if err != nil {
		return earliest
	}
	for _, r := range rd.Requests[1:] {
		_, _, t, err := trace.InfoFromTID(r.TraceID)
		if err != nil {
			continue
		}
		if t < earliestTime {
			earliest = r
			earliestTime = t
		}
	}
	return earliest
}
