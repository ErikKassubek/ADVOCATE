// Copyright (c) 2026 Erik Kassubek
//
// File: backwardsPass.go
// Brief: Implement Phase 1 out of 3 for the static blocking bug detection
//
// Author: Erik Kassubek
// Created: 2026-07-13
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/trace"
)

// backwards determines the path from the creation of a resource to the end point of a routine
//
// Parameter:
//   - routineID int: routine for the id
//   - resourceID int: routine for the resource
//
// Returns:
//
//	-
func backwards(routineID int, resourceID int) []string {
	res := make([]string, 0)

	rout := a_base.MainTrace.GetRoutineTrace(routineID)

	for i := rout.Len() - 1; i >= 0; i-- {
		op := rout.At(i)

		// Note: the first element in each routine is always the function that is called in this routine.
		// We therefore do not need to record spawns.

		switch o := op.(type) {
		case *trace.ElementAlloc:
			if op.GetObjId() == resourceID {
				return res
			}
		case *trace.ElementFunc:
			res = append([]string{o.GetName()}, res...)
		}
	}

	return res
}
