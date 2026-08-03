// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSend.go
// Brief: Send Instruction
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

type InstructionSend struct {
	InstructionBase
}

func newSend(f *Function, inst *ssa.Send, i int) *InstructionSend {
	return &InstructionSend{InstructionBase: newInstructionBase(f, Ic_send, inst, i)}
}

func (this *InstructionSend) Instruction() *ssa.Send {
	return this.inst.(*ssa.Send)
}

func (this *InstructionSend) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionSend) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	// TODO: implement
	log.Error("InstructionSend IMPLEMENTED YET")
	return addPathInstr(data, rout, this, nil)
}

func (this *InstructionSend) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	return this.Next(), info
}
