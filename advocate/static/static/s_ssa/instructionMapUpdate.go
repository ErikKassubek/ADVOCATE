// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMapUpdate.go
// Brief: Map update Instruction
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

type InstructionMapUpdate struct {
	InstructionBase
}

func newMapUpdate(f *Function, inst *ssa.MapUpdate, i int) *InstructionMapUpdate {
	return &InstructionMapUpdate{InstructionBase: newInstructionBase(f, Ic_mapUpdate, inst, i)}
}

func (this *InstructionMapUpdate) Instruction() *ssa.MapUpdate {
	return this.inst.(*ssa.MapUpdate)
}

func (this *InstructionMapUpdate) setRelevant(_ *Data) {
	this.relevant = this.Conc().Resource()
	this.inTrace = false
}

func (this *InstructionMapUpdate) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionMapUpdate IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionMapUpdate) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)
	return this.Next(), info
}
