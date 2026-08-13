// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionPhi.go
// Brief: Phi Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
)

// TODO: can this be used to reduce double pass

func instInfoPhi(inst *s_ssa.InstructionPhi, rout int, _ trace.Element) *instructionWithInfo {
	lastBlock := blocking.lastBlockIdPerRoutine[rout]

	pred := inst.GetPred()[lastBlock]

	ssaVar := getDecOfSSAVar(rout, pred)

	return addPathInstr(rout, inst, ssaVar.Resource)
}

func ParsePhi(inst *s_ssa.InstructionPhi, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoPhi(inst, rout, elem)
	return inst.Next(), info
}
