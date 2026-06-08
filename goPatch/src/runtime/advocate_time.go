// OOSC-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: oosc_time.go
// Brief: Get the timer
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var ooscGlobalCounter atomic.Int64

// GetOoscCounter will update the timer and return the new value
// Return:
//   - new time value
func GetNextTimeStep() int64 {
	return ooscGlobalCounter.Add(2)
}

// OOSC-FILE-END
