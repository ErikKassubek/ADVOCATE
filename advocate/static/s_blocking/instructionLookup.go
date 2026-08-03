// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionLookup.go
// Brief: Lookup Instruction
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

func instInfoLookup(inst *s_ssa.InstructionLookup, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionLookup NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseLookup(inst *s_ssa.InstructionLookup, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoLookup(inst, rout, elem)
	return inst.Next(), info
}
