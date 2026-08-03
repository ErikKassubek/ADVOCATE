// Copyright (c) 2026 Erik Kassubek
//
// File: InstructioRunDefers.go
// Brief: Run defers Instruction
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

func instInfoRunDefer(inst *s_ssa.InstructionRunDefers, rout int, _ trace.Element) *instructionWithInfo {
	log.Todo("InstructionRunDefers NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseRunDefer(inst *s_ssa.InstructionRunDefers, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoRunDefer(inst, rout, elem)

	log.Todo("InstructionRunDefers NOT IMPLEMENTED YET")

	return inst.Next(), info // TODO: jump in the defer?
}
