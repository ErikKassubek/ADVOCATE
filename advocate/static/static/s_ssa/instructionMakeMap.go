// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeMap.go
// Brief: Make map Instruction
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

type InstructionMakeMap struct {
	InstructionBase
}

func newMakeMap(f *Function, inst *ssa.MakeMap, i int) *InstructionMakeMap {
	return &InstructionMakeMap{InstructionBase: newInstructionBase(f, Ic_makeMap, inst, i)}
}

func (this *InstructionMakeMap) Instruction() *ssa.MakeMap {
	return this.inst.(*ssa.MakeMap)
}

func (this *InstructionMakeMap) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = false
}

func (this *InstructionMakeMap) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionMakeMap IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionMakeMap) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
