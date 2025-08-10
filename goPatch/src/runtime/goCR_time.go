// GOCP-FILE_START

// File: goCR_time.go
// Brief: Get the timer
//
// Created: 2024-12-04
//
// License: BSD-3-Clause

package runtime

import "internal/runtime/atomic"

var goCRGlobalCounter atomic.Int64

// GetNextTimeStep will update the timer and return the new value
// Return:
//   - new time value
func GetNextTimeStep() int64 {
	return goCRGlobalCounter.Add(2)
}

// GOCP-FILE-END
