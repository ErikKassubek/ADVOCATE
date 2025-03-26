// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcMutex.go
// Brief: Update functions for vector clocks from mutex operation
//        Some of the functions start analysis functions
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"analyzer/timer"
)

/*
 * Create a new relW and relR if needed
 * Args:
 *   index (int): The id of the atomic variable
 *   nRout (int): The number of routines in the trace
 */
func newRel(index int, nRout int) {
	if _, ok := relW[index]; !ok {
		relW[index] = clock.NewVectorClock(nRout)
	}
	if _, ok := relR[index]; !ok {
		relR[index] = clock.NewVectorClock(nRout)
	}
}

/*
 * Update and calculate the vector clocks given a lock operation
 * Args:
 *   mu (*TraceElementMutex): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 *   wVc (map[int]VectorClock): The current weak vector clocks
 */
func Lock(mu *TraceElementMutex, vc map[int]clock.VectorClock, wVc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if mu.tPost == 0 {
		vc[mu.routine] = vc[mu.routine].Inc(mu.routine)
		return
	}

	newRel(mu.id, vc[mu.routine].GetSize())
	vc[mu.routine] = vc[mu.routine].Sync(relW[mu.id])
	vc[mu.routine] = vc[mu.routine].Sync(relR[mu.id])
	vc[mu.routine] = vc[mu.routine].Inc(mu.routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["leak"] {
		addMostRecentAcquireTotal(mu, vc[mu.routine], 0)
	}

	lockSetAddLock(mu, wVc[mu.routine])

	// for fuzzing
	currentlyHoldLock[mu.id] = mu
	incFuzzingCounter(mu)
}

/*
 * Update and calculate the vector clocks given a unlock operation
 * Args:
 *   mu (*TraceElementMutex): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func Unlock(mu *TraceElementMutex, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if mu.tPost == 0 {
		return
	}

	newRel(mu.id, vc[mu.routine].GetSize())
	relW[mu.id] = vc[mu.routine].Copy()
	relR[mu.id] = vc[mu.routine].Copy()
	vc[mu.routine] = vc[mu.routine].Inc(mu.routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(mu.routine, mu.id)

	// for fuzzing
	currentlyHoldLock[mu.id] = nil
}

/*
 * Update and calculate the vector clocks given a rlock operation
 * Args:
 *   mu (*TraceElementMutex): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 *   wVc (map[int]VectorClock): The current weak vector clocks
 * Returns:
 *   (vectorClock): The new vector clock
 */
func RLock(mu *TraceElementMutex, vc map[int]clock.VectorClock, wVc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if mu.tPost == 0 {
		vc[mu.routine] = vc[mu.routine].Inc(mu.routine)
		return
	}

	newRel(mu.id, vc[mu.routine].GetSize())
	vc[mu.routine] = vc[mu.routine].Sync(relW[mu.id])
	vc[mu.routine] = vc[mu.routine].Inc(mu.routine)

	timer.Stop(timer.AnaHb)

	if analysisCases["leak"] {
		addMostRecentAcquireTotal(mu, vc[mu.routine], 1)
	}

	lockSetAddLock(mu, wVc[mu.routine])

	// for fuzzing
	currentlyHoldLock[mu.id] = mu
	incFuzzingCounter(mu)
}

/*
 * Update and calculate the vector clocks given a runlock operation
 * Args:
 *   mu (*TraceElementMutex): The trace element
 *   vc (map[int]VectorClock): The current vector clocks
 */
func RUnlock(mu *TraceElementMutex, vc map[int]clock.VectorClock) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if mu.tPost == 0 {
		vc[mu.routine] = vc[mu.routine].Inc(mu.routine)
		return
	}

	newRel(mu.id, vc[mu.routine].GetSize())
	relR[mu.id] = relR[mu.id].Sync(vc[mu.routine])
	vc[mu.routine] = vc[mu.routine].Inc(mu.routine)

	timer.Stop(timer.AnaHb)

	lockSetRemoveLock(mu.routine, mu.id)
	// for fuzzing
	currentlyHoldLock[mu.id] = nil
}
