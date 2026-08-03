// Copyright (c) 2026 Erik Kassubek
//
// File: InstructioRunDefers.go
// Brief: Run defers Instruction
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

type InstructionRunDefers struct {
	InstructionBase
}

func newRunDefers(f *Function, inst *ssa.RunDefers, i int) *InstructionRunDefers {
	return &InstructionRunDefers{InstructionBase: newInstructionBase(f, Ic_runDefers, inst, i)}
}

func (this *InstructionRunDefers) Instruction() *ssa.RunDefers {
	return this.inst.(*ssa.RunDefers)
}

func (this *InstructionRunDefers) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionRunDefers) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionRunDefers IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionRunDefers) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	log.Error("InstructionRunDefers IMPLEMENTED YET")

	return this.Next(), info // TODO: jump in the defer?
}
