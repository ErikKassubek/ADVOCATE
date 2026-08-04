// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMultiConvert.go
// Brief: Multi convert Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"golang.org/x/tools/go/ssa"
)

type InstructionMultiConvert struct {
	InstructionBase
}

func newMultiConvert(f *Function, inst *ssa.MultiConvert, i int) *InstructionMultiConvert {
	return &InstructionMultiConvert{InstructionBase: newInstructionBase(f, Ic_multiConvert, inst, i)}
}

func (this *InstructionMultiConvert) Instruction() *ssa.MultiConvert {
	return this.inst.(*ssa.MultiConvert)
}

func (this *InstructionMultiConvert) setRelevant(_ *Data) {
	this.relevant = false
	this.inTrace = false
}
