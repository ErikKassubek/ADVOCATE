// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionStore.go
// Brief: Store Instruction
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

type InstructionStore struct {
	InstructionBase
}

func newStore(f *Function, inst *ssa.Store, i int) *InstructionStore {
	return &InstructionStore{InstructionBase: newInstructionBase(f, Ic_store, inst, i)}
}

func (this *InstructionStore) Instruction() *ssa.Store {
	return this.inst.(*ssa.Store)
}

func (this *InstructionStore) setRelevant(_ *Data) {
	resource := this.Conc().Resource()

	this.relevant = resource
	this.inTrace = false
}

func (this *InstructionStore) addInstructionWithInfo(data *BlockingData, rout int, _ trace.Element) *InstructionWithInfo {
	ssaVar := findDefOfSSAVar(data, rout, this.Term(), this.TermGlobal())
	if ssaVar == nil {
		log.Errorf("Could not find ssa var %s for %s", this.Term(), this)
	}
	res := addPathInstr(data, rout, this, ssaVar.Resource)
	switch this.Inst().(*ssa.Store).Addr.(type) {
	case *ssa.Global:
		data.GlobalVars[this.Variable()] = res
	}
	return res
}

func (this *InstructionStore) Parse(data *Data, rout int, elem trace.Element) (Instruction, *InstructionWithInfo) {
	info := this.addInstructionWithInfo(data.Blocking, rout, elem)

	return this.Next(), info
}
