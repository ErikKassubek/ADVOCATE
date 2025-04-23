// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_time.go
// Brief: Get the timer for linux
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

//go:build linux

package runtime

// GetAdvocateCounter will update the timer and return the new value
// Return:
//   - new time value
func GetNextTimeStep() int64 {
	return (nanotime() - tracingStartNano)
}

// ADVOCATE-FILE-END
