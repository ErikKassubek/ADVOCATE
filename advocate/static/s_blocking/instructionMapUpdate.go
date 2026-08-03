// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMapUpdate.go
// Brief: Map update Instruction
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

func instInfoMapUpdate(inst *s_ssa.InstructionMapUpdate, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionMapUpdate NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseMapUpdate(inst *s_ssa.InstructionMapUpdate, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMapUpdate(inst, rout, elem)
	return inst.Next(), info
}
