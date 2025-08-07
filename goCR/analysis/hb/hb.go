//
// File: hb.go
// Brief: Happens before enum
//
// Created: 2025-07-08
//
// License: BSD-3-Clause

package hb

// HappensBefore is an enum for the possible hb values
type HappensBefore int

// Possible hb values
const (
	Before = iota
	After
	Concurrent
	None
)
