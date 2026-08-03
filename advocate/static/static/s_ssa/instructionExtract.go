// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionExtrace.go
// Brief: Extract Instruction
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

type InstructionExtract struct {
	InstructionBase
}

func NewExtract(f *Function, inst *ssa.Extract, i int) *InstructionExtract {
	return &InstructionExtract{InstructionBase: newInstructionBase(f, Ic_extract, inst, i)}
}

func (this *InstructionExtract) Instruction() *ssa.Extract {
	return this.inst.(*ssa.Extract)
}

func (this *InstructionExtract) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = false
}

func (this *InstructionExtract) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("INInstructionExtractOT IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionExtract) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
