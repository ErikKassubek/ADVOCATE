// Copyrigth (c) 2024 Erik Kassubek
//
// File: vcCond.go
// Brief: Update functions for vector clocks from conditional variables operations 
// 
// Author: Erik Kassubek <kassubek.erik@gmail.com>
// Created: 2024-01-09
// LastChange: 2024-09-01
//
// License: BSD-3-Clause

package analysis

import "analyzer/clock"

var lastCondRelease = make(map[int]int) // -> id -> routine

/*
 * Update and calculate the vector clocks given a wait operation
 * Args:
 *   id (int): The id of the condition variable
 *   routine (int): The routine id
 *   vc (map[int]VectorClock): The current vector clocks
 *   leak (bool): If the operation is a leak (tPost = 0)
 */
func CondWait(id int, routine int, vc map[int]clock.VectorClock, leak bool) {
	if !leak {
		vc[routine].Sync(vc[lastCondRelease[id]])
	}
	vc[routine].Inc(routine)
}

/*
 * Update and calculate the vector clocks given a signal operation
 * Args:
 *   id (int): The id of the condition variable
 *   routine (int): The routine id
 *   vc (map[int]VectorClock): The current vector clocks
 */
func CondSignal(id int, routine int, vc map[int]clock.VectorClock) {
	vc[routine].Inc(routine)

	lastCondRelease[id] = routine
}

/*
 * Update and calculate the vector clocks given a broadcast operation
 * Args:
 *   id (int): The id of the condition variable
 *   routine (int): The routine id
 *   vc (map[int]VectorClock): The current vector clocks
 */
func CondBroadcast(id int, routine int, vc map[int]clock.VectorClock) {
	vc[routine].Inc(routine)
	lastCondRelease[id] = routine
}
