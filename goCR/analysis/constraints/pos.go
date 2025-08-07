// Copyright (c) 2025 Erik Kassubek
//
// File: pos.go
// Brief: Positive constraints
//
// Author: Erik Kassubek
// Created: 2025-07-14
//
// License: BSD-3-Clause

package constraints

import "advocate/trace"

// A positive constraint consisting of two elements
// The positive constraints signals, that the first element influences the
// second, e.g. for an atomic operation, the second element reads from the first
// or the recv in the second element receives from the send in the first element
type posConstraint2 struct {
	first  trace.Element
	second trace.Element
}

// Return if the constraint is positive.
// Always return true
// Required to implement constraint interface
func (p posConstraint2) isPos() bool {
	return true
}
