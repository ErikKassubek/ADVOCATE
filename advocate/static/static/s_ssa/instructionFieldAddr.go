// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionFieldAddr.go
// Brief: Field Address Instruction
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

type InstructionFieldAddr struct {
	InstructionBase
}

func NewFieldAddr(f *Function, inst *ssa.FieldAddr, i int) *InstructionFieldAddr {
	return &InstructionFieldAddr{InstructionBase: newInstructionBase(f, Ic_field, inst, i)}
}

func (this *InstructionFieldAddr) Instruction() *ssa.FieldAddr {
	return this.inst.(*ssa.FieldAddr)
}

func (this *InstructionFieldAddr) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionFieldAddr) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionFieldAddr IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionFieldAddr) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
