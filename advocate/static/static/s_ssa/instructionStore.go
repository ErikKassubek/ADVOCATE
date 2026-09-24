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
