// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIndex.go
// Brief: Index Instruction
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

type InstructionIndex struct {
	InstructionBase
}

func NewIndex(f *Function, inst *ssa.Index, i int) *InstructionIndex {
	return &InstructionIndex{InstructionBase: newInstructionBase(f, Ic_index, inst, i)}
}

func (this *InstructionIndex) Instruction() *ssa.Index {
	return this.inst.(*ssa.Index)
}

func (this *InstructionIndex) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionIndex) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionIndex IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionIndex) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
