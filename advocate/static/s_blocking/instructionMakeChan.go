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
		return addPathInstr(rout, inst, newInstructionWithInfoResorce(nil))
	}

	resources := make(map[int]trace.Resource)
	if res, ok := blocking.blockedResources[elem.ResourceID()]; ok {
		resources[res.Id()] = res
	}

	return addPathInstr(rout, inst, newInstructionWithInfoResorce(resources))
}

func ParseMakeChan(inst *s_ssa.InstructionMakeChan, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoMakeChan(inst, rout, elem)
	return inst.Next(), info
}
