// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionJump.go
// Brief: Jump Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

func ParseJump(inst *s_ssa.InstructionJump, _ int, _ trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	return s_ssa.NewSsaPosFuncBlock(inst.Function(), inst.To()), nil
}
