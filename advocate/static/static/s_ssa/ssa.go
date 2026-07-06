// Copyright (c) 2026 Erik Kassubek
//
// File: ssa.go
// Brief: Static Single Assigned
// Author: Erik Kassubek
// Created: 2026-05-07
//
// License: BSD-3-Clause

package s_ssa

import (
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// buildSsa creates the ssa.
// Assumes that packages are already loaded (self.pkgs)
func (self *Data) buildSsa(pkgs []*packages.Package) {
	self.ssa, self.ssaPkgs = ssautil.AllPackages(pkgs, ssa.SanityCheckFunctions)
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
func (self *Data) Print(onlyOne bool) {
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

		var printFn func(*ssa.Function)
		printFn = func(fn *ssa.Function) {
			if fn == nil {
				return
			}

			fmt.Printf("\n============ %s ============\n", fn.String())
			fn.WriteTo(os.Stdout)

			for _, anon := range fn.AnonFuncs {
				printFn(anon)
			}
		}

		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				printFn(fn)
			}
		}
	}
}
