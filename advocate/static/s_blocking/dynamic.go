// Copyright (c) 2026 Erik Kassubek
//
// File: dynamic.go
// Brief: The dynamic parts of the blocking detection
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
)

// getBlockedResources returns the resources of which at least one routine is blocking
//
// Returns:
//   - map[int]trace.Resource: blocked object id to blocked resource.
func getBlockedResources() map[int]*trace.Resource {
	res := make(map[int]*trace.Resource)

	for _, e := range a_base.MainTrace.GetBlocked() {
		log.Debug("Blocked: ", e)
		resources := a_base.MainTrace.GetResources(e)
		for _, r := range resources {
			res[r.Alloc().ObjID()] = r
		}
	}

	return res
}

func buildFuncCallToSSAFunc() {
	blocking.FuncCallToSSAFunc = make(map[*trace.ElementFunc]*s_ssa.Function)

	for f := range a_base.MainTrace.CallTree().GetTree() {
		fn := s_ssa.GetSSAFuncFromName(data.Ssa(), f.GetSSAName())
		if fn == nil {
			log.Errorf("Could not find ssa function for %s", f.GetSSAName())
			continue
		}
		blocking.FuncCallToSSAFunc[f] = fn
	}
}
