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

var blockedRoutines = make(map[uint64][]uintptr)        // blocked routine id -> blocked operations
var currentParkedToRoutine = make(map[uintptr][]uint64) // pointer to parked operation -> list of routines parked on operation
var haveRef = make(map[uintptr][]bool)                  // pointer to parked operation -> list of routines with reference to this
var routinesByID = make(map[uint64]*g)                  // internal id to g
var alreadyReportedPartialDeadlock = make(map[uintptr]struct{})
var collectPartialDeadlockInfo = false

// StorePark stores in a routine, a pointer to the last concurrency element,
// on which the routine parked
//
// Parameter:
//   - p unsafe.Pointer: pointer to the chan, (rw)mutex, wait group or conditional variable
func StorePark(p unsafe.Pointer) {
	currentGoRoutineInfo().parkOn = []unsafe.Pointer{p}
}

// StorePark stores in a routine, a pointers to the channels involved in a
// select on which a routine parked.
// Do not call if the select has a default.
//
// Parameter:
//   - cas0 *scase: cas0 from the select implementation
//   - order0 *uint16: order0 from the select implementation
//   - ncases int: number of cases in the select (nsends+nrecvs from the select implementation)
func StoreParkSelect(cas0 *scase, order0 *uint16, ncases int) {
	cas1 := (*[1 << 16]scase)(unsafe.Pointer(cas0))

	scases := cas1[:ncases:ncases]

	currentGoRoutineInfo().parkOn = []unsafe.Pointer{}

	for _, scase := range scases {
		currentGoRoutineInfo().parkOn = append(currentGoRoutineInfo().parkOn, unsafe.Pointer(scase.c))
	}
}

func DetectLocalDeadlock() {
	go func() {
		for {
			routinesByID := make(map[uint64]*g)
			routineRunning := make(map[uint64]bool) // routine id -> is running
			currentParkedToRoutine = make(map[uintptr][]uint64)

			// search for routines, that are blocked on a concurrency primitive
			numberRoutines := 0
			forEachG(func(gp *g) {
				numberRoutines++
				id := gp.goid

				routinesByID[id] = gp

				if !isRoutineWaitingOnConcurrency(gp) {
					routineRunning[id] = true
					return
				}

				routineRunning[id] = false

				if gp.advocateRoutineInfo.parkOn == nil {
					return
				}

				for _, p := range gp.advocateRoutineInfo.parkOn {
					parkOn := uintptr(p)
					currentParkedToRoutine[parkOn] = append(currentParkedToRoutine[parkOn], id)
				}
			})

			// initialize haveRef. For each waiting element, we store a list
			// containing one bool variable initialized to false per routine.
			// This is necessary, since we need to count the number of unique
			// routines that hold a reference, while at the same time we should
			// avoid allocating memory while the GC is running (therefore we cannot
			// use a map)
			// We add 10 more places for the case, that between the allocation and
			// running the GC, more routines are created
			haveRef = make(map[uintptr][]bool)
			for obj, _ := range currentParkedToRoutine {
				haveRef[obj] = make([]bool, numberRoutines+10)
			}

			// Run the garbage collector, to find for which sleeping operations, other routines have a reference
			collectPartialDeadlockInfo = true
			GC()
			collectPartialDeadlockInfo = false

			// split all references in waiting and references on runnable/running routines
			// and waiting routines
			aliveRef := make(map[uintptr][]int)
			waitingRef := make(map[uintptr][]int)

			for opID := range currentParkedToRoutine {
				for routID, hasRef := range haveRef[opID] {
					if hasRef && routineRunning[uint64(routID)] {
						aliveRef[opID] = append(aliveRef[opID], routID)
					} else if hasRef {
						waitingRef[opID] = append(waitingRef[opID], routID)
					}
				}
			}

			// check for references without any alive routines
			for opID := range currentParkedToRoutine {
				// no alive references -> deadlock
				if len(aliveRef[opID]) == 0 {
					if _, ok := alreadyReportedPartialDeadlock[opID]; ok {
						continue
					}
					alreadyReportedPartialDeadlock[opID] = struct{}{}

					print("FOUND DEADLOCK")
					for _, ref := range waitingRef[opID] {
						g := routinesByID[uint64(ref)]
						if g.advocateRoutineInfo.id != 0 {
							print("\t", g.advocateRoutineInfo.id, ": ", getWaitingReasonString(g.waitreason))
						} else {
							print("\t", g.goid, ": ", getWaitingReasonString(g.waitreason))
						}
					}
				}
				println("\n")
			}

			sleep(1)
		}
	}()
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
