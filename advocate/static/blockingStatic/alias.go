// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Determine if two variables in the ast correspond to the same object
//
// Author: Erik Kassubek
// Created: 2026-05-07
//
// License: BSD-3-Clause

package blockingStatic

import (
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type AliasResult int

const (
	never AliasResult = iota
	sometimes
	always
)

func (self *AliasResult) String() string {
	switch *self {
	case never:
		return "never"
	case sometimes:
		return "sometimes"
	case always:
		return "always"
	default:
		return "unknown"
	}
}

func (self *staticData) buildSsa() {
	self.ssa, self.ssaPkgs = ssautil.AllPackages(self.pkgs, ssa.SanityCheckFunctions)
	self.ssa.Build()
	for _, p := range self.ssaPkgs {
		p.Build()
	}
}

func (self *staticData) runPointerAnalysis() {
	// TODO: implement
}

func (self *staticData) runAliasAnalysis() {
	// TODO: implement
}
