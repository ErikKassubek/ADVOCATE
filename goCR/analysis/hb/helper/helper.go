//
// File: helper.go
// Brief: Helper functions for concurrency used by multiple methods
//
// Created: 2025-07-08
//
// License: BSD-3-Clause

package helper

import "goCR/trace"

// Valid filters out elements which do not correspond to valid operations, e.g.
// end of a routine
//
// Parameter:
//   - elem trace.Element: the element to test
//
// Returns:
//   - bool: true if the element is valid, false otherwise
func Valid(elem trace.Element) bool {
	t := elem.GetObjType(false)
	return !(t == trace.ObjectTypeReplay || t == trace.ObjectTypeNew || t == trace.ObjectTypeRoutineEnd)
}
