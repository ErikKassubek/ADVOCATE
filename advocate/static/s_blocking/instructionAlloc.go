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
		return addPathInstr(rout, inst, make(map[*trace.Resource]struct{}))
	}

	resources := make(map[*trace.Resource]struct{})
	if r, ok := blocking.blocked[elem.ObjID()]; ok {
	resources[r] = struct{}{}
	}

	return addPathInstr(rout, inst, resources)
}
