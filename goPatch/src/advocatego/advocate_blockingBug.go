// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Entry point for partial deadlock detection
//
// Author: Erik Kassubek
// Created: 2025-07-23
//
// License: BSD-3-Clause

package advocatego

import (
	"runtime"
	"strconv"
	"strings"
)

type routineStatus int

const (
	unknown routineStatus = iota
	alive
	suspect
	dead
	waiting
)

var partialDeadlocks = make([]string, 0)

var currentParkedToRoutine = make(map[uintptr][]uint64) // poruntimeinter to parked operation -> list of routines parked onreadgstatus(gp) != _Gwaiting operation
var parkedOpsPerRoutine = make(map[uint64][]uintptr, 0) // routine -> waiting elements
var routinesByID = make(map[uint64]*runtime.AdvocateG)  // internal id to g
var alreadyReportedPartialDeadlock = make(map[uint64]struct{})
var routineStatusInfo = make(map[uint64]routineStatus)
var routinesWithRef = make(map[uintptr][]uint64)
var AdvocatePDDetectionStopped = false

// DetectBlockingGC runs a partial deadlock detection in the current execution
// Parameter:
//   - interval bool: interval to run the detector. Set to 0 to run only once
func DetectBlockingGC() int {
	AdvocatePDDetectionStopped = false

	if AdvocatePDDetectionStopped {
		return 0
	}

	println("Request")
	req := AdvocateRequest("Check blocking bug")
	println("RES:", req)

	res := AdvocateDetectBlocking()
	containsChan := false
	for _, r := range res {
		fields := strings.Split(r, "@")
		if len(fields) == 0 {
			continue
		}
		elems := strings.Split(fields[len(fields)-1], ":")
		if len(elems) == 0 {
			continue
		}
		if elems[0] == "chan" {
			containsChan = true
			break
		}
	}
	if len(res) != 0 {
		partialDeadlocks = append(partialDeadlocks, res...)
		if containsChan {
			return runtime.ExitCodeMixedDeadlock
		}
		return runtime.ExitCodeCyclic
	}

	return 0
}

// StopPartialDeadlockDetection stops the partial deadlock detection before
// the start of a new run
// If the detection is currently not run in a loop, this has no effect
func StopPartialDeadlockDetection() {
	AdvocatePDDetectionStopped = true
}

// detectPD checks, if the currently running program
// contains a deadlock. Is this the case it print a corresponding info.
func AdvocateDetectBlocking() []string {
	currentParkedToRoutine = make(map[uintptr][]uint64)
	parkedOpsPerRoutine = make(map[uint64][]uintptr)
	routinesByID = make(map[uint64]*runtime.AdvocateG)

	// search for routines, that are blocked on a concurrency primitive
	_, _, maxID := getWaitingRoutines()

	// initialize haveRef. For each waiting element, we store a list
	// containing one bool variable initialized to false per routine.
	// This is necessary, since we need to count the number of unique
	// routines that hold a reference, while at the same time we should
	// avoid allocating memory while the GC is running (therefore we cannot
	// use a map)
	// We add 10 more places for the case, that between the allocation and
	// running the GC, more routines are created
	runtime.DeadlockInfoHaveRef = make(map[uintptr][]bool)
	for obj := range currentParkedToRoutine {
		runtime.DeadlockInfoHaveRef[obj] = make([]bool, maxID+10)
	}

	// Run the garbage collector, to find for which sleeping operations, other routines have a reference
	runtime.CollectPartialDeadlockInfo = true
	runtime.GC()
	runtime.CollectPartialDeadlockInfo = false

	// for obj, routine := range haveRef {
	// 	print(obj, " -> ")
	// 	for i, rout := range routine {
	// 		if rout {
	// 			print(i+1, " ")
	// 		}
	// 	}
	// 	print("\n")
	// }

	return checkForBlocked()
}

// getWaitingRoutines searches for waiting routines and stores the corresponding
// infos in the corresponding global maps
//
// Returns:
//   - int: total number of running routines
//   - int: number of waiting routines
//   - uint64: maximum ID
func getWaitingRoutines() (int, int, uint64) {
	numberRoutines := 0
	numberWaitingRoutines := 0
	var maxID uint64 = 0
	runtime.ForEachAdvocateG(func(gp *runtime.AdvocateG) {
		numberRoutines++
		id := gp.GoId()

		if id > maxID {
			maxID = id
		}

		routinesByID[id] = gp

		if routineStatusInfo[id] != dead {
			routineStatusInfo[id] = alive
		}

		if !isRoutineWaitingOnConcurrency(gp) {
			return
		}

		if gp.ParkOn() == nil {
			return
		}

		numberWaitingRoutines++
		if routineStatusInfo[id] != dead {
			routineStatusInfo[id] = waiting
		}

		for _, p := range gp.ParkOn() {
			parkOn := uintptr(p)
			currentParkedToRoutine[parkOn] = append(currentParkedToRoutine[parkOn], id)
			parkedOpsPerRoutine[id] = append(parkedOpsPerRoutine[id], uintptr(p))
		}
	})

	return numberRoutines, numberWaitingRoutines, maxID
}

func checkForBlocked() []string {
	routinesWithRef = make(map[uintptr][]uint64)

	for opID := range currentParkedToRoutine {
		for routID, hasRef := range runtime.DeadlockInfoHaveRef[opID] {
			if hasRef {
				routinesWithRef[opID] = append(routinesWithRef[opID], uint64(routID))
			}
		}
	}

	routineRefs := make(map[uint64][]uint64) // for each blocke routine, the routines that have a reference
	for routineID, opIDs := range parkedOpsPerRoutine {
		for _, opID := range opIDs {
			for _, ref := range routinesWithRef[opID] {
				if routineID == ref {
					continue
				}
				routineRefs[routineID] = append(routineRefs[routineID], ref)
			}
		}
	}

	for {
		couldApplyRule := false
		for rID, status := range routineStatusInfo {
			// not in waiting'
			if status != waiting {
				continue
			}

			// NoReference
			if len(routineRefs[rID]) == 0 {
				routineStatusInfo[rID] = dead
				couldApplyRule = true
				continue
			}

			allRefDead := true
			for _, ref := range routineRefs[rID] {
				refStatus := routineStatusInfo[ref]

				// DeadReference
				if refStatus != dead {
					allRefDead = false
				}

				// NonDeadReference
				if refStatus == alive || refStatus == suspect {
					routineStatusInfo[rID] = suspect
					couldApplyRule = true
					continue
				}
			}

			// DeadReference
			if allRefDead {
				routineStatusInfo[rID] = dead
				couldApplyRule = true
			}
		}

		if !couldApplyRule {
			break
		}
	}

	// NoOtherRule
	for rID, status := range routineStatusInfo {
		if status == waiting {
			routineStatusInfo[rID] = dead
		}
	}

	// Check for cyclic deadlock
	deadlock := checkCyclic()

	// Report dead routines
	foundDeadlocks := make([]string, 0)
	for rID, status := range routineStatusInfo {
		if status == dead {
			_, dl := deadlock[rID]
			res := reportDeadlock(rID, dl)
			if res != "" {
				foundDeadlocks = append(foundDeadlocks, res)
			}
		}
	}

	return foundDeadlocks
}

// check for cyclic dependencies
func checkCyclic() map[uint64]struct{} {
	deadRoutines := map[uint64]struct{}{}
	for rID, status := range routineStatusInfo {
		if status == dead {
			deadRoutines[rID] = struct{}{}
		}
	}

	graph := map[uint64][]uint64{}
	selfLoop := map[uint64]bool{}

	for rID := range deadRoutines {
		for _, e := range parkedOpsPerRoutine[rID] {
			for _, rID2 := range routinesWithRef[e] {
				if _, ok := deadRoutines[rID2]; ok {
					graph[rID] = append(graph[rID], rID2)
					if rID == rID2 {
						selfLoop[rID] = true
					}
				}
			}
		}
	}

	// Tarjan SCC
	index := 0
	stack := []uint64{}
	onStack := map[uint64]bool{}
	indices := map[uint64]int{}
	lowlink := map[uint64]int{}

	result := map[uint64]struct{}{}

	var strongConnect func(uint64)
	strongConnect = func(v uint64) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range graph[v] {
			if _, seen := indices[w]; !seen {
				strongConnect(w)
				lowlink[v] = min(lowlink[v], lowlink[w])
			} else if onStack[w] {
				lowlink[v] = min(lowlink[v], indices[w])
			}
		}

		// Root of SCC
		if lowlink[v] == indices[v] {
			scc := map[uint64]struct{}{}
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc[w] = struct{}{}
				if w == v {
					break
				}
			}

			// must be cycle with multiple elements
			if len(scc) == 1 {
				return
			}

			// must be closed
			for r := range scc {
				for _, e := range parkedOpsPerRoutine[r] {
					for _, r2 := range routinesWithRef[e] {
						if _, ok := scc[r2]; !ok {
							return
						}
					}
				}
			}

			for r := range scc {
				result[r] = struct{}{}
			}
		}
	}

	for r := range deadRoutines {
		if _, seen := indices[r]; !seen {
			strongConnect(r)
		}
	}

	return result
}

func reportDeadlock(routineID uint64, deadlock bool) string {
	if _, ok := alreadyReportedPartialDeadlock[routineID]; ok {
		return ""
	}
	alreadyReportedPartialDeadlock[routineID] = struct{}{}

	g := routinesByID[routineID]

	if g.ParkForeverReplay() {
		return ""
	}

	if runtime.AdvocateIgnore(g.ParkPos()) {
		return ""
	}

	res := ""

	header := "LEAK_GC"
	if deadlock {
		header = "DEADLOCK_GC"
	}

	if g.ParkPos() == "" {
		g.SetParkPos("-")
	}
	if g.Id() != 0 {
		res = header + "@" + strconv.FormatUint(g.Id(), 10) + "@" + g.ParkPos() + "@" + runtime.GetWaitingReasonString(g.GetWaitReason())
		print(res, "\n")
	} else {
		res = header + "@" + strconv.FormatUint(g.GoId(), 10) + "@" + g.ParkPos() + "@" + runtime.GetWaitingReasonString(g.GetWaitReason())
		print(res, "\n")
	}

	return res
}

func isRoutineWaitingOnConcurrency(gp *runtime.AdvocateG) bool {
	if !runtime.ReadyStatusWaiting(gp) {
		return false
	}

	if !runtime.IsInSlice(runtime.BlockedConcurrencyReasons, gp.GetWaitReason()) {
		return false
	}

	return true
}
