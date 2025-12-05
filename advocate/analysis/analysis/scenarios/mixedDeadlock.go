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
	"advocate/analysis/baseA"
	"advocate/analysis/hb/clock"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/timer"
	"strconv"
)

// LockSetAddLock adds a lock to the lockSet of a routine. It also saves the
// vector clock of the acquire
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
	id := mu.GetObjId()

	if _, ok := baseA.LockSet[routine]; !ok {
		baseA.LockSet[routine] = make(map[int]string)
	}
	if _, ok := baseA.MostRecentAcquire[routine]; !ok {
		baseA.MostRecentAcquire[routine] = make(map[int]baseA.ElemWithVc)
	}

	baseA.LockSet[routine][id] = mu.GetTID()
	baseA.MostRecentAcquire[routine][id] = baseA.ElemWithVc{
		Vc:   vc,
		Elem: mu,
	}
}

// LockSetRemoveLock removes a lock from the lockSet of a routine
//
// Parameter:
//   - routine int: The routine id
//   - lock int: The id of the mutex
func LockSetRemoveLock(routine int, lock int) {
	timer.Start(timer.AnaLeak)
	defer timer.Stop(timer.AnaLeak)
	timer.Start(timer.AnaResource)
	defer timer.Stop(timer.AnaResource)

	if _, ok := baseA.LockSet[routine][lock]; !ok {
		errorMsg := "Lock " + strconv.Itoa(lock) +
			" not in lockSet for routine " + strconv.Itoa(routine)
		log.Error(errorMsg)
		return
	}
	delete(baseA.LockSet[routine], lock)
}

// CheckForMixedDeadlock checks for mixed deadlocks
//
// Parameter:
//   - routineSend int: The routine id of the send operation
//   - routineRevc int: The routine id of the receive operation
func CheckForMixedDeadlock(routineSend int, routineRevc int) {

}
