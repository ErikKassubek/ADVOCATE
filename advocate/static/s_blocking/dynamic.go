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
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
)

// getBlockedResources returns the resources of which at least one routine is blocking
//
// Returns:
//   - map[trace.Resource]*trace.ElementAlloc: slice of map from blocked resource to alloc of resource.
func getBlockedResources() map[trace.Resource]*trace.ElementAlloc {
	res := make(map[trace.Resource]*trace.ElementAlloc)

	for _, e := range a_base.MainTrace.GetBlocked() {
		switch s := e.(type) {
		case *trace.ElementSelect:
			ca := s.GetCases()
			for _, ch := range ca {
				addAlloc(ch, res)
			}
		default:
			addAlloc(e, res)
		}

	}

	return res
}

func addAlloc(e trace.Element, res map[trace.Resource]*trace.ElementAlloc) {
	ob := e.ObjID()

	r := trace.NewResource(ob)

	if _, ok := res[r]; ok {
		return
	}

	alloc := a_base.MainTrace.GetAlloc(e)
	res[r] = alloc
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
