// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeMap.go
// Brief: Make map Instruction
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

func instInfoMakeMap(inst *s_ssa.InstructionMakeMap, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionMakeMap NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseMakeMap(inst *s_ssa.InstructionMakeMap, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeMap(inst, rout, elem)
	return inst.Next(), info
}
