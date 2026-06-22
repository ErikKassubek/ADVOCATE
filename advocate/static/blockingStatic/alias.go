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
	"fmt"
	"go/types"

	"golang.org/x/tools/go/pointer"
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
	self.ssaMains = ssautil.MainPackages(self.ssaPkgs)
}

// TODO: internal panic T
func (self *staticData) runPointerAnalysis() error {
	return nil // TODO: remove if function is fixed
	config := &pointer.Config{
		Mains: self.ssaMains,
	}

	config.Queries = make(map[ssa.Value]struct{})

	for fn := range ssautil.AllFunctions(self.ssa) {
		if fn == nil {
			continue
		}

		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {

				v, ok := instr.(ssa.Value)
				if !ok {
					continue
				}

				if _, ok := v.Type().Underlying().(*types.Pointer); ok {
					config.Queries[v] = struct{}{}
				}
			}
		}
	}

	result, err := pointer.Analyze(config)
	if err != nil {
		return err
	}

	// THE FOLLOWING IS ONLY FOR TESTING, REMOVE
	for value := range config.Queries {
		ptr := result.Queries[value]

		fmt.Printf("%s\n", value.String())

		pts := ptr.PointsTo()

		fmt.Println(pts)
	}

	for fn := range ssautil.AllFunctions(self.ssa) {
		fmt.Println("FUNCTION:", fn)

		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				fmt.Printf("  %T: %s\n", instr, instr)
			}
		}
	}

	return nil
}

func (self *staticData) runAliasAnalysis() {
	// TODO: implement
}
