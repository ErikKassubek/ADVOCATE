// ADVOCATE-FILE_START

// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_replay_data.go
// Brief: Data for constraint execution
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

var (
	constratins Constraint
	waiting     waitingOps

	spawns  = make(map[int][]int)
	selects = make(map[string][]int)

	replayEnabled = false
	selectEnabled = false
)

var (
	waitDeadlockDetect     bool
	waitDeadlockDetectLock mutex

	mgcRoutine uint64 = 0
)

var printDebug = true

func SetSpawns(sp map[int][]int) {
	println("SET SPAWN")
	for i, j := range sp {
		print(i, " -> ")
		for _, e := range j {
			print(e, ", ")
		}
		println("")
	}
	println("SPAWN SET")
	spawns = sp
}
