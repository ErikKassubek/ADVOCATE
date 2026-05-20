// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_time.go
// Brief: Get the timer
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var gocdrGlobalCounter atomic.Int64

// GetGocdrCounter will update the timer and return the new value
// Return:
//   - new time value
func GetNextTimeStep() int64 {
	return gocdrGlobalCounter.Add(2)
}

// GOCDR-FILE-END
