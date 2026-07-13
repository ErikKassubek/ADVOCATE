// Copyright (c) 2026 Erik Kassubek
//
// File: dynamic.go
// Brief: The dynamic parts of the blocking detection
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/trace"
)

// getBlockedResources returns the resources of which at least one routine is blocking
func getBlockedResources() map[trace.Resource]struct{} {
	tr := &a_base.MainTrace

	res := make(map[trace.Resource]struct{})

	for _, t := range tr.GetBlocked() {
		ob := t.GetObjId()
		res[trace.NewResource(ob)] = struct{}{}
	}

	return res
}

// activeRoutines returns the active, i.e. not terminated routines and the resources they can access
//
// Parameter:
//   - res []resource: the relevant resources
// func activeRoutines(res map[trace.Resource]struct{}) map[int][]trace.Resource {
// 	tr := &a_base.MainTrace

// 	for rout, _ := range tr.GetTraces() {
// 		lastElem := tr.GetLastElemInRout(rout)

// 		switch lastElem.(type) {
// 		case *trace.ElementRoutineEnd:
// 		default:
// 			continue
// 		}

// 	}

// 	return
// }
