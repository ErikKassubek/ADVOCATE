// Copyright (c) 2026 Erik Kassubek
//
// File: equivalent.go
// Brief: Check if an roc is considered equivalent to the trace it is based on
//
// Author: Erik Kassubek
// Created: 2026-01-19
//
// License: BSD-3-Clause

package roc

import (
	"advocate/analysis/hb"
	"advocate/analysis/hb/vc"
	"advocate/fuzzing/baseF"
)

// isEquivalent checks if the constraint is sufficiently different from the
// trace it is based on to warred an execution of the constraint
// It is sufficiently different if at least one hb relation is broken
//
//	Parameter:
//	  - cr baseF.Constraint: the constraint to check
//
// Return:
//   - bool: true if the constraint should be skipped, false if it should be executed
func isEquivalent(cr baseF.Constraint) bool {
	for i, elem1 := range cr.Elems {
		for j := i + 1; j < len(cr.Elems); j++ {
			elem2 := cr.Elems[j]
			if vc.GetHappensBefore(elem2, elem1, false) == hb.Before {
				return false
			}
		}
	}

	return true
}
