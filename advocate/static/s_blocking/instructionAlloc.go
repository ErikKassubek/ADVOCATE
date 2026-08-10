// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionAlloc.go
// Brief: Alloc Instruciton
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

func ParseAlloc(inst *s_ssa.InstructionAlloc, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoAlloc(inst, rout, elem)

	return inst.Next(), info
}

func instInfoAlloc(inst *s_ssa.InstructionAlloc, rout int, elem trace.Element) *instructionWithInfo {
	elem, ok := elem.(*trace.ElementAlloc)
	if !ok {
		return addPathInstr(rout, inst, newInstructionWithInfoResorce(nil))
	}

	resources := make(map[int]trace.Resource)
	if res, ok := blocking.blocked[elem]; ok {
		for _, r := range res {
			resources[r.Id()] = r
		}
	}

	return addPathInstr(rout, inst, newInstructionWithInfoResorce(resources))
}
