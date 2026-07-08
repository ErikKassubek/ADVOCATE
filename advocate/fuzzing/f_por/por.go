// Copyright (c) 2026 Erik Kassubek
//
// File: por.go
// Brief: Entry point for partial order reduction
//
// Author: Erik Kassubek
// Created: 2026-03-16
//
// License: BSD-3-Clause

package f_por

import "advocate/fuzzing/f_base"

// TODO: how to identify the same event over multiple runs?

// Return, if the constraint has a previous, equivalent constraint.
// If not, the constraint is added to alreadyRunROC
//
// Parameter:
//   - constraint f_base.Constraint: the constraint to check
//
// Returns:
//   - bool: true if the is a previous, equivalent constraint, false otherwise
func HasEquivalent(constraint f_base.Constraint) bool {
	for _, constraint2 := range alreadyRunROC {
		if isEquiv := isEquivalent(constraint, constraint2); isEquiv {
			alreadyRunROC = append(alreadyRunROC, constraint)
			return true
		}
	}

	return false
}

// Determine if the two constraints are equivalent regarding por
//
// Parameter:
//   - constraint1 f_base.Constraint: the first roc
//   - constraint2 f_base.Constraint: the second roc
//
// Returns:
//   - bool: true if the constraints are equivalent, false if not
func isEquivalent(constraint1, constraint2 f_base.Constraint) bool {
	ok, c1, c2 := isSubset(constraint1, constraint2)
	if !ok {
		return false
	}
	return canBeReordered(c1, c2)
}

// Determine if one of the constraint is a subset or equal to the other constraint.
//
// Parameter:
//   - constraint1 f_base.Constraint: the first roc
//   - constraint2 f_base.Constraint: the second roc
//
// Returns:
//   - bool: true if one roc is a subset of the other or if the sets (unordered) are equal. False otherwise.
//   - f_base.Constraint: constraint1 with only the events in both constraints
//   - f_base.Constraint: constraint2 with only the events in both constraints
func isSubset(constraint1, constraint2 f_base.Constraint) (bool, f_base.Constraint, f_base.Constraint) {
	// if the constraints are of equal length, constraint1 should be the longer one
	if constraint1.Len() < constraint2.Len() {
		constraint1, constraint2 = constraint2, constraint1
	}

	// remove events from constraint1 not in constraint2
	if constraint1.Len() != constraint2.Len() {
		constraint1 = removeDifferrence(constraint1, constraint2)
	}

	return false, constraint1, constraint2
}

// Assume c1 is longer than c2. Remove all elements from c1 not in c2.
//
// Parameter:
//   - c1 f_base.Constraint: longer constraint
//   - c2 f_base.Constraint: shorter constraint
//
// Returns:
//   - f_base.Constraint: copy of c1 containing only the values tha are in c1
func removeDifferrence(c1, c2 f_base.Constraint) f_base.Constraint {
	// TODO: implement
	return c1
}

// Determine if constraint1 can be reordered to be constraint2 by only swapping neighboring concurrent events.
// The function assumes that the two constraints contain the same events.
//
// Parameter:
//   - constraint1 f_base.Constraint: the first constraint
//   - constraint2 f_base.Constraint: the second constraint
//
// Returns:
//   - bool: true constraint1 can be reordered to be constraint2, false otherwise
func canBeReordered(constraint1, constraint2 f_base.Constraint) bool {
	// TODO: implement
	return false
}

// Determine if two constraints are equal for the por
//
// Parameter:
//   - first f_base.Constraint: the first constraint
//   - second f_base.Constraint: the second constraint
//
// Returns:
//   - bool: true if the events are considered equal, false otherwise
func IsEqualPOR(first, second f_base.Constraint) bool {
	// TODO: implement
	return false
}
