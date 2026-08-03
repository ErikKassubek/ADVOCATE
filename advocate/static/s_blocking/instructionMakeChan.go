// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionMakeChan.go
// Brief: Make channel Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

func instInfoMakeChan(inst *s_ssa.InstructionMakeChan, rout int, elem trace.Element) *instructionWithInfo {
	elem, ok := elem.(*trace.ElementAlloc)
	if !ok {
		return addPathInstr(rout, inst, make(map[*trace.Resource]struct{}))
	}

	resources := make(map[*trace.Resource]struct{})
	if r, ok := blocking.Blocked[elem.ObjID()]; ok {
		resources[r] = struct{}{}
	}

	return addPathInstr(rout, inst, resources)
}

func ParseMakeChan(inst *s_ssa.InstructionMakeChan, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeChan(inst, rout, elem)
	return inst.Next(), info
}
