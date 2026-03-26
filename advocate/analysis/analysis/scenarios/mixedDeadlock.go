// advocate/analysis/analysis/scenarios/mixedDeadlock.go

// Copyright (c) 2024 Erik Kassubek
// File: mixedDeadlock.go
// Brief: Two-phase mixed deadlock detection.
//
// Author: Ilian Kohl
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/baseA"
	"advocate/analysis/hb"
	"advocate/analysis/hb/clock"
	"advocate/analysis/hb/vc"
	"advocate/results/results"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
)

// -----------------------------------------------------------------------
// Two-phase mixed deadlock detection
// -----------------------------------------------------------------------
// Phase 1 (online):  mutex events     -> AcqHist / RelHist  (mutex.go)
//                    channel events   -> CommHist           (channel.go)
// Phase 2 (offline): build RD ∪ CD dependency graph, find cycles, filter, report

// -----------------------------------------------------------------------
// Data Structures
// -----------------------------------------------------------------------

// ResetMixedDeadlockState resets all MD-specific recording state
func ResetMixedDeadlockState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	baseA.CurrentMDState = baseA.MDState{
		AcqHist:  make(map[int]map[int][]baseA.MDLockHistEntry), // routine -> lockID -> list of acquires
		RelHist:  make(map[int]map[int][]baseA.MDLockHistEntry), // routine -> lockID -> list of releases
		CommHist: make(map[int]map[int][]baseA.MDCommHistEntry), // routine -> chanID -> list of comm events
	}
}

// MDNodeKind for RD and CD nodes in graph
type mdNodeKind int

const (
	mdNodeRD mdNodeKind = iota // resource (lock) dependency
	mdNodeCD                   // communication dependency
)

// mdRD wraps lock dependency: RD = (T, l, LS, Req)
type mdRD struct {
	dep baseA.LockDependency
}

// mdCD wraps communication dependency: CD = (T, op(c)?, LS, Req)
type mdCD struct {
	routine  int
	chanID   int
	op       trace.OperationType
	lockset  mdLockset // contains InCS/PCS info
	buffered bool
	req      baseA.MDCommHistEntry
}

// mdLockset maps lockID -> mdLockEntry
type mdLockset map[int]mdLockEntry

// mdLockEntry is ONE entry in a CD's lockset
type mdLockEntry struct {
	acq   baseA.MDLockHistEntry
	inCS  bool // inCS:  the routine held this lock at channel-event time
	isPCS bool // isPCS: the routine had a preceding CS on this lock
}

// mdNode as tagged union of mdRD and mdCD
type mdNode struct {
	kind mdNodeKind
	rd   *mdRD
	cd   *mdCD
}

// -----------------------------------------------------------------------
// Online recording of dependencies
// -----------------------------------------------------------------------

// HandleChannelEventForMixedDeadlock records a completed channel event
// Pupulates MD communication history (baseA.CurrentMDState.CommHist)
func HandleChannelEventForMixedDeadlock(ch *trace.ElementChannel) {
	if ch.GetTPost() == 0 {
		return
	}
	routine := ch.GetRoutine()
	entry := baseA.MDCommHistEntry{
		Routine:  routine,
		ChanID:   ch.GetObjId(),
		Op:       ch.GetType(true),
		OID:      ch.GetOID(),
		TraceID:  ch.GetTID(),
		VC:       vc.CurrentVC[routine].Copy(),
		WVC:      vc.CurrentWVC[routine].Copy(),
		Elem:     ch,
		Buffered: ch.IsBuffered(),
	}
	baseA.RecordMDCommEvent(entry)
}

// HandleMutexEventForMixedDeadlock records a mutex acquire or release
// Populates MD lock history (baseA.CurrentMDState.AcqHist / RelHist)
func HandleMutexEventForMixedDeadlock(mu *trace.ElementMutex) {
	routine := mu.GetRoutine()

	entry := baseA.MDLockHistEntry{
		Routine: routine,
		LockID:  mu.GetObjId(),
		TraceID: mu.GetTID(),
		VC:      vc.CurrentVC[routine].Copy(),
		WVC:     vc.CurrentWVC[routine].Copy(),
	}

	switch mu.GetType(true) {
	case trace.MutexLock, trace.MutexTryLock:
		entry.IsRead = false
		entry.IsAcq = true
		baseA.RecordMDAcquire(entry)
	case trace.MutexRLock, trace.MutexTryRLock:
		entry.IsRead = true
		entry.IsAcq = true
		baseA.RecordMDAcquire(entry)
	case trace.MutexUnlock:
		entry.IsRead = false
		entry.IsAcq = false
		baseA.RecordMDRelease(entry)
	case trace.MutexRUnlock:
		entry.IsRead = true
		entry.IsAcq = false
		baseA.RecordMDRelease(entry)
	default:
		return
	}
}

// -----------------------------------------------------------------------
// Lockset reconstruction
// -----------------------------------------------------------------------

// reconstructLockset returns  lock context of `routine` at `eventVC`
//
// For each lock x acquired by routine:
//   - nacq = number of acquires with VC <HB eventVC
//   - nrel = number of releases with VC <HB eventVC
//   - nacq > nrel  -> InCS
//   - nacq == nrel -> PCS
//
// Counting handles re-entrant RLocks correctly (2 RLocks - 1 RUnlock = InCS)
func reconstructLockset(routine int, eventVC *clock.VectorClock) mdLockset {
	ls := make(mdLockset)
	if eventVC == nil {
		return ls
	}

	acqMap := baseA.CurrentMDState.AcqHist[routine]
	relMap := baseA.CurrentMDState.RelHist[routine]

	for lockID, acqList := range acqMap {
		nacq := 0
		var latestAcq *baseA.MDLockHistEntry
		for i := range acqList {
			a := &acqList[i]
			if a.VC == nil || clock.GetHappensBefore(a.VC, eventVC) != hb.Before {
				continue
			}
			nacq++
			if latestAcq == nil || clock.GetHappensBefore(latestAcq.VC, a.VC) == hb.Before {
				latestAcq = a
			}
		}
		if nacq == 0 || latestAcq == nil {
			continue
		}

		nrel := 0
		for i := range relMap[lockID] {
			r := &relMap[lockID][i]
			if r.VC != nil && clock.GetHappensBefore(r.VC, eventVC) == hb.Before {
				nrel++
			}
		}

		if nacq > nrel {
			ls[lockID] = mdLockEntry{acq: *latestAcq, inCS: true}
		} else {
			ls[lockID] = mdLockEntry{acq: *latestAcq, isPCS: true}
		}
	}
	return ls
}

// buildCD constructs CD node from recorded comm event
func buildCD(g baseA.MDCommHistEntry) *mdCD {
	return &mdCD{
		routine:  g.Routine,
		chanID:   g.ChanID,
		op:       g.Op,
		lockset:  reconstructLockset(g.Routine, g.VC),
		buffered: g.Buffered,
		req:      g,
	}
}

// hasRelevantLock returns true if CD has at least one InCS or PCS lock
func (cd *mdCD) hasRelevantLock() bool {
	return len(cd.lockset) > 0
}

// -----------------------------------------------------------------------
// Typed dependency edges
// -----------------------------------------------------------------------

// edgeRD_RD: l_a ∈ LS_b = RD a requests lock that RD b holds (a blocks on b)
func edgeRD_RD(a, b *mdRD) bool {
	for l := range b.dep.Lockset {
		if a.dep.Lock.EqualsCouldBlock(l) {
			return true
		}
	}
	return false
}

// edgeRD_CD: l_rd ∈ InCS(cd) = CD holds lock RD is requesting (CD blocks RD)
func edgeRD_CD(rd *mdRD, cd *mdCD) bool {
	for lockID, entry := range cd.lockset {
		if entry.inCS && rd.dep.Lock.ID == lockID {
			return true
		}
	}
	return false
}

// edgeCD_CD detects two-cycle cases:
//
//	MD2-1B  InCS/InCS  buffered            both hold lock; channel async
//	MD2-1U  InCS/InCS  unbuffered snd/rcv  rejected - not observable / mutually exclusive
//	MD2-2U  InCS/PCS   unbuffered          sender blocks on send; recv can't re-acquire
//	MD2-2B  InCS/PCS   buffered            rejected - send unblocking and recv lock acq/rel unblocking
//	MD2-3   PCS/InCS   any                 receiver blocks on recv; sender can't re-acquire
//	        InCS/PCS   close side          rejected - close is non-blocking
//	        PCS/PCS    any                 rejected - neither side blocking
//	        Read/Read  any                 rejected - RLocks compatible
func edgeCD_CD(a, b *mdCD) bool {
	if a.chanID != b.chanID || a.routine == b.routine {
		return false
	}

	aIsSend := a.op == trace.ChannelSend
	bIsSend := b.op == trace.ChannelSend
	aIsRecv := a.op == trace.ChannelRecv
	bIsRecv := b.op == trace.ChannelRecv
	aIsClose := a.op == trace.ChannelClose
	bIsClose := b.op == trace.ChannelClose

	if !(aIsSend && bIsRecv) && !(aIsClose && bIsRecv) &&
		!(bIsSend && aIsRecv) && !(bIsClose && aIsRecv) {
		return false
	}

	for lockID, aEntry := range a.lockset {
		bEntry, ok := b.lockset[lockID]
		if !ok {
			continue
		}
		if aEntry.acq.IsRead && bEntry.acq.IsRead {
			continue // Read/Read
		}
		if aEntry.isPCS && bEntry.isPCS {
			continue // PCS/PCS
		}
		if aEntry.inCS && bEntry.inCS {
			if !a.buffered && !b.buffered && ((aIsSend && bIsRecv) || (bIsSend && aIsRecv)) {
				continue // MD2-1U
			}
			return true // MD2-1B or close/recv InCS/InCS
		}
		// InCS/PCS or PCS/InCS
		senderInCS_recvPCS := (aIsSend && aEntry.inCS && bIsRecv && bEntry.isPCS) ||
			(bIsSend && bEntry.inCS && aIsRecv && aEntry.isPCS)
		closeInCS_recvPCS := (aIsClose && aEntry.inCS && bIsRecv && bEntry.isPCS) ||
			(bIsClose && bEntry.inCS && aIsRecv && aEntry.isPCS)
		if closeInCS_recvPCS {
			continue // MD2-2-Close
		}
		if senderInCS_recvPCS && (a.buffered || b.buffered) {
			continue // MD2-2B
		}
		return true // MD2-2U or MD2-3
	}
	return false
}

// -----------------------------------------------------------------------
// Graph construction and cycle search
// -----------------------------------------------------------------------

type mdGraph struct {
	nodes []mdNode
	adj   [][]int
}

func buildGraph(rds []mdRD, cds []mdCD) mdGraph {
	n := len(rds) + len(cds)
	g := mdGraph{
		nodes: make([]mdNode, 0, n),
		adj:   make([][]int, 0, n),
	}
	for i := range rds {
		g.nodes = append(g.nodes, mdNode{kind: mdNodeRD, rd: &rds[i]})
		g.adj = append(g.adj, nil)
	}
	for i := range cds {
		g.nodes = append(g.nodes, mdNode{kind: mdNodeCD, cd: &cds[i]})
		g.adj = append(g.adj, nil)
	}
	for i, ni := range g.nodes {
		for j, nj := range g.nodes {
			if i == j {
				continue
			}
			var has bool
			switch {
			case ni.kind == mdNodeRD && nj.kind == mdNodeRD:
				has = edgeRD_RD(ni.rd, nj.rd)
			case ni.kind == mdNodeRD && nj.kind == mdNodeCD:
				has = edgeRD_CD(ni.rd, nj.cd)
			case ni.kind == mdNodeCD && nj.kind == mdNodeCD:
				has = edgeCD_CD(ni.cd, nj.cd)
			}
			if has {
				g.adj[i] = append(g.adj[i], j)
			}
		}
	}
	return g
}

type mdCycle []int

// findCycles finds all simple cycles using Johnson-style DFS
// Only follow edges to nodes with index > start to avoid duplicate rotations
func findCycles(g *mdGraph) []mdCycle {
	n := len(g.nodes)
	var cycles []mdCycle
	onStack := make([]bool, n)
	var stack []int

	var dfs func(start, cur int)
	dfs = func(start, cur int) {
		onStack[cur] = true
		stack = append(stack, cur)
		for _, next := range g.adj[cur] {
			if next == start && len(stack) >= 2 {
				c := make(mdCycle, len(stack))
				copy(c, stack)
				cycles = append(cycles, c)
			} else if !onStack[next] && next > start {
				dfs(start, next)
			}
		}
		stack = stack[:len(stack)-1]
		onStack[cur] = false
	}

	for start := 0; start < n; start++ {
		dfs(start, start)
	}
	return cycles
}

// -----------------------------------------------------------------------
// Feasibility filters
// -----------------------------------------------------------------------

// wvcForNode returns WVCs used by the WMHB filter
func wvcForNode(node mdNode) []*clock.VectorClock {
	switch node.kind {
	case mdNodeRD:
		wvcs := make([]*clock.VectorClock, 0, len(node.rd.dep.Requests))
		for _, r := range node.rd.dep.Requests {
			if r.VectorClock != nil {
				wvcs = append(wvcs, r.VectorClock)
			}
		}
		return wvcs
	case mdNodeCD:
		var wvcs []*clock.VectorClock
		for _, entry := range node.cd.lockset {
			if entry.acq.WVC != nil {
				wvcs = append(wvcs, entry.acq.WVC)
			}
		}
		return wvcs
	}
	return nil
}

// passesWMHBFilter rejects cycle if any consecutive node pair has no concurrent WVC pair
func passesWMHBFilter(g *mdGraph, cycle mdCycle) bool {
	n := len(cycle)
	for i := 0; i < n; i++ {
		wvcA := wvcForNode(g.nodes[cycle[i]])
		wvcB := wvcForNode(g.nodes[cycle[(i+1)%n]])
		if len(wvcA) == 0 || len(wvcB) == 0 {
			continue
		}
		anyConcurrent := false
		for _, va := range wvcA {
			for _, vb := range wvcB {
				if clock.GetHappensBefore(va, vb) == hb.Concurrent {
					anyConcurrent = true
					break
				}
			}
			if anyConcurrent {
				break
			}
		}
		if !anyConcurrent {
			return false
		}
	}
	return true
}

// passesRWFilter rejects cycles where all RD nodes involve only read locks
func passesRWFilter(g *mdGraph, cycle mdCycle) bool {
	hasRD := false
	for _, idx := range cycle {
		if g.nodes[idx].kind == mdNodeRD {
			hasRD = true
			if g.nodes[idx].rd.dep.Lock.IsWrite() {
				return true
			}
		}
	}
	return !hasRD
}

// -----------------------------------------------------------------------
// Reporting
// -----------------------------------------------------------------------
// Encoding layout in TraceElement2 (cycle witness list), one block per node:
//
//	RD node:
//	- ElementMutex  ObjType "DC"
//	-> rewriter shortens to DC.tPre (re-executes lock request, blocks)
//
//	CD node, InCS side (routine holds lock at channel-op time):
//	- ElementChannel  ObjType "CB"  - channel blocking point
//	- ElementMutex    ObjType "DC"  - InCS lock acquire
//	-> rewriter shortens to CB.tPre (re-executes channel op while holding lock)
//	-> DC gives lock ID and original acquire time for ordering
//
//	CD node, PCS side (routine released lock before channel op):
//	- ElementMutex    ObjType "MB"  	- PCS lock acquire (already released)
//	- ElementChannel  ObjType CS/CR/CC  - informational only
//	-> rewriter shortens to MB.tPre (re-executes lock acquire, which blocks because InCS side still holds it
//
// ObjType tags signal to rewriter:
// - "CB" -> ShortenRoutine to this element's tPre
// - "MB" -> ShortenRoutine to this element's tPre
// - "DC" -> collect lock ID; used for acquire ordering (not for shortening)
// - CS/CR/CC -> ignored by rewriter (informational)
func reportMixedDeadlockCycle(g *mdGraph, cycle mdCycle) {
	var cycleElements []results.ResultElem

	for _, idx := range cycle {
		node := g.nodes[idx]

		switch node.kind {

		// RD node
		case mdNodeRD:
			if len(node.rd.dep.Requests) == 0 {
				continue
			}
			req := node.rd.dep.Requests[0]
			f, l, tp, err := trace.InfoFromTID(req.TraceID)
			if err != nil {
				log.Error("mixedDeadlock report RD: InfoFromTID: ", err.Error())
				continue
			}
			cycleElements = append(cycleElements, results.TraceElementResult{
				RoutineID: int(req.ThreadID),
				ObjID:     req.LockID,
				TPre:      tp,
				ObjType:   trace.OperationType("DC"),
				File:      f,
				Line:      l,
			})

		// CD node
		case mdNodeCD:
			req := node.cd.req

			cf, cl, ctp, cerr := trace.InfoFromTID(req.TraceID)
			if cerr != nil {
				log.Error("mixedDeadlock report CD chan: InfoFromTID: ", cerr.Error())
				continue
			}

			var chanObjType trace.OperationType
			switch req.Op {
			case trace.ChannelSend:
				chanObjType = trace.ChannelSend
			case trace.ChannelRecv:
				chanObjType = trace.ChannelRecv
			case trace.ChannelClose:
				chanObjType = trace.ChannelClose
			default:
				chanObjType = trace.UnknownOperation
			}

			// Find best lock entry: prefer InCS; within same category take latest by VC.
			var bestLockID int
			var bestEntry *mdLockEntry
			for lockID := range node.cd.lockset {
				e := node.cd.lockset[lockID]
				eCopy := e
				if bestEntry == nil {
					bestEntry = &eCopy
					bestLockID = lockID
					continue
				}
				if e.inCS && !bestEntry.inCS {
					bestEntry = &eCopy
					bestLockID = lockID
					continue
				}
				if e.inCS == bestEntry.inCS &&
					clock.GetHappensBefore(bestEntry.acq.VC, e.acq.VC) == hb.Before {
					bestEntry = &eCopy
					bestLockID = lockID
				}
			}
			_ = bestLockID

			if bestEntry == nil {
				// No lock context - emit channel element informally only
				cycleElements = append(cycleElements, results.TraceElementResult{
					RoutineID: req.Routine,
					ObjID:     req.ChanID,
					TPre:      ctp,
					ObjType:   chanObjType,
					File:      cf,
					Line:      cl,
				})
				continue
			}

			af, al, atp, aerr := trace.InfoFromTID(bestEntry.acq.TraceID)
			if aerr != nil {
				log.Error("mixedDeadlock report CD acq: InfoFromTID: ", aerr.Error())
				cycleElements = append(cycleElements, results.TraceElementResult{
					RoutineID: req.Routine,
					ObjID:     req.ChanID,
					TPre:      ctp,
					ObjType:   chanObjType,
					File:      cf,
					Line:      cl,
				})
				continue
			}

			if bestEntry.inCS {
				// InCS: goroutine holds lock and blocks on channel op
				// CB = channel blocking point, DC = lock acquire for ordering
				cycleElements = append(cycleElements,
					results.TraceElementResult{
						RoutineID: req.Routine,
						ObjID:     req.ChanID,
						TPre:      ctp,
						ObjType:   trace.OperationType("CB"),
						File:      cf,
						Line:      cl,
					},
					results.TraceElementResult{
						RoutineID: bestEntry.acq.Routine,
						ObjID:     bestEntry.acq.LockID,
						TPre:      atp,
						ObjType:   trace.OperationType("DC"),
						File:      af,
						Line:      al,
					},
				)
			} else {
				// PCS: goroutine released lock; blocking point is re-acquiring it
				// MB = mutex blocking point, channel element is informational
				cycleElements = append(cycleElements,
					results.TraceElementResult{
						RoutineID: bestEntry.acq.Routine,
						ObjID:     bestEntry.acq.LockID,
						TPre:      atp,
						ObjType:   trace.OperationType("MB"),
						File:      af,
						Line:      al,
					},
					results.TraceElementResult{
						RoutineID: req.Routine,
						ObjID:     req.ChanID,
						TPre:      ctp,
						ObjType:   chanObjType,
						File:      cf,
						Line:      cl,
					},
				)
			}
		}
	}

	if len(cycleElements) == 0 {
		return
	}

	// stuckElem (arg1, ObjType "DH") is the last "CB" in the list
	// the channel op that will be observed as stuck during replay
	var stuckElem results.TraceElementResult
	foundStuck := false
	for i := len(cycleElements) - 1; i >= 0; i-- {
		te, ok := cycleElements[i].(results.TraceElementResult)
		if !ok {
			continue
		}
		if string(te.ObjType) == "CB" {
			stuckElem = te
			stuckElem.ObjType = "DH"
			foundStuck = true
			break
		}
	}
	if !foundStuck {
		stuckElem = cycleElements[len(cycleElements)-1].(results.TraceElementResult)
		stuckElem.ObjType = "DH"
	}

	results.Result(
		results.CRITICAL,
		helper.PMixedDeadlock,
		"stuck",
		[]results.ResultElem{stuckElem},
		"cycle",
		cycleElements,
	)

	log.Info("Mixed deadlock cycle reported with ", len(cycle), " nodes")
}

// -----------------------------------------------------------------------
// Offline cycle detection
// -----------------------------------------------------------------------

// CheckForMixedDeadlock runs offline detection phase after trace loop
func CheckForMixedDeadlock() {
	// Step 1: collect RD objects from CurrentState
	// CurrentState is populated by HandleMutexEventForRessourceDeadlock which
	// analysis.go calls whenever MixedDeadlock is active.  ResetState() in
	// analysis.go resets it for both ResourceDeadlock and MixedDeadlock, so
	// MD-only runs also get a correctly populated CurrentState.
	var rds []mdRD
	for _, thread := range baseA.CurrentState.Threads {
		for lock, deps := range thread.LockDependencies {
			for _, dep := range deps {
				if len(dep.Requests) == 0 {
					continue
				}
				rds = append(rds, mdRD{
					dep: baseA.LockDependency{
						Thread:   dep.Requests[0].ThreadID,
						Lock:     lock,
						Lockset:  dep.Lockset,
						Requests: dep.Requests,
					},
				})
			}
		}
	}

	// Step 2: construct CD objects; discard events with no lock context
	var cds []mdCD
	for _, chanMap := range baseA.CurrentMDState.CommHist {
		for _, entries := range chanMap {
			for _, g := range entries {
				cd := buildCD(g)
				if cd.hasRelevantLock() {
					cds = append(cds, *cd)
				}
			}
		}
	}

	log.Info("Mixed deadlock: ", len(rds), " RD nodes, ", len(cds), " CD nodes")

	if len(cds) == 0 {
		return
	}

	// Step 3: build typed dependency graph G = (V_RD ∪ V_CD, E)
	g := buildGraph(rds, cds)
	if len(g.nodes) < 2 {
		return
	}

	// Step 4: find all simple cycles
	cycles := findCycles(&g)
	log.Info("Mixed deadlock: ", len(cycles), " raw cycle(s) found")
	if len(cycles) == 0 {
		return
	}

	// Steps 5-6: apply feasibility filters and report survivors
	reported := 0
	for _, cycle := range cycles {
		if !passesWMHBFilter(&g, cycle) {
			continue
		}

		if !passesRWFilter(&g, cycle) {
			continue
		}
		reportMixedDeadlockCycle(&g, cycle)
		reported++
	}

	log.Info("Mixed deadlock analysis complete: ", reported, " cycle(s) reported")
}
