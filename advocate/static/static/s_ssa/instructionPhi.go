// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionPhi.go
// Brief: Phi Instruction
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

type InstructionPhi struct {
	InstructionBase
}

func newPhi(f *Function, inst *ssa.Phi, i int) *InstructionPhi {
	return &InstructionPhi{InstructionBase: newInstructionBase(f, Ic_phi, inst, i)}
}

func (this *InstructionPhi) Instruction() *ssa.Phi {
	return this.inst.(*ssa.Phi)
}

func (this *InstructionPhi) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionPhi) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionPhi IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionPhi) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
