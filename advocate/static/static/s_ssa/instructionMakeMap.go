// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeMap.go
// Brief: Make map Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import "golang.org/x/tools/go/ssa"

type InstructionMakeMap struct {
	InstructionBase
}

func newMakeMap(f *Function, inst *ssa.MakeMap, i int) *InstructionMakeMap {
	return &InstructionMakeMap{InstructionBase: newInstructionBase(f, Ic_makeMap, inst, i)}
}

func (this *InstructionMakeMap) Instruction() *ssa.MakeMap {
	return this.inst.(*ssa.MakeMap)
}

func (this *InstructionMakeMap) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}
