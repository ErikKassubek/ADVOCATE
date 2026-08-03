// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMapUpdate.go
// Brief: Map update Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionMapUpdate struct {
	InstructionBase
}

func newMapUpdate(f *Function, inst *ssa.MapUpdate, i int) *InstructionMapUpdate {
	return &InstructionMapUpdate{InstructionBase: newInstructionBase(f, Ic_mapUpdate, inst, i)}
}

func (this *InstructionMapUpdate) Instruction() *ssa.MapUpdate {
	return this.inst.(*ssa.MapUpdate)
}

func (this *InstructionMapUpdate) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = false
}
