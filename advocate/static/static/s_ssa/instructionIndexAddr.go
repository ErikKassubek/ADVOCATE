// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIndexAddr.go
// Brief: Index address Instruction
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

type InstructionIndexAddr struct {
	InstructionBase
}

func NewIndexAddr(f *Function, inst *ssa.IndexAddr, i int) *InstructionIndexAddr {
	return &InstructionIndexAddr{InstructionBase: newInstructionBase(f, Ic_indexAddr, inst, i)}
}

func (this *InstructionIndexAddr) Instruction() *ssa.IndexAddr {
	return this.inst.(*ssa.IndexAddr)
}

func (this *InstructionIndexAddr) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionIndexAddr) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionIndexAddr IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionIndexAddr) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
