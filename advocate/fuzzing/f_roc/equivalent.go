// Copyright (c) 2026 Erik Kassubek
//
// File: equivalent.go
// Brief: Check if an roc is considered equivalent to the trace it is based on
//
// Author: Erik Kassubek
// Created: 2026-01-19
//
// License: BSD-3-Clause

package f_roc

import (
	"advocate/analysis/a_hb"
	"advocate/analysis/hb/a_vc"
	"advocate/fuzzing/f_base"
)

// isEquivalent checks if the constraint is sufficiently different from the
// trace it is based on to warred an execution of the constraint
// It is sufficiently different if at least one hb relation is broken
//
//	Parameter:
//	  - cr f_base.Constraint: the constraint to check
//
// Return:
//   - bool: true if the constraint should be skipped, false if it should be executed
func isEquivalent(cr f_base.Constraint) bool {
	for i, elem1 := range cr.Elems {
		for j := i + 1; j < len(cr.Elems); j++ {
			elem2 := cr.Elems[j]
			if a_vc.GetHappensBefore(elem2, elem1, false) == a_hb.Before {
				return false
			}
		}
	}

	return true
}
