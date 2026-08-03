// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSlice.go
// Brief: Slice Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
)

func instInfoSlice(inst *s_ssa.InstructionSlice, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionSlice NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseSlice(inst *s_ssa.InstructionSlice, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSlice(inst, rout, elem)

	return inst.Next(), info
}
