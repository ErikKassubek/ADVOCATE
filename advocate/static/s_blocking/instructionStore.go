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
	decIwi := getDecOfSSAVar(rout, inst.Term())
	if decIwi == nil {
		log.Errorf("Could not find ssa var %s for %s", inst.Term(), inst)
	}
	iwi := newIwiFromIwi(inst, decIwi)

	iwi.addReferenceIndex(decIwi, 0, 0)

	if decIwi != nil && decIwi.Inst != nil {
		log.Debug("AAAAAAAAA: ", decIwi.Inst.String(), " -> ", iwi.Inst.String())
	}

	// TODO: this is not correct
	iwiOrg := getDecOfSSAVar(rout, inst.Variable())
	log.Debug(iwiOrg == nil, inst.Variable())
	if iwiOrg != nil {
		iwi.addReferenceAll(iwiOrg)
	}

	res := addPathInstr(rout, iwi)

	switch inst.Inst().(*ssa.Store).Addr.(type) {
	case *ssa.Global:
		blocking.globalVars[inst.Variable()] = res
	}
	return res
}

func ParseStore(inst *s_ssa.InstructionStore, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoStore(inst, rout, elem)

	return inst.Next(), info
}
