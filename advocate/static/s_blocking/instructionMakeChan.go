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
		iwi := newIWI2(inst)
		return addPathInstr(rout, iwi)
	}

	resources := make(map[int]trace.Resource)
	if res, ok := blocking.blockedResources[elem.ResourceID()]; ok {
		resources[res.Id()] = res
	}

	iwi := newIWI1(inst, fmtInstRes(resources))
	return addPathInstr(rout, iwi)
}

func ParseMakeChan(inst *s_ssa.InstructionMakeChan, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeChan(inst, rout, elem)
	return inst.Next(), info
}
