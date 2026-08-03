// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeChan.go
// Brief: Make channel Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"

	"golang.org/x/tools/go/ssa"
)

type InstructionMakeChan struct {
	InstructionBase
}

func newMakeChan(f *Function, inst *ssa.MakeChan, i int) *InstructionMakeChan {
	return &InstructionMakeChan{InstructionBase: newInstructionBase(f, Ic_makeChan, inst, i)}
}

func (this *InstructionMakeChan) Instruction() *ssa.MakeChan {
	return this.inst.(*ssa.MakeChan)
}

func (this *InstructionMakeChan) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionMakeChan) addInstructionWithInfo(data *BlockingData, rout int, elem trace.Element) *InstructionWithInfo {
	elem, ok := elem.(*trace.ElementAlloc)
	if !ok {
		return addPathInstr(data, rout, this, make(map[*trace.Resource]struct{}))
	}

	resources := make(map[*trace.Resource]struct{})
	if r, ok := data.Blocked[elem.ObjID()]; ok {
		resources[r] = struct{}{}
	}

	return addPathInstr(data, rout, this, resources)
}

func (this *InstructionMakeChan) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
