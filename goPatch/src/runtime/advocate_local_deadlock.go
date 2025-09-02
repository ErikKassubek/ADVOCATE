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

var currentParkedToRoutine = make(map[uintptr][]uint64) // pointer to parked operation -> list of routines parked on operation
var haveRef = make(map[uintptr][]bool)                  // pointer to parked operation -> list of routines with reference to this

func StoreLastPark(w unsafe.Pointer) {
	currentGoRoutineInfo().parkOn = w
}

func ClearLastPark() {
	currentGoRoutineInfo().parkOn = nil
}

func DetectLocalDeadlock() {
	go func() {
		for {
			routineRunning := make(map[uint64]bool) // routine id -> is running
			currentParkedToRoutine = make(map[uintptr][]uint64)

			// search for routines, that are blocked on a concurrency primitive
			numberRoutines := 0
			forEachG(func(gp *g) {
				numberRoutines++
				id := gp.goid

				if !isRoutineWaitingOnConcurrency(gp) {
					routineRunning[id] = true
					return
				}

				routineRunning[id] = false

				if gp.advocateRoutineInfo.parkOn == nil {
					return
				}

				parkOn := uintptr(gp.advocateRoutineInfo.parkOn)

				// store currently sleeping operation
				if _, ok := currentParkedToRoutine[parkOn]; !ok {
					currentParkedToRoutine[parkOn] = make([]uint64, 0)
				}
				currentParkedToRoutine[parkOn] = append(currentParkedToRoutine[parkOn], id)
				println("WAITING: ", id, " ", parkOn)
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
			GC()

			for opId, _ := range currentParkedToRoutine {
				aliveRefs := 0
				for routId, hasRef := range haveRef[opId] {
					if hasRef && routineRunning[uint64(routId)] {
						aliveRefs++
						println("ALIVE REF: ", routId)
					} else if hasRef {
						println("WAITING REF: ", routId)
					}
				}

				if aliveRefs == 0 {
					println("FOUND DEADLOCK")
				}
			}

			println("\n\n")
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
