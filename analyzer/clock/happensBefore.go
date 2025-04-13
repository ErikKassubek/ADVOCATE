// Copyright (c) 2024 Erik Kassubek
//
// File: happensBefore.go
// Brief: Type for happens before
//
// Author: Erik Kassubek
// Created: 2023-11-30
//
// License: BSD-3-Clause

package clock

// HappensBefore provides an enum for possible happens before relations
type HappensBefore int

const (
	Before HappensBefore = iota
	After
	Concurrent
	None
)

// Check if vc1 is a cause of vc2
//
// Parameter:
//   - vc1 *VectorClock: The first vector clock
//   - vc2 *VectorClock: The second vector clock
//
// Returns:
//   - bool: True if vc1 is a cause of vc2, false otherwise
func isCause(vc1 *VectorClock, vc2 *VectorClock) bool {
	atLeastOneSmaller := false
	for i := 1; i <= vc1.size; i++ {
		if vc1.GetValue(i) > vc2.GetValue(i) {
			return false
		} else if vc1.GetValue(i) < vc2.GetValue(i) {
			atLeastOneSmaller = true
		}
	}
	return atLeastOneSmaller
}

// Get the happens before relation between two operations given there
// vector clocks
//
// Parameter:
//   - vc1 *VectorClock: The first vector clock
//   - vc2 *VectorClock: The second vector clock
//
// Returns:
//   - happensBefore: The happens before relation between the two vector clocks
func GetHappensBefore(vc1 *VectorClock, vc2 *VectorClock) HappensBefore {
	if vc1.size != vc2.size {
		return None
	}

	if isCause(vc1, vc2) {
		return Before
	}
	if isCause(vc2, vc1) {
		return After
	}
	return Concurrent
}
