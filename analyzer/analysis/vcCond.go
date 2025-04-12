// Copyright (c) 2024 Erik Kassubek
//
// File: vcCond.go
// Brief: Update functions for vector clocks from conditional variables operations
//
// Author: Erik Kassubek
// Created: 2024-01-09
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/timer"
)

// Update and calculate the vector clocks given a wait operation
//
// Parameter:
//   - co (*TraceElementCond): The trace element
func CondWait(co *TraceElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if co.tPost != 0 { // not leak
		if _, ok := currentlyWaiting[co.id]; !ok {
			currentlyWaiting[co.id] = make([]int, 0)
		}
		currentlyWaiting[co.id] = append(currentlyWaiting[co.id], co.routine)
	}
	currentVC[co.routine].Inc(co.routine)
	currentWVC[co.routine].Inc(co.routine)
}

// Update and calculate the vector clocks given a signal operation
//
// Parameter:
//   - co (*TraceElementCond): The trace element
func CondSignal(co *TraceElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	if len(currentlyWaiting[co.id]) != 0 {
		tWait := currentlyWaiting[co.id][0]
		currentlyWaiting[co.id] = currentlyWaiting[co.id][1:]
		currentVC[tWait].Sync(currentVC[co.routine])
	}

	currentVC[co.routine].Inc(co.routine)
	currentWVC[co.routine].Inc(co.routine)
}

// Update and calculate the vector clocks given a broadcast operation
//
// Parameter:
//   - co (*TraceElementCond): The trace element
func CondBroadcast(co *TraceElementCond) {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	for _, wait := range currentlyWaiting[co.id] {
		currentVC[wait].Sync(currentVC[co.routine])
	}
	currentlyWaiting[co.id] = make([]int, 0)

	currentVC[co.routine].Inc(co.routine)
	currentWVC[co.routine].Inc(co.routine)
}
