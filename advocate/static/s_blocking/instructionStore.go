// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionStore.go
// Brief: Store Instruction
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

func instInfoStore(inst *s_ssa.InstructionStore, rout int, _ trace.Element) *instructionWithInfo {
	ssaVar := findDefOfSSAVar(rout, inst.Term(), inst.TermGlobal())
	if ssaVar == nil {
		log.Errorf("Could not find ssa var %s for %s", inst.Term(), inst)
	}
	res := addPathInstr(rout, inst, ssaVar.Resource)

	switch inst.Inst().(*ssa.Store).Addr.(type) {
	case *ssa.Global:
		blocking.GlobalVars[inst.Variable()] = res
	}
	return res
}

func ParseStore(inst *s_ssa.InstructionStore, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoStore(inst, rout, elem)

	return inst.Next(), info
}
