// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_partial_deadlock.go
// Brief: Detect partial deadlocks while running
//
// Author: Erik Kassubek
// Created: 2025-08-01
//
// License: BSD-3-Clause

package runtime

var CollectPartialDeadlockInfo = false
var InfoHaveRef map[uintptr][]bool // pointer to parked operation -> list of routines with reference to this

// build oat creates the object aware trace.
// It first determines the blocked concurrency objects.
// It then determines which non-terminated routines have access to those objects
func BuildOAT() {
	blocked, maxRoutId := blockeInfo()

	// from blocked obj memory address to slice.
	// Each elem in slice represents one routine, id -> index
	// If true, routine has access, otherwise it does not
	// Done to prevent alloc during GC
	InfoHaveRef = make(map[uintptr][]bool)
	for b := range blocked {
		InfoHaveRef[b] = make([]bool, maxRoutId+1)
	}

	CollectPartialDeadlockInfo = true
	GC()
	CollectPartialDeadlockInfo = false

	// invert InfoHaveRef, rout id -> []object id
	ref := make(map[uint64][]uint64)
	for obj, routInfo := range InfoHaveRef {
		for rout, ok := range routInfo {
			if !ok {
				continue
			}

			r := uint64(rout)

			if _, ok := ref[r]; !ok {
				ref[r] = make([]uint64, 0)
			}

			ref[r] = append(ref[r], blocked[obj].id)
		}
	}

	ForEachAdvocateG(func(adGp *AdvocateG) {
		if adGp.isNil() {
			return
		}

		id := adGp.id()

		if obj, ok := ref[id]; ok {
			adGp.setOAT(obj)
		}
	})
}

// blockeInfo determines all blocked objects
// It additionally determines the max routine id
//
// Returns:
//   - map[uintptr]park: blocked objects, from object memory addres to object info
//   - uint64: max rout id
func blockeInfo() (map[uintptr]park, uint64) {
	blocked := make(map[uintptr]park, 0)
	var maxID uint64 = 0

	ForEachAdvocateG(func(gp *AdvocateG) {
		if gp.isNil() {
			return
		}

		if gp.isRoutineWaitingOnConcurrency() {
			for _, parkObj := range gp.parkInfo() {
				blocked[uintptr(parkObj.addr)] = parkObj
			}
		}

		maxID = max(maxID, gp.id())
	})

	return blocked, maxID
}
