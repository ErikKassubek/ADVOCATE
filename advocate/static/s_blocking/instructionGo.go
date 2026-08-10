// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionGo.go
// Brief: Go Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/types"

	"golang.org/x/tools/go/ssa"
)

func instInfoGo(inst *s_ssa.InstructionGo, rout int, _ trace.Element) *instructionWithInfo {
	return addPathInstr(rout, inst, nil)
}

func ParseGo(inst *s_ssa.InstructionGo, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoGo(inst, rout, elem)

	v := inst.Inst().(*ssa.Go)

	var fName string

	switch v := v.Call.Value.(type) {
	case *ssa.MakeClosure:
		fName = v.Fn.(*ssa.Function).String()

	case *ssa.Function:
		fName = v.String()
	}

	f := s_ssa.GetSSAFuncFromName(data.Ssa(), fName)

	newRoutId := blocking.maxRoutId + 1
	if elem != nil {
		newRoutId = elem.ObjID()
	}

	firstInFunc := s_ssa.NewSsaPosFunc(f)

	blocking.jumpBackPos[newRoutId] = types.NewStack[s_ssa.Instruction]()

	blocking.NewPathPerRoutine(newRoutId)

	// we skip the func call in this case. For this case, perform it here
	parseCallParameter(inst.Instruction(), nil, rout, newRoutId, f, "")

	blocking.nextPerRout[newRoutId] = skipNonRelevant(firstInFunc, newRoutId)

	blocking.maxRoutId = max(blocking.maxRoutId, newRoutId)

	return inst.Next(), info
}
