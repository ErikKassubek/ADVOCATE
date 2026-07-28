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
//   - []trace.Resource: slice of blocked resource.
func getBlockedResources() map[*trace.ElementAlloc]*trace.Resource {
	res := make(map[*trace.ElementAlloc]*trace.Resource)

	// TODO: r.Alloc not correct if copy of function parameter
	for _, e := range a_base.MainTrace.GetBlocked() {
		for _, r := range a_base.MainTrace.GetResources(e) {
			res[r.Alloc()] = r
		}
	}

	return res
}

func buildFuncCallToSSAFunc() {
	funcCallToSSAFunc = make(map[*trace.ElementFunc]*s_ssa.Function)

	for f := range a_base.MainTrace.CallTree().GetTree() {
		fn := getSSAFuncFromName(f.GetSSAName())
		if fn == nil {
			log.Errorf("Could not find ssa function for %s", f.GetSSAName())
			continue
		}
		funcCallToSSAFunc[f] = fn
	}
}
