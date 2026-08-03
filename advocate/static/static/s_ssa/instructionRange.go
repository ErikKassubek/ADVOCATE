// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionRange.go
// Brief: Range Instruction
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

type InstructionRange struct {
	InstructionBase
}

func newRange(f *Function, inst *ssa.Range, i int) *InstructionRange {
	return &InstructionRange{InstructionBase: newInstructionBase(f, Ic_range, inst, i)}
}

func (this *InstructionRange) Instruction() *ssa.Range {
	return this.inst.(*ssa.Range)
}

func (this *InstructionRange) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionRange) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionRange IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionRange) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
