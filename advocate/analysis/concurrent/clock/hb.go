// Copyright (c) 2024 Erik Kassubek
//
// File: hb.go
// Brief: Type for happens before
//
// Author: Erik Kassubek
// Created: 2023-11-30
//
// License: BSD-3-Clause

package clock

import "advocate/analysis/concurrent/hb"

// Check if vc1 happens before vc2
//
// Parameter:
//   - vc1 *VectorClock: The first vector clock
//   - vc2 *VectorClock: The second vector clock
//
// Returns:
//   - bool: True if vc1 is a cause of vc2, false otherwise
func happensBefore(vc1 *VectorClock, vc2 *VectorClock) bool {
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

// GetHappensBefore returns the happens before relation between two operations given there
// vector clocks
//
// Parameter:
//   - vc1 *VectorClock: The first vector clock
//   - vc2 *VectorClock: The second vector clock
//
// Returns:
//   - happensBefore: The happens before relation between the two vector clocks
func GetHappensBefore(vc1 *VectorClock, vc2 *VectorClock) hb.HappensBefore {
	if vc1 == nil || vc2 == nil {
		return hb.None
	}

	if vc1.size != vc2.size {
		return hb.None
	}

	if happensBefore(vc1, vc2) {
		return hb.Before
	}
	if happensBefore(vc2, vc1) {
		return hb.After
	}
	return hb.Concurrent
}

// IsConcurrent returns if the vector clocks are concurrent
// Use this instead of GetHappensBefore(vc1, vc2) == Concurrent
//
// Parameter:
//   - vc1 *Vector clock: first vector clock
//   - vc2 *Vector clock: second vector clock
//
// Returns:
//   - bool: true if vc1 and vc2 are concurrent, false otherwise
func IsConcurrent(vc1 *VectorClock, vc2 *VectorClock) bool {
	if vc1 == nil || vc2 == nil {
		return false
	}

	if vc1.size != vc2.size {
		return false
	}

	hasSmaller := false
	hasBigger := false
	for i := 1; i < vc1.size+1; i++ {
		vc1V := vc1.GetValue(i)
		vc2V := vc2.GetValue(i)
		if vc1V < vc2V {
			if hasBigger {
				return true
			}
			hasSmaller = true
		} else if vc1V > vc2V {
			if hasSmaller {
				return true
			}
			hasBigger = true
		}
	}
	return false
}
