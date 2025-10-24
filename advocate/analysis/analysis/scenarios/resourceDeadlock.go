// advocate/analysis/analysis/scenarios/resourceDeadlock.go

// Copyright (c) 2024 Erik Kassubek
//
// File: analysisResourceDeadlock.go
// Brief: Alternative analysis for cyclic mutex deadlocks.
//
// Author: Sebastian Pohsner
// Created: 2025-01-01
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
)

// TODO: fix comments

// Computation of "abstract" lock dependencies
// Lock dependencies are 3-tuples of the following form:
//    (ThreadID, Lock, LockSet)
// Lock dependencies are computed thread-local.
// For each thread there might be several (acquire) events that lead to "lock" acquired under some "lockset".
//
// Each acquire event carries its own vector clock.
// We wish to make use of vector clocks to eliminate infeasible replay candidates.
//
// This means that lock dependencies are 4-tuples of the following form:
//    (ThreadID, Lock, LockSet, []Event)

// ALGORITHM
//
// There are two phases.
//  1. Recording of lock dependencies.
//  2. Checking if lock dependencies imply a cycle.

// Algorithm phase 1

// We show the event processing functions for acquire and release.

func acquire(s *baseA.State, readLock bool, event baseA.LockEvent) {
	if _, exists := s.Threads[event.ThreadID]; !exists {
		s.Threads[event.ThreadID] = baseA.Thread{
			CurrentLockset:   make(baseA.Lockset),
			LockDependencies: make(map[baseA.LockID][]baseA.Dependency),
			ReaderCounter:    make(map[baseA.LockID]int),
		}
	}

	lockID := baseA.LockID{
		ID:       event.LockID,
		ReadLock: readLock,
	}

	ls := s.Threads[event.ThreadID].CurrentLockset
	if !ls.Empty() {
		deps := s.Threads[event.ThreadID].LockDependencies
		deps[lockID] = insert(deps[lockID], ls, event.Clone())
	}

	if lockID.IsRead() {
		lockID.AddReader(s.Threads[event.ThreadID])
	}
	s.Threads[event.ThreadID].CurrentLockset.Add(lockID)
}

func release(s *baseA.State, readLock bool, event baseA.LockEvent) {
	lockID := baseA.LockID{
		ID:       event.LockID,
		ReadLock: readLock,
	}

	if lockID.IsRead() {
		lockID.RemoveReader(s.Threads[event.ThreadID])
		for _, thread := range s.Threads {
			if lockID.HasReaders(thread) {
				continue
			}
			thread.CurrentLockset.Remove(lockID)
		}
		s.Threads[event.ThreadID].CurrentLockset.Remove(lockID)
	} else {
		if !s.Threads[event.ThreadID].CurrentLockset.Remove(lockID) {
			// "Lock not found in lockset! Has probably been released in another thread, this is an unsupported case."
			s.Failed = true
		}
	}
}

// Insert a new lock dependency for a given thread and lock x.
// We assume that event e acquired lock x.
// We might have already an entry that shares the same lock and lockset!
func insert(dependencies []baseA.Dependency, ls baseA.Lockset, event baseA.LockEvent) []baseA.Dependency {
	for i, v := range dependencies {
		if v.Lockset.Equal(ls) {
			dependencies[i].Requests = append(dependencies[i].Requests, event)
			return dependencies
		}
	}
	return append(dependencies, baseA.Dependency{
		Lockset:  ls.Clone(),
		Requests: []baseA.LockEvent{event}},
	)
}

// The above insert function records all requests that share the same dependency (tid,l,ls).
// In case of loops, we may end up with many request entries.
// For performance reasons, we may want to reduce their size.
//
// Eviction strategy.
// Insert variant where we evict event an already stored event f by e,
// if in between f and e no intra-thread synchronization took place.
// This can be checked via helper function equalModuloTID.
// Assumption: Vector clocks underapproximate the must happen-before relation.
// func insert2(dependencies []baseA.Dependency, lockset baseA.Lockset, event baseA.LockEvent) []baseA.Dependency {
// 	// Helper function.
// 	// Assumes that vc1 and vc2 are connected to two events that are from the same thread tid.
// 	// Yields true if vc1[k] == vc2[k] for all threads k but tid.
// 	// Since vc1 and vc2 are underapproximations of the must happen before relation and ignores locks, we also need to check tid itself
// 	equalModuloTID := func(tid baseA.ThreadID, vc1 *clock.VectorClock, vc2 *clock.VectorClock) bool {
// 		if vc1.GetSize() != vc2.GetSize() {
// 			return false
// 		}

// 		for i := 1; i <= vc1.GetSize(); i++ {
// 			// if i == int(tid) {
// 			// 	continue
// 			// }

// 			if vc1.GetValue(i) != vc2.GetValue(i) {
// 				return false
// 			}
// 		}

// 		return true
// 	}

// 	for i, v := range dependencies {
// 		if v.Lockset.Equal(lockset) {
// 			addVc := true

// 			for _, f := range dependencies[i].Requests {
// 				if equalModuloTID(event.ThreadID, event.VectorClock, f.VectorClock) {
// 					// dependencies[i].requests[j] = event // We want to keep the first request for a better replay
// 					fmt.Println("Ignoring an event because it is concurrent with an already stored event")
// 					addVc = false
// 				}

// 			}

// 			if addVc {
// 				dependencies[i].Requests = append(dependencies[i].Requests, event)
// 			}

// 			return dependencies
// 		}
// 	}
// 	return append(dependencies, baseA.Dependency{
// 		Lockset:  lockset.Clone(),
// 		Requests: []baseA.LockEvent{event},
// 	})
// }

// Algorithm phase 2

// Based on lock dependencies we can check for cycles.
// A cycle involves n threads and results from some n lock dependencies.
// For each thread we record the requests that might block.

func report(s *baseA.State, c baseA.Cycle) {
	s.Cycles = append(s.Cycles, c)
}

// After phase 1, the following function yields all cycle lock dependencies.

// The implementation below follows the algorithm used in UNDEAD (https://github.com/UTSASRG/UnDead/blob/master/analyzer.hh)
func getCycles(s *baseA.State) []baseA.Cycle {
	if s.Cycles != nil {
		return s.Cycles
	}
	s.Cycles = []baseA.Cycle{}

	traversedThread := make(map[baseA.ThreadID]bool)
	for tid := range s.Threads {
		traversedThread[tid] = false
	}

	var chainStack []baseA.LockDependency
	for threadID := range traversedThread {
		traversedThread[threadID] = true
		for lock, dependencies := range s.Threads[threadID].LockDependencies {
			for _, dependency := range dependencies {
				chainStack = append(chainStack, baseA.LockDependency{
					Thread:   threadID,
					Lock:     lock,
					Lockset:  dependency.Lockset,
					Requests: dependency.Requests,
				}) // push
				dfs(s, &chainStack, traversedThread)
				chainStack = chainStack[:len(chainStack)-1] // pop
			}
		}
	}

	return s.Cycles
}

func dfs(s *baseA.State, chainStack *[]baseA.LockDependency, traversedThread map[baseA.ThreadID]bool) {
	for tid, isTraversed := range traversedThread {
		if isTraversed {
			continue
		}

		for l, lD := range s.Threads[tid].LockDependencies {
			for _, lLsD := range lD {
				ld := baseA.LockDependency{
					Thread:   tid,
					Lock:     l,
					Lockset:  lLsD.Lockset,
					Requests: lLsD.Requests,
				}
				if isChain(chainStack, ld) {
					if isCycleChain(chainStack, ld) {
						var c baseA.Cycle = make([]baseA.LockDependency, len(*chainStack)+1)
						for i, d := range *chainStack {
							c[i] = d.Clone()
						}
						c[len(*chainStack)] = ld

						// Check for infeasible deadlocks
						if checkAndFilterConcurrentRequests(&c) {
							report(s, c)
						}
					} else {
						traversedThread[tid] = true
						*chainStack = append(*chainStack, ld) // push
						dfs(s, chainStack, traversedThread)
						*chainStack = (*chainStack)[:len(*chainStack)-1] // pop
						traversedThread[tid] = false
					}
				}
			}
		}
	}
}

// Check if adding dependency to chain will still be a chain.
func isChain(chainStack *[]baseA.LockDependency, dependency baseA.LockDependency) bool {

	for _, d := range *chainStack {
		// Exit early. No two deps can hold the same lock. - Except for read locks
		if d.Lock == dependency.Lock && dependency.Lock.IsWrite() {
			// Two dependencies hold the same lock (early exit)
			return false
		}
		// Check (LD-1) LS(ls_j) cap LS(ls_i+1) for j in {1,..,i}
		// Also (RW-LD-1)
		if !d.Lockset.DisjointCouldBlock(dependency.Lockset) {
			// Locksets are not disjoint (guard)
			return false
		}
	}

	// Check (LD-2) l_i in ls_i+1
	for l := range dependency.Lockset {

		// Also (RW-LD-2)
		if (*chainStack)[len(*chainStack)-1].Lock.EqualsCouldBlock(l) {
			return true
		}

	}
	// Previous lock not in current lockset or both are read locks
	return false
}

// Check (LD-3) l_n in ls_1
// Also (RW-LD-3)
func isCycleChain(chainStack *[]baseA.LockDependency, dependency baseA.LockDependency) bool {
	for l := range (*chainStack)[0].Lockset {
		if l.EqualsCouldBlock(dependency.Lock) {
			return true
		}
	}
	// Cycle Chain does not close
	return false
}

// checkAndFilterConcurrentRequests checks if there is one ore more chains of
// concurrent requests and filter out any requests that are not part of them
//
// Parameter:
//   - cycle *baseA.Cycle: a cycle to check
//
// Returns:
//   - bool: true if the cycle is valid regarding hb, false otherwise
func checkAndFilterConcurrentRequests(cycle *baseA.Cycle) bool {
	for i := range *cycle {
		// Check if each request has a concurrent request in the element before and after
		// All requests that have a previous request
		requestsWithPrev := []baseA.LockEvent{}
		for _, req := range (*cycle)[i].Requests {
			for _, prevReq := range (*cycle)[(len(*cycle)+i-1)%len(*cycle)].Requests {
				if clock.GetHappensBefore(req.VectorClock, prevReq.VectorClock) == hb.Concurrent {
					requestsWithPrev = append(requestsWithPrev, req)
					break
				}
			}
		}
		// All requests that have a next request
		requestsWithBoth := []baseA.LockEvent{}
		for _, req := range requestsWithPrev {
			for _, nextReq := range (*cycle)[(i+1)%len(*cycle)].Requests {
				if clock.GetHappensBefore(req.VectorClock, nextReq.VectorClock) == hb.Concurrent {
					requestsWithBoth = append(requestsWithBoth, req)
					break
				}
			}
		}

		if len(requestsWithBoth) > 0 {
			// Only requests with predecessors and successors remain
			(*cycle)[i].Requests = requestsWithBoth
		} else {
			// An entry with no requests mean that we no longer have a valid cycle
			// Cycle Entry with no concurrent requests
			return false
		}
	}
	return true
}

// ResetState resets the current state of the resource deadlock detection
func ResetState() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	baseA.CurrentState = baseA.State{
		Threads: make(map[baseA.ThreadID]baseA.Thread),
		Cycles:  nil,
		Failed:  false,
	}
}

// HandleMutexEventForRessourceDeadlock processes an mutex operation for the
// resource deadlock detection
//
// Parameter:
//   - element trace.ElementMutex: the trace element
func HandleMutexEventForRessourceDeadlock(element trace.ElementMutex) {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	if baseA.CurrentState.Failed {
		return
	}

	event := baseA.LockEvent{
		ThreadID:    baseA.ThreadID(element.GetRoutine()),
		TraceID:     element.GetTID(),
		LockID:      element.GetID(),
		VectorClock: element.GetWVC().Copy(),
	}

	switch element.GetType(true) {
	case trace.MutexLock, trace.MutexTryLock:
		acquire(&baseA.CurrentState, false, event)
		// We do not check event.suc because that could led to false negatives
	case trace.MutexRLock:
		acquire(&baseA.CurrentState, true, event)
	case trace.MutexUnlock:
		release(&baseA.CurrentState, false, event)
	case trace.MutexRUnlock:
		release(&baseA.CurrentState, true, event)
	}
}

// CheckForResourceDeadlock searches for cycles which imply a cyclic resource
// deadlock
func CheckForResourceDeadlock() {
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)
	if baseA.CurrentState.Failed {
		log.Error("Failed flag is set, probably encountered unsupported lock operation. No deadlock analysis possible.")
		return
	}
	// for i, t := range baseA.CurrentState.threads {
	// 	debugLog("Found", len(t.lockDependencies), "dependencies in Thread", i)
	// }

	getCycles(&baseA.CurrentState)

	// debugLog("Found", len(baseA.CurrentState.cycles), "cycles")

	for _, cycle := range baseA.CurrentState.Cycles {
		var cycleElements []results.ResultElem
		var request = findEarliestRequest(cycle)

		// debugLog("Found cycle with the following entries:", cycle)
		for i := 0; i < len(cycle); i++ {
			// 	debugLog("Entry in routine", cycle[i].thread, ":")
			// 	debugLog("\tLockset:", cycle[i].lockset)
			// 	debugLog("\tAmount of different lock requests that might block it:", len(cycle[i].requests))
			// 	for i, r := range cycle[i].requests {
			// 		debugLog("\t\tLock request", i, ":", r)
			// 	}

			for _, r := range cycle[i].Requests {
				if clock.GetHappensBefore(request.VectorClock, r.VectorClock) == hb.Concurrent {
					request = r
					break
				}
			}

			if request.ThreadID != cycle[i].Thread {
				log.Error("Request thread id ", request.ThreadID, "does not match entry thread id", cycle[i].Thread, ". Ignoring circle!")
				break
			}

			file, line, tPre, err := trace.InfoFromTID(request.TraceID)
			if err != nil {
				log.Error(err.Error())
				break
			}

			cycleElements = append(cycleElements, results.TraceElementResult{
				RoutineID: int(request.ThreadID),
				ObjID:     request.LockID,
				TPre:      tPre,
				ObjType:   "DC",
				File:      file,
				Line:      line,
			})
		}

		var stuckElement = cycleElements[len(cycleElements)-1].(results.TraceElementResult)
		stuckElement.ObjType = "DH"

		results.Result(results.CRITICAL, helper.PCyclicDeadlock, "stuck", []results.ResultElem{stuckElement}, "cycle", cycleElements)
	}
}

/////////////////////////////////
// Auxiliary functions.

// Finds the earliest request in a cycle.
func findEarliestRequest(cycle []baseA.LockDependency) baseA.LockEvent {
	earliest := cycle[0].Requests[0]
	_, _, earliestTime, err := trace.InfoFromTID(earliest.TraceID)
	if err != nil {
		log.Error(err.Error())
		return earliest
	}
	for _, c := range cycle {
		for _, r := range c.Requests {
			_, _, requestTime, err := trace.InfoFromTID(r.TraceID)
			if err != nil {
				log.Error(err.Error())
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
