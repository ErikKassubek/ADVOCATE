// Copyright (c) 2024 Erik Kassubek
//
// File: analysisMixedDeadlock.go
// Brief: Trace analysis for mixed deadlocks. Currently not used.
//
// Author: Erik Kassubek
// Created: 2024-01-27
//
// License: BSD-3-Clause

package scenarios

import (
	"advocate/analysis/data"
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
	"strconv"
)

// Add a lock to the lockSet of a routine. Also save the vector clock of the acquire
//
// Parameter:
//   - routine int: The routine id
//   - lock int: The id of the mutex
//   - tId string: The trace id of the mutex operation
//   - vc VectorClock: The current vector clock
func LockSetAddLock(mu *trace.ElementMutex, vc *clock.VectorClock) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	routine := mu.GetRoutine()
	id := mu.GetID()

	if _, ok := data.LockSet[routine]; !ok {
		data.LockSet[routine] = make(map[int]string)
	}
	if _, ok := data.MostRecentAcquire[routine]; !ok {
		data.MostRecentAcquire[routine] = make(map[int]data.ElemWithVc)
	}

	// if _, ok := data.LockSet[routine][lock]; ok {
	// TODO: TODO: add a result. Deadlock detection is currently disabled
	// errorMsg := "Lock " + strconv.Itoa(lock) +
	// 	" already in lockSet for routine " + strconv.Itoa(routine)
	// results.Debug(errorMsg, results.ERROR)

	// // this is a double locking
	// found := "Double locking:\n"
	// found += "\tlock1: " + posOld + "\n"
	// found += "\tlock2: " + tID
	// results.Result(found, results.CRITICAL)
	// }

	data.LockSet[routine][id] = mu.GetTID()
	data.MostRecentAcquire[routine][id] = data.ElemWithVc{
		Vc:   vc,
		Elem: mu,
	}
}

// Remove a lock from the lockSet of a routine
//
// Parameter:
//   - routine int: The routine id
//   - lock int: The id of the mutex
func LockSetRemoveLock(routine int, lock int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	if _, ok := data.LockSet[routine][lock]; !ok {
		errorMsg := "Lock " + strconv.Itoa(lock) +
			" not in lockSet for routine " + strconv.Itoa(routine)
		log.Error(errorMsg)
		return
	}
	delete(data.LockSet[routine], lock)
}

// Check for mixed deadlocks
//
// Parameter:
//   - routineSend int: The routine id of the send operation
//   - routineRevc int: The routine id of the receive operation
func CheckForMixedDeadlock(routineSend int, routineRevc int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)

	for m := range data.LockSet[routineSend] {
		_, ok1 := data.MostRecentAcquire[routineRevc][m]
		_, ok2 := data.MostRecentAcquire[routineSend][m]
		if ok1 && ok2 && data.MostRecentAcquire[routineSend][m].Elem.GetTID() != data.MostRecentAcquire[routineRevc][m].Elem.GetTID() {
			// found possible mixed deadlock
			// TODO: add a result. Deadlock detection is currently disabled
			// found := "Possible mixed deadlock:\n"
			// found += "\tlocks: \t\t" + data.MostRecentAcquire[routineSend][m].TID + "\t\t" + data.MostRecentAcquire[routineRevc][m].TID + "\n"
			// found += "\tsend/close-recv: \t\t" + tIDSend + "\t\t" + tIDRecv

			// results.Result(found, results.CRITICAL)
		}
	}

	for m := range data.LockSet[routineRevc] {
		_, ok1 := data.MostRecentAcquire[routineRevc][m]
		_, ok2 := data.MostRecentAcquire[routineSend][m]
		if ok1 && ok2 && data.MostRecentAcquire[routineSend][m].Elem.GetTID() != data.MostRecentAcquire[routineRevc][m].Elem.GetTID() {
			// found possible mixed deadlock
			// TODO: add a result. Deadlock detection is currently disabled
			// found := "Possible mixed deadlock:\n"
			// found += "\tlocks: \t\t" + data.MostRecentAcquire[routineSend][m].TID + "\t\t" + data.MostRecentAcquire[routineRevc][m].TID + "\n"
			// found += "\tsend/close-recv: \t\t" + tIDSend + "\t\t" + tIDRecv

			// results.Result(found, results.CRITICAL)
		}
	}
}

// func CheckForMixedDeadlock2(routine int) {
// 	for m := range data.LockSet[routine] {
// 		// if the lock was not acquired by the routine, continue. Should not happen
// 		vc1, okS := data.MostRecentAcquire[routine][m]
// 		if !okS {
// 			continue
// 		}

// 		for routine2, acquire := range data.MostRecentAcquire {
// 			if routine == routine2 {
// 				continue
// 			}

// 			if vc2, ok := acquire[m]; ok {
// 				weakHappensBefore := clock.GetHappensBefore(vc1, vc2)
// 				if weakHappensBefore != Concurrent {
// 					continue
// 				}

// 				// found possible mixed deadlock
// 				found := "Possible mixed deadlock:\n"
// 				found += "\tlock1: " + data.LockSet[routine][m] + "\n"
// 				found += "\tlock2: " + data.LockSet[routine2][m]

// 				results.Result(found, results.CRITICAL)
// 			}

// 		}
// 	}
// }
