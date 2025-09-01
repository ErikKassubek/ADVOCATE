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

import "unsafe"

var currentSleeping uintptr

var haveRef = make([]bool, 1000)

func StoreLastPark(w unsafe.Pointer) {
	currentGoRoutineInfo().parkOn = w
}

func ClearLastPark() {
	currentGoRoutineInfo().parkOn = nil
}

func DetectLocalDeadlock() {
	go func() {
		for {
			println("============== START ===================\n")
			var numberWaiting = 0
			var foundRunningReference = false

			routineRunning := make(map[uint64]bool)

			forEachG(func(gp *g) {
				status := readgstatus(gp)
				id := gp.goid

				routineRunning[id] = false
				if status != _Gwaiting {
					routineRunning[id] = true
					return
				}

				// TODO:other valid reasons
				if gp.waitreason != waitReasonSyncMutexLock {
					routineRunning[id] = true
					return
				}

				parkOn := gp.advocateRoutineInfo.parkOn
				if parkOn == nil {
					return
				}

				println("WAITING: ", gp.goid, " ", uintptr(gp.advocateRoutineInfo.parkOn))
				numberWaiting++
			})

			haveRef = make([]bool, 1000)
			
			GC()
			
			for index, ok := range haveRef {
				if ok {
					if routineRunning[uint64(index)] {
						println("RUNNING REFERENCE: ", index)
						foundRunningReference = true
					} else {
						println("SLEEPING REFERENCE: ", index)
					}
				}
			}
			if numberWaiting > 0 && foundRunningReference {
				println("DEADLOCK")
			} else {
				println("NO DEADLOCK")
			}
			println("\n============== START ===================")
			println("\n\n")
			sleep(1)
		}
	}()
}
