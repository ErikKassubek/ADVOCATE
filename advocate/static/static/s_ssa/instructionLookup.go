// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionLookup.go
// Brief: Lookup Instruction
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

type InstructionLookup struct {
	InstructionBase
}

func newLookup(f *Function, inst *ssa.Lookup, i int) *InstructionLookup {
	return &InstructionLookup{InstructionBase: newInstructionBase(f, Ic_lookup, inst, i)}
}

func (this *InstructionLookup) Instruction() *ssa.Jump {
	return this.inst.(*ssa.Jump)
}

func (this *InstructionLookup) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionLookup) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionLookup IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionLookup) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
