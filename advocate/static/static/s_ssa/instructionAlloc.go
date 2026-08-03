// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionAlloc.go
// Brief: Alloc Instruciton
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

type InstructionAlloc struct {
	InstructionBase
}

func NewAlloc(f *Function, inst *ssa.Alloc, i int) *InstructionAlloc {
	return &InstructionAlloc{InstructionBase: newInstructionBase(f, Ic_alloc, inst, i)}
}

func (this *InstructionAlloc) Instruction() *ssa.Alloc {
	return this.inst.(*ssa.Alloc)
}

func (this *InstructionAlloc) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = !this.Conc().Channel()
}

func (this *InstructionAlloc) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	return this.Next(), info
}

func (this *InstructionAlloc) addInstructionWithInfo(blocking *BlockingData, rout int, elem trace.Element) *InstructionWithInfo {
	elem, ok := elem.(*trace.ElementAlloc)
	if !ok {
		return addPathInstr(blocking, rout, this, make(map[*trace.Resource]struct{}))
	}

	resources := make(map[*trace.Resource]struct{})
	if r, ok := blocking.Blocked[elem.ObjID()]; ok {
		resources[r] = struct{}{}
	}

	return addPathInstr(blocking, rout, this, resources)
}
