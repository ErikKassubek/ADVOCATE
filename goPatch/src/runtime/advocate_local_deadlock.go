// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_local_deadlock.go
// Brief: Detect local deadlocks while running
//
// Author: Erik Kassubek
// Created: 2025-08-01
//
// License: BSD-3-Clause

package runtime

import (
	"unsafe"
)

type routineStatus int

const (
	unknown routineStatus = iota
	alive
	suspect
	dead
)

var blockedConcurrencyReasons = []waitReason{
	waitReasonChanReceiveNilChan,
	waitReasonChanSendNilChan,
	waitReasonSelect,
	waitReasonSelectNoCases,
	waitReasonChanReceive,
	waitReasonChanSend,
	waitReasonSyncCondWait,
	waitReasonSyncMutexLock,
	waitReasonSyncRWMutexRLock,
	waitReasonSyncRWMutexLock,
	waitReasonSyncWaitGroupWait,
}

var currentParkedToRoutine = make(map[uintptr][]uint64) // pointer to parked operation -> list of routines parked on operation
var parkedOpsPerRoutine = make(map[uint64][]uintptr, 0) // routine -> waiting elements
var haveRef = make(map[uintptr][]bool)                  // pointer to parked operation -> list of routines with reference to this
var routinesByID = make(map[uint64]*g)                  // internal id to g
var alreadyReportedPartialDeadlock = make(map[uint64]struct{})
var collectPartialDeadlockInfo = false
var routineStatusInfo = make(map[uint64]routineStatus)
var routinesWithRef = make(map[uintptr][]uint64)

// StorePark stores in a routine, a pointer to the last concurrency element,
// on which the routine parked
//
// Parameter:
//   - p unsafe.Pointer: pointer to the chan, (rw)mutex, wait group or conditional variable
//   - skip int: caller skip
//   - replay bool: park is forever park due to replay
func StorePark(p unsafe.Pointer, skip int, replay bool) {
	currentGoRoutineInfo().parkOn = []unsafe.Pointer{p}
	currentGoRoutineInfo().parkPos = posFromCaller(skip)
	currentGoRoutineInfo().parkForeverReplay = replay
}

// StorePark stores in a routine, a pointers to the channels involved in a
// select on which a routine parked.
// Do not call if the select has a default.
//
// Parameter:
//   - cas0 *scase: cas0 from the select implementation
//   - order0 *uint16: order0 from the select implementation
//   - ncases int: number of cases in the select (nsends+nrecvs from the select implementation)
//   - skip int: caller skip
func StoreParkSelect(cas0 *scase, order0 *uint16, ncases int, skip int) {
	cas1 := (*[1 << 16]scase)(unsafe.Pointer(cas0))

	scases := cas1[:ncases:ncases]

	currentGoRoutineInfo().parkOn = []unsafe.Pointer{}

	for _, scase := range scases {
		currentGoRoutineInfo().parkOn = append(currentGoRoutineInfo().parkOn, unsafe.Pointer(scase.c))
	}
	currentGoRoutineInfo().parkPos = posFromCaller(skip)
}

// DetectLocalDeadlock checks once per second, if the currently running program
// contains a deadlock. Is this the case it print a corresponding info.
func DetectLocalDeadlock() {
	go func() {
		for {
			currentParkedToRoutine = make(map[uintptr][]uint64)
			parkedOpsPerRoutine = make(map[uint64][]uintptr)
			routinesByID = make(map[uint64]*g)

			// search for routines, that are blocked on a concurrency primitive
			numberRoutines, _ := getWaitingRoutines()

			// initialize haveRef. For each waiting element, we store a list
			// containing one bool variable initialized to false per routine.
			// This is necessary, since we need to count the number of unique
			// routines that hold a reference, while at the same time we should
			// avoid allocating memory while the GC is running (therefore we cannot
			// use a map)
			// We add 10 more places for the case, that between the allocation and
			// running the GC, more routines are created
			haveRef = make(map[uintptr][]bool)
			for obj := range currentParkedToRoutine {
				haveRef[obj] = make([]bool, numberRoutines+10)
			}

			// Run the garbage collector, to find for which sleeping operations, other routines have a reference
			collectPartialDeadlockInfo = true
			GC()
			collectPartialDeadlockInfo = false

			checkForLocalDeadlock()

			sleep(1)
		}
	}()
}

// getWaitingRoutines searches for waiting routines and stores the corresponding
// infos in the corresponding global maps
//
// Returns:
//   - int: total number of running routines
//   - int: number of waiting routines
func getWaitingRoutines() (int, int) {
	numberRoutines := 0
	numberWaitingRoutines := 0
	forEachG(func(gp *g) {
		numberRoutines++
		id := gp.goid

		routinesByID[id] = gp

		if routineStatusInfo[id] != dead {
			routineStatusInfo[id] = alive
		}

		if !isRoutineWaitingOnConcurrency(gp) {
			return
		}

		if gp.advocateRoutineInfo.parkOn == nil {
			return
		}

		numberWaitingRoutines++
		if routineStatusInfo[id] != dead {
			routineStatusInfo[id] = unknown
		}

		for _, p := range gp.advocateRoutineInfo.parkOn {
			parkOn := uintptr(p)
			currentParkedToRoutine[parkOn] = append(currentParkedToRoutine[parkOn], id)
			parkedOpsPerRoutine[id] = append(parkedOpsPerRoutine[id], uintptr(p))
		}
	})

	return numberRoutines, numberWaitingRoutines
}

func checkForLocalDeadlock() {
	routinesWithRef = make(map[uintptr][]uint64)

	for opID := range currentParkedToRoutine {
		for routID, hasRef := range haveRef[opID] {
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

	routinesWithUnknownReferences := make(map[uint64]struct{}, 0)
	// check for references without any alive routines
	for routineID, refIDs := range routineRefs {
		// no other references to any of the blocked elements -> deadlock
		if len(refIDs) == 0 {
			routineStatusInfo[routineID] = dead
			printDeadlockInfo(routineID)
			continue
		}

		// all other routines with references are dead
		foundAliveRef := false
		foundWaitingRef := false

	outer:
		for _, refID := range refIDs {
			switch routineStatusInfo[refID] {
			case alive:
				foundAliveRef = true
				break outer
			case suspect:
				routinesWithUnknownReferences[routineID] = struct{}{}
				foundWaitingRef = true
				break outer
			}
		}

		// alive ref -> no deadlock
		if foundAliveRef {
			continue
		}

		// only dead references -> deadlock
		if !foundWaitingRef {
			printDeadlockInfo(routineID)
		}
	}

	// for routines that are waiting but not dead, check if the waitings can
	// be declared definitely waiting or dead, depending on there references
	oldSize := 0
	for {
		if len(routinesWithUnknownReferences) == oldSize {
			break
		}
		oldSize = len(routinesWithUnknownReferences)

		for routID := range routinesWithUnknownReferences {
			allDead := true
			for _, ref := range routineRefs[routID] {
				// AliveReference and SuspectReference
				if routineStatusInfo[ref] == alive || routineStatusInfo[ref] == suspect {
					routineStatusInfo[routID] = suspect
					delete(routinesWithUnknownReferences, routID)
					allDead = false
					break
				}

				if routineStatusInfo[ref] != dead {
					allDead = false
				}
			}

			// NoReference and DeadReference
			if allDead {
				routineStatusInfo[routID] = dead
				delete(routinesWithUnknownReferences, routID)
			}
		}
	}

	// everything left is deadlock
	for routID := range routinesWithUnknownReferences {
		printDeadlockInfo(routID)
	}
}

func printDeadlockInfo(routineID uint64) {
	for _, opID := range parkedOpsPerRoutine[routineID] {
		if _, ok := alreadyReportedPartialDeadlock[routineID]; ok {
			continue
		}
		alreadyReportedPartialDeadlock[routineID] = struct{}{}

		if g := routinesByID[routineID]; g.advocateRoutineInfo.parkForeverReplay {
			continue
		}

		print("DEADLOCK\n")
		for _, ref := range routinesWithRef[opID] {
			g := routinesByID[uint64(ref)]
			if g.advocateRoutineInfo.id != 0 {
				print("\t", g.advocateRoutineInfo.id, "@", getWaitingReasonString(g.waitreason), "@", g.advocateRoutineInfo.parkPos, "\n")
			} else {
				print("\t", g.goid, ": ", getWaitingReasonString(g.waitreason), "\n")
			}
		}
	}

}

func isRoutineWaitingOnConcurrency(gp *g) bool {
	if readgstatus(gp) != _Gwaiting {
		return false
	}

	if !isInSlice(blockedConcurrencyReasons, gp.waitreason) {
		return false
	}

	return true
}

// getWaitingReasonString takes a waitReason of a routine and returns a
// string representation
//
// Parameter:
//   - wr waitReason: the wait reason enum value
//
// Returns:
//   - string: the string representation of wr
func getWaitingReasonString(wr waitReason) string {
	switch wr {
	case waitReasonChanReceiveNilChan:
		return " chan (recv on nil)"
	case waitReasonChanSendNilChan:
		return "chan (send on nil)"
	case waitReasonSelect:
		return "select"
	case waitReasonSelectNoCases:
		return "select (without cases)"
	case waitReasonChanReceive:
		return "chan (revc)"
	case waitReasonChanSend:
		return "chan (send)"
	case waitReasonSyncCondWait:
		return "cond (wait)"
	case waitReasonSyncMutexLock:
		return "mutex (lock)"
	case waitReasonSyncRWMutexRLock:
		return "rwmutex (rlock)"
	case waitReasonSyncRWMutexLock:
		return "rwmutex (lock)"
	case waitReasonSyncWaitGroupWait:
		return "wait group (wait)"
	}
	return "unknown"
}

// noDeadlockSelect checks for a blocked element, if it is blocked in a select,
// and if so if all cases in the select have no running routines
//
// Parameter:
//   - opID uintptr: the element to check
//
// Returns:
//   - bool: true if the op is in a routine, where another case has channel
//     with a reference in a running routine, false if it is not blocked in
//     a select or if the select has another live reference
// func noDeadlockSelect(opID uintptr) bool {
// 	for _, ref := range waitingRef[opID] {
// 		g := routinesByID[uint64(ref)]

// 		// TODO: this should not happen, but does
// 		if g == nil {
// 			continue
// 		}
// 		if g.waitreason != waitReasonSelect {
// 			continue
// 		}

// 		for _, r := range g.advocateRoutineInfo.parkOn {
// 			if len(aliveRef[uintptr(r)]) > 0 {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }
