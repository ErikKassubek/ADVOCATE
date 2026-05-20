// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Run alias analysis
//
// Author: Erik Kassubek
// Created: 2026-03-25
//
// License: BSD-3-Clause

package blockingStatic

import (
	"advocate/utils/log"
	"go/ast"

	"golang.org/x/tools/go/ssa"
)

func (self *staticData) runAliasAnalysis() {
	for _, pkg := range self.ssaPkgs {
		for key, member := range pkg.Members {
			log.Important(key, " -> ", member.String())
			fn, ok := member.(*ssa.Function)
			if !ok || fn.Blocks == nil {
				continue
			}

			// *ssa.Store: write to memory
			// *ssa.Assign: direct assignment
			// *ssa.Alloc + *ssa.Store: variable initialization
			// *ssa.FieldAddr + *ssa.Store: struct field assignment
			// *ssa.IndexAddr + *ssa.Store: slice/array assignment
			for _, block := range fn.Blocks {
				for _, instr := range block.Instrs {
					switch instr.(type) {
					case *ssa.Store:
						log.Debug("Assign: ", instr.String())

					case *ssa.Alloc:
						log.Debug("Alloc: ", instr.String())
					}
				}
			}
		}

		// for key, val := range pkg.St
	}
}

// sameVar determines, if two variables are the same
// TODO: this is not correct.
// TODO: replace
//
// Parameter:
//   - id1 *ast.Ident: first variable
//   - id2 *ast.Ident: second variable
func (self *staticData) sameVar(id1, id2 *ast.Ident) bool {
	obj1, pkg1 := self.getObject(id1)
	obj2, pkg2 := self.getObject(id2)

	if obj1 == nil || obj2 == nil || pkg1 != pkg2 {
		return false
	}

	return obj1 == obj2
}
