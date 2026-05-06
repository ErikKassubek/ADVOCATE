// Copyright (c) 2026 Erik Kassubek
//
// File: callGrahp.go
// Brief: Create and work on the call graph
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blockingStatic

import (
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func (self *staticData) buildSsa() {
	self.ssa, self.ssaPkgs = ssautil.AllPackages(self.pkgs, ssa.SanityCheckFunctions)
	self.ssa.Build()
}

func (self *staticData) buildCallGraph() {
	var roots []*ssa.Function

	for _, pkg := range ssautil.MainPackages(self.ssaPkgs) {
		if fn := pkg.Func("main"); fn != nil {
			roots = append(roots, fn)
		}

		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok && fn.Name() == "init" {
				roots = append(roots, fn)
			}
		}
	}

	result := rta.Analyze(roots, true) // true = build call graph

	self.callGraph = result.CallGraph
}
