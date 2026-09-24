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
	"go/types"
)

func ParseAlloc(inst *s_ssa.InstructionAlloc, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoAlloc(inst, rout, elem)

	return inst.Next(), info
}

func instInfoAlloc(inst *s_ssa.InstructionAlloc, rout int, elem trace.Element) *instructionWithInfo {
	elem, ok := elem.(*trace.ElementAlloc)
	if !ok {
		// get number of fields in structs
		st := inst.Instruction().Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct)
		n := st.NumFields()
		iwi := newIWI4(inst, n)
		return addPathInstr(rout, iwi)
	}

	resources := make(map[trace.Resource]bool)
	if res, ok := blocking.blockedResources[elem.ResourceID()]; ok {
		resources[res] = true
	}

	iwi := newIWI1(inst, fmtInstRes(resources))
	return addPathInstr(rout, iwi)
}
