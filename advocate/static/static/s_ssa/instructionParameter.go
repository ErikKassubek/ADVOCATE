// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionAlloc.go
// Brief: Alloc Instruciton
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

type InstructionParameter struct {
	InstructionBase
}

func NewParameter(f *Function) *InstructionAlloc {
	return &InstructionAlloc{InstructionBase: newInstructionBase(f, Ic_param, nil, 0)}
}
