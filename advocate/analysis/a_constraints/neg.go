// Copyright (c) 2025 Erik Kassubek
//
// File: neg.go
// Brief: Negative constraints
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package a_constraints

import "advocate/trace"

// A negative constraint consisting of two elements
// The positive constraints signals, that the first element does not influences the
// second, e.g. for an atomic operation, the second element does not reads from the first
// or the recv in the second element does not receives from the send in the first element
type negConstraint2 struct {
	first  trace.Element
	second trace.Element
	pos    bool
}

// Return if the constraint is positive.
// Always return false
// Required to implement constraint interface
func (this negConstraint2) isPos() bool {
	return false
}
