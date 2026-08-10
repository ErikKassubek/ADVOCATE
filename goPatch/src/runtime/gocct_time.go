// GOCCT-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocct_time.go
// Brief: Get the timer
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var gocctGlobalCounter atomic.Int64

// GetGoCCTCounter will update the timer and return the new value
// Return:
//   - new time value
func GetNextTimeStep() int64 {
	return gocctGlobalCounter.Add(2)
}

// GOCCT-FILE-END
