// Copyright (c) 2026 Erik Kassubek
//
// File: ssa.go
// Brief: Static Single Assigned
// Author: Erik Kassubek
// Created: 2026-05-07
//
// License: BSD-3-Clause

package blockingStatic

import (
	"os"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func (self *staticData) buildSsa() {
	self.ssa, self.ssaPkgs = ssautil.AllPackages(self.pkgs, ssa.SanityCheckFunctions)
	self.ssa.Build()
	for _, p := range self.ssaPkgs {
		p.Build()
	}
	self.ssaMains = ssautil.MainPackages(self.ssaPkgs)
}

// Print the ssa
//
// Parameter:
//   - onlyOne bool: every file/package can be included multiple times in the ssa if it is included with different contexts,
//     e.g. main vs test.
//     If onlyOne is set, print only the first
func (self *staticData) printSSA(onlyOne bool) {
	seen := make(map[string]bool)

	for _, pkg := range self.ssaPkgs {
		if pkg == nil {
			continue
		}

		if onlyOne {

			path := pkg.Pkg.Path()
			if seen[path] {
				continue
			}
			seen[path] = true
		}

		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				println("====", fn.String(), "====")
				fn.WriteTo(os.Stdout)
			}
		}
	}
}
