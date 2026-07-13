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

type ft int

const (
	call ft = iota
	spawn
)

type funcId struct {
	pkg      string
	name     string
	funcType ft
}

// backwards determines the path from the creation of a resource to the end point of a routine
//
// Parameter:
//   - routineID int: routine for the id
//   - resourceID int: routine for the resource
//
// Returns:
//
//	-
func backwards(routineID int, resourceID int) []funcId {
	res := make([]funcId, 0)

	tr := a_base.MainTrace.GetRoutineTrace(routineID)

	for i := len(tr) - 1; i >= 0; i-- {
		op := tr[i]

		switch op.(type) {
		case *trace.ElementAlloc:
			if op.GetObjId() == resourceID {
				return res
			}
			// case trace.ElementF
		}
	}

	return res
}
