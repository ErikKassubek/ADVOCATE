// Copyright (c) 2026 Erik Kassubek
//
// File: por.go
// Brief: Entry point for partial order reduction
//
// Author: Erik Kassubek
// Created: 2026-03-16
//
// License: BSD-3-Clause

package por

import "advocate/fuzzing/baseF"

// TODO: how to identify the same event over multiple runs?

// Return, if the constraint has a previous, equivalent constraint.
// If not, the constraint is added to alreadyRunROC
//
// Parameter:
//   - constraint baseF.Constraint: the constraint to check
//
// Returns:
//   - bool: true if the is a previous, equivalent constraint, false otherwise
func HasEquivalent(constraint baseF.Constraint) bool {
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
//   - constraint1 baseF.Constraint: the first roc
//   - constraint2 baseF.Constraint: the second roc
//
// Returns:
//   - bool: true if the constraints are equivalent, false if not
func isEquivalent(constraint1, constraint2 baseF.Constraint) bool {
	ok, c1, c2 := isSubset(constraint1, constraint2)
	if !ok {
		return false
	}
	return canBeReordered(c1, c2)
}

// Determine if one of the constraint is a subset or equal to the other constraint.
//
// Parameter:
//   - constraint1 baseF.Constraint: the first roc
//   - constraint2 baseF.Constraint: the second roc
//
// Returns:
//   - bool: true if one roc is a subset of the other or if the sets (unordered) are equal. False otherwise.
//   - baseF.Constraint: constraint1 with only the events in both constraints
//   - baseF.Constraint: constraint2 with only the events in both constraints
func isSubset(constraint1, constraint2 baseF.Constraint) (bool, baseF.Constraint, baseF.Constraint) {
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
//   - c1 baseF.Constraint: longer constraint
//   - c2 baseF.Constraint: shorter constraint
//
// Returns:
//   - baseF.Constraint: copy of c1 containing only the values tha are in c1
func removeDifferrence(c1, c2 baseF.Constraint) baseF.Constraint {
	// TODO: implement
	return c1
}

// Determine if constraint1 can be reordered to be constraint2 by only swapping neighboring concurrent events.
// The function assumes that the two constraints contain the same events.
//
// Parameter:
//   - constraint1 baseF.Constraint: the first constraint
//   - constraint2 baseF.Constraint: the second constraint
//
// Returns:
//   - bool: true constraint1 can be reordered to be constraint2, false otherwise
func canBeReordered(constraint1, constraint2 baseF.Constraint) bool {
	// TODO: implement
	return false
}

// Determine if two constraints are equal for the por
//
// Parameter:
//   - first baseF.Constraint: the first constraint
//   - second baseF.Constraint: the second constraint
//
// Returns:
//   - bool: true if the events are considered equal, false otherwise
func IsEqualPOR(first, second baseF.Constraint) bool {
	// TODO: implement
	return false
}
