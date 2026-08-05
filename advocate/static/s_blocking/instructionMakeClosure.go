// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeClosure.go
// Brief: Make closure Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"

	"golang.org/x/tools/go/ssa"
)

func instInfoMakeClosure(inst *s_ssa.InstructionMakeClosure, rout int, elem trace.Element) *instructionWithInfo {
	log.Todo("InstructionMakeClosure NOT IMPLEMENTED YET")
	return addPathInstr(rout, inst, nil)
}

func ParseMakeClosure(inst *s_ssa.InstructionMakeClosure, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeClosure(inst, rout, elem)

	bindings := inst.Inst().(*ssa.MakeClosure).Bindings
	blocking.lastClosure[rout] = make([]*instructionWithInfo, len(bindings))
	for i, b := range bindings {
		blocking.lastClosure[rout][i] = getDecOfSSAVar(rout, b.Name())
	}

	return inst.Next(), info
}
