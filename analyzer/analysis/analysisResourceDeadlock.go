// Copyright (c) 2024 Erik Kassubek
//
// File: analysisResourceDeadlock.go
// Brief: Alternative analysis for cyclic mutex deadlocks.
//
// Author: Sebastian Pohsner
// Created: 2025-01-01
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/results"
	"analyzer/timer"
	"analyzer/utils"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Computation of "abstract" lock dependencies
// Lock dependencies are 3-tuples of the following form:
//    (ThreadID, Lock, LockSet)
// Lock dependencies are computed thread-local.
// For each thread there might be several (acquire) events that lead to "lock" acquired under some "lockset".
//
// Each acquire event carries its own vector clock.
// We wish to make use of vector clocks to elimiante infeasible replay candidates.
//
// This means that lock dependencies are 4-tuples of the following form:
//    (ThreadID, Lock, LockSet, []Event)

////////////////////////
// DATA STRUCTURES

// Lock dependencies are computed thread-local. We make use of the following structures.

type Thread struct {
	lock_dependencies map[LockId][]Dependency
	current_lockset   Lockset        // The thread's current lockset.
	readerCounter     map[LockId]int // Store how many readers a readlock has
}

// Unfortunately, we can't use double-indexed map of the following form in Go.
// type Deps map[Lock]map[Lockset][]Event
// Hence, we introduce some intermediate structure.

type Dependency struct {
	lockset  Lockset
	requests []LockEvent
}

// Representation of vector clocks, events, threads, lock and lockset.

type LockEvent struct {
	thread_id    ThreadId
	trace_id     string
	lock_id      int
	vector_clock *clock.VectorClock
}

type ThreadId int
type LockId struct {
	id       int
	readLock bool
}
type Lockset map[LockId]struct{}

///////////////////////////////
// ALGORITHM
//
// There are two phases.
//  1. Recording of lock dependencies.
//  2. Checking if lock dependencies imply a cycle.

type State struct {
	threads map[ThreadId]Thread // Recording lock dependencies in phase 1
	cycles  []Cycle             // Computing cycles in phase 2
	failed  bool                // Analysis failed (encountered unsupported lock action)
}

var currentState State

// Algorithm phase 1

// We show the event processing functions for acquire and release.

func acquire(s *State, read_lock bool, event LockEvent) {
	if _, exists := s.threads[event.thread_id]; !exists {
		s.threads[event.thread_id] = Thread{
			current_lockset:   make(Lockset),
			lock_dependencies: make(map[LockId][]Dependency),
			readerCounter:     make(map[LockId]int),
		}
	}

	lockId := LockId{event.lock_id, read_lock}

	ls := s.threads[event.thread_id].current_lockset
	if !ls.empty() {
		deps := s.threads[event.thread_id].lock_dependencies
		deps[lockId] = insert(deps[lockId], ls, event.Clone())
	}

	if lockId.isRead() {
		lockId.addReader(s.threads[event.thread_id])
	}
	s.threads[event.thread_id].current_lockset.add(lockId)
}

func release(s *State, read_lock bool, event LockEvent) {
	lockId := LockId{event.lock_id, read_lock}
	if lockId.isRead() {
		lockId.removeReader(s.threads[event.thread_id])
		for _, thread := range s.threads {
			if lockId.hasReaders(thread) {
				continue
			}
			thread.current_lockset.remove(lockId)
		}
		s.threads[event.thread_id].current_lockset.remove(lockId)
	} else {
		if !s.threads[event.thread_id].current_lockset.remove(lockId) {
			logAbortReason("Lock not found in lockset! Has probably been released in another thread, this is an unsupported case.")
			s.failed = true
		}
	}
}

// Insert a new lock dependency for a given thread and lock x.
// We assume that event e acquired lock x.
// We might have already an entry that shares the same lock and lockset!
func insert(dependencies []Dependency, ls Lockset, event LockEvent) []Dependency {
	for i, v := range dependencies {
		if v.lockset.equal(ls) {
			dependencies[i].requests = append(dependencies[i].requests, event)
			return dependencies
		}
	}
	return append(dependencies, Dependency{ls.Clone(), []LockEvent{event}})
}

// The above insert function records all requests that share the same dependency (tid,l,ls).
// In case of loops, we may end up with many request entrys.
// For performance reasons, we may want to reduce their size.
//
// Eviction strategy.
// Insert variant where we evict event an already stored event f by e,
// if in between f and e no intra-thread synchronization took place.
// This can be checked via helper function equalModuloTID.
// Assumption: Vector clocks underapproximate the must happen-before relation.
func insert2(dependencies []Dependency, lockset Lockset, event LockEvent) []Dependency {
	// Helper function.
	// Assumes that vc1 and vc2 are connected to two events that are from the same thread tid.
	// Yields true if vc1[k] == vc2[k] for all threads k but tid.
	// Since vc1 and vc2 are underapproximations of the must happen before relation and ignores locks, we also need to check tid itself
	equalModuloTID := func(tid ThreadId, vc1 *clock.VectorClock, vc2 *clock.VectorClock) bool {
		if vc1.GetSize() != vc2.GetSize() {
			return false
		}

		for i := 1; i <= vc1.GetSize(); i++ {
			// if i == int(tid) {
			// 	continue
			// }

			if vc1.GetValue(i) != vc2.GetValue(i) {
				return false
			}
		}

		return true
	}

	for i, v := range dependencies {
		if v.lockset.equal(lockset) {
			add_vc := true

			for _, f := range dependencies[i].requests {
				if equalModuloTID(event.thread_id, event.vector_clock, f.vector_clock) {
					// dependencies[i].requests[j] = event // We want to keep the first request for a better replay
					fmt.Println("Ignoring an event because it is concurrent with an already stored event")
					add_vc = false
				}

			}

			if add_vc {
				dependencies[i].requests = append(dependencies[i].requests, event)
			}

			return dependencies
		}
	}
	return append(dependencies, Dependency{lockset.Clone(), []LockEvent{event}})
}

// Algorithm phase 2

// Based on lock dependencies we can check for cycles.
// A cycle involves n threads and results from some n lock dependencies.
// For each thread we record the requests that might block.

type LockDependency struct {
	thread   ThreadId
	lock     LockId
	lockset  Lockset
	requests []LockEvent
}

type Cycle []LockDependency

func report(s *State, c Cycle) {
	s.cycles = append(s.cycles, c)
}

// After phase 1, the following function yields all cyclice lock depndencies.

// The implementation below follows the algorithm used in UNDEAD (https://github.com/UTSASRG/UnDead/blob/master/analyzer.hh)
func getCycles(s *State) []Cycle {
	if s.cycles != nil {
		return s.cycles
	}
	s.cycles = []Cycle{}

	traversed_thread := make(map[ThreadId]bool)
	for tid := range s.threads {
		traversed_thread[tid] = false
	}

	var chain_stack []LockDependency
	for thread_id := range traversed_thread {
		traversed_thread[thread_id] = true
		for lock, dependencies := range s.threads[thread_id].lock_dependencies {
			for _, dependency := range dependencies {
				chain_stack = append(chain_stack, LockDependency{thread_id, lock, dependency.lockset, dependency.requests}) // push
				dfs(s, &chain_stack, traversed_thread)
				chain_stack = chain_stack[:len(chain_stack)-1] // pop
			}
		}
	}

	return s.cycles
}

func dfs(s *State, chain_stack *[]LockDependency, traversed_thread map[ThreadId]bool) {
	for tid, is_traversed := range traversed_thread {
		if is_traversed {
			continue
		}

		for l, l_d := range s.threads[tid].lock_dependencies {
			for _, l_ls_d := range l_d {
				ld := LockDependency{tid, l, l_ls_d.lockset, l_ls_d.requests}
				if isChain(chain_stack, ld) {
					if isCycleChain(chain_stack, ld) {
						var c Cycle = make([]LockDependency, len(*chain_stack)+1)
						for i, d := range *chain_stack {
							c[i] = d.Clone()
						}
						c[len(*chain_stack)] = ld

						// Check for infeasible deadlocks
						if checkAndFilterConcurrentRequests(&c) {
							report(s, c)
						}
					} else {
						traversed_thread[tid] = true
						*chain_stack = append(*chain_stack, ld) // push
						dfs(s, chain_stack, traversed_thread)
						*chain_stack = (*chain_stack)[:len(*chain_stack)-1] // pop
						traversed_thread[tid] = false
					}
				}
			}
		}
	}
}

// Check if adding dependency to chain will still be a chain.
func isChain(chain_stack *[]LockDependency, dependency LockDependency) bool {

	for _, d := range *chain_stack {
		// Exit early. No two deps can hold the same lock. - Except for read locks
		if d.lock == dependency.lock && dependency.lock.isWrite() {
			logAbortReason("Two dependencies hold the same lock (early exit)")
			return false
		}
		// Check (LD-1) LS(ls_j) cap LS(ls_i+1) for j in {1,..,i}
		// Also (RW-LD-1)
		if !d.lockset.disjointCouldBlock(dependency.lockset) {
			logAbortReason("Locksets are not disjoint (guard)")
			return false
		}
	}

	// Check (LD-2) l_i in ls_i+1
	for l := range dependency.lockset {

		// Also (RW-LD-2)
		if (*chain_stack)[len(*chain_stack)-1].lock.equalsCouldBlock(l) {
			return true
		}

	}
	logAbortReason("Previous lock not in current lockset or both are read locks")
	return false
}

// Check (LD-3) l_n in ls_1
// Also (RW-LD-3)
func isCycleChain(chain_stack *[]LockDependency, dependency LockDependency) bool {
	for l := range (*chain_stack)[0].lockset {
		if l.equalsCouldBlock(dependency.lock) {
			return true
		}
	}
	logAbortReason("Cycle Chain does not close")
	return false
}

// Check if there is one ore more chains of concurrent requests and filter out any requests that are not part of them
func checkAndFilterConcurrentRequests(cycle *Cycle) bool {
	for i := range *cycle {
		// Check if each request has a concurrent request in the element before and after
		// All requests that have a previous request
		requests_with_prev := []LockEvent{}
		for _, req := range (*cycle)[i].requests {
			for _, prev_req := range (*cycle)[(len(*cycle)+i-1)%len(*cycle)].requests {
				if clock.GetHappensBefore(req.vector_clock, prev_req.vector_clock) == clock.Concurrent {
					requests_with_prev = append(requests_with_prev, req)
					break
				}
			}
		}
		// All requests that have a next request
		requests_with_both := []LockEvent{}
		for _, req := range requests_with_prev {
			for _, next_req := range (*cycle)[(i+1)%len(*cycle)].requests {
				if clock.GetHappensBefore(req.vector_clock, next_req.vector_clock) == clock.Concurrent {
					requests_with_both = append(requests_with_both, req)
					break
				}
			}
		}

		if len(requests_with_both) > 0 {
			// Only requests with predecessors and successors remain
			(*cycle)[i].requests = requests_with_both
		} else {
			// An entry with no requests mean that we no longer have a valid cycle
			logAbortReason("Cycle Entry with no concurrent requests")
			return false
		}
	}
	return true
}

// ////////////////////////////////
// High level functions for integration with Advocate
func ResetState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	currentState = State{
		threads: make(map[ThreadId]Thread),
		cycles:  nil,
		failed:  false,
	}
}

func HandleMutexEventForRessourceDeadlock(element TraceElementMutex) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	if currentState.failed {
		return
	}

	event := LockEvent{
		thread_id:    ThreadId(element.GetRoutine()),
		trace_id:     element.GetTID(),
		lock_id:      element.GetID(),
		vector_clock: element.vcWmHB.Copy(),
	}

	switch element.opM {
	case LockOp:
		acquire(&currentState, false, event)
	case TryLockOp:
		// We do not check event.suc because that could led to false negatives
		acquire(&currentState, false, event)
	case RLockOp:
		acquire(&currentState, true, event)
	case UnlockOp:
		release(&currentState, false, event)
	case RUnlockOp:
		release(&currentState, true, event)
	}
}

func CheckForResourceDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)
	if currentState.failed {
		utils.LogError("Failed flag is set, probably encountered unsupported lock operation. No deadlock analysis possible.")
		return
	}
	for i, t := range currentState.threads {
		debugLog("Found", len(t.lock_dependencies), "dependencies in Thread", i)
	}

	getCycles(&currentState)

	debugLog("Found", len(currentState.cycles), "cycles")

	for _, cycle := range currentState.cycles {
		var cycleElements []results.ResultElem
		var request = findEarliesRequest(cycle)

		debugLog("Found cycle with the following entries:", cycle)
		for i := 0; i < len(cycle); i++ {
			debugLog("Entry in routine", cycle[i].thread, ":")
			debugLog("\tLockset:", cycle[i].lockset)
			debugLog("\tAmount of different lock requests that might block it:", len(cycle[i].requests))
			for i, r := range cycle[i].requests {
				debugLog("\t\tLock request", i, ":", r)
			}

			for _, r := range cycle[i].requests {
				if clock.GetHappensBefore(request.vector_clock, r.vector_clock) == clock.Concurrent {
					request = r
					break
				}
			}

			if request.thread_id != cycle[i].thread {
				utils.LogError("Request thread id ", request.thread_id, "does not match entry thread id", cycle[i].thread, ". Ignoring circle!")
				break
			}

			file, line, tPre, err := infoFromTID(request.trace_id)
			if err != nil {
				utils.LogError(err.Error())
				break
			}

			cycleElements = append(cycleElements, results.TraceElementResult{
				RoutineID: int(request.thread_id),
				ObjID:     request.lock_id,
				TPre:      tPre,
				ObjType:   "DC",
				File:      file,
				Line:      line,
			})
		}

		var stuckElement = cycleElements[len(cycleElements)-1].(results.TraceElementResult)
		stuckElement.ObjType = "DH"

		results.Result(results.CRITICAL, results.PCyclicDeadlock, "stuck", []results.ResultElem{stuckElement}, "cycle", cycleElements)
	}
}

/////////////////////////////////
// Auxiliary functions.

// Finds the earliest request in a cycle.
func findEarliesRequest(cycle []LockDependency) LockEvent {
	earliest := cycle[0].requests[0]
	_, _, earliestTime, err := infoFromTID(earliest.trace_id)
	if err != nil {
		utils.LogError(err.Error())
		return earliest
	}
	for _, c := range cycle {
		for _, r := range c.requests {
			_, _, requestTime, err := infoFromTID(r.trace_id)
			if err != nil {
				utils.LogError(err.Error())
				return earliest
			}
			if requestTime < earliestTime {
				earliest = r
				earliestTime = requestTime
			}
		}
	}
	return earliest
}

// Debug logging.

func debugLog(v ...any) {
	log.Println(v...)
}

func logAbortReason(reason ...any) {
	r := []any{"No Deadlock:"}
	r = append(r, reason...)
	utils.LogInfo(r...)
	log.Println(r...)
}

// Lock Depdendency methods.

func (l LockDependency) Clone() LockDependency {
	reqs := make([]LockEvent, len(l.requests))
	for i, r := range l.requests {
		reqs[i] = r.Clone()
	}
	return LockDependency{
		thread:   l.thread,
		lock:     l.lock,
		lockset:  l.lockset.Clone(),
		requests: reqs,
	}
}

// Event methods.

func (e LockEvent) Clone() LockEvent {
	return LockEvent{
		thread_id:    e.thread_id,
		trace_id:     e.trace_id,
		lock_id:      e.lock_id,
		vector_clock: e.vector_clock.Copy(),
	}
}

// Lock methods.

func (l LockId) isRead() bool {
	return l.readLock
}

func (l LockId) isWrite() bool {
	return !l.readLock
}

func (l LockId) addReader(s Thread) {
	s.readerCounter[l]++
}

func (l LockId) removeReader(s Thread) {
	if !l.hasReaders(s) {
		return
	}
	s.readerCounter[l]--
	if s.readerCounter[l] <= 0 {
		delete(s.readerCounter, l)
	}
}

func (l LockId) hasReaders(s Thread) bool {
	if _, exists := s.readerCounter[l]; !exists {
		return false
	}
	return s.readerCounter[l] > 0
}

// Check if two locks are equal ignoring whether they are read or write locks.
func (l LockId) equalsIgnoreRW(other LockId) bool {
	return l.id == other.id
}

// Check if two locks are the same and at least one of them is a write lock.
func (l LockId) equalsCouldBlock(other LockId) bool {
	if !l.equalsIgnoreRW(other) {
		return false
	}
	return l.isWrite() || other.isWrite()
}

// Lockset methods.

func (ls Lockset) empty() bool {
	return len(ls) == 0

}

func (ls Lockset) add(x LockId) {
	ls[x] = struct{}{}
}

func (ls Lockset) remove(x LockId) bool {
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
		b.WriteString(strconv.Itoa(int(l.id)))
	}
	b.WriteString("}")
	return b.String()
}

func (ls Lockset) equal(ls2 Lockset) bool {
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

func (ls Lockset) disjoint(ls2 Lockset) bool {
	for l := range ls {
		if _, contains := ls2[l]; contains {
			return false
		}
	}
	return true
}

func (ls Lockset) disjointCouldBlock(ls2 Lockset) bool {
	for l := range ls {
		for l2 := range ls2 {
			if l.equalsCouldBlock(l2) {
				return false
			}
		}
	}
	return true
}

// Further notes.
//
// If possible we would like to use a double-indexed map of the following form.
//
// type Deps map[Lock]map[Lockset][]Event
//
// Unfortunately, this is not possible in Go because keys must be comparable (but slices, maps, ... are not comparable).
// This is not an issue in Haskell or C++ where we can extend the set of comparable types (but providing additional definitions for "==",...)
//
// Hence, we use single-indexed (by Lock) map.
