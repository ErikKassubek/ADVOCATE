// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeClosure.go
// Brief: Make closure Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"advocate/utils/log"

	"golang.org/x/tools/go/ssa"
)

type InstructionMakeClosure struct {
	InstructionBase
}

func newMakeClosure(f *Function, inst *ssa.MakeClosure, i int) *InstructionMakeClosure {
	return &InstructionMakeClosure{InstructionBase: newInstructionBase(f, Ic_makeClosure, inst, i)}
}

func (this *InstructionMakeClosure) Instruction() *ssa.MakeClosure {
	return this.inst.(*ssa.MakeClosure)
}

func (this *InstructionMakeClosure) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionMakeClosure) addInstructionWithInfo(data *BlockingData, rout int, elem trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionMakeClosure IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionMakeClosure) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	blocking := data.Blocking

	info := this.addInstructionWithInfo(blocking, rout, elem)

	bindings := this.Inst().(*ssa.MakeClosure).Bindings
	blocking.LastClosure[rout] = make([]*InstructionWithInfo, len(bindings))
	for i, b := range bindings {
		blocking.LastClosure[rout][i] = findDefOfSSAVar(blocking, rout, b.Name(), false)
	}

	return this.Next(), info
}
