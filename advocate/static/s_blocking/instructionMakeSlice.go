// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeSlice.go
// Brief: Make slice Instruction
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

func instInfoMakeSlice(inst *s_ssa.InstructionMakeSlice, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionMakeSlice NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseMakeSlice(inst *s_ssa.InstructionMakeSlice, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeSlice(inst, rout, elem)
	return inst.Next(), info
}
