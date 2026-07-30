// Copyright (c) 2026 Erik Kassubek
//
// File: resource.go
// Brief: Determine instructions which can point to resources
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"regexp"
)

type instructionWithInfo struct {
	inst     s_ssa.Instruction
	resource map[*trace.Resource]struct{}
	variable string
}

var pathPerRoutine = make(map[int][]*instructionWithInfo)

func newPathPerRoutine(rout int) {
	pathPerRoutine[rout] = make([]*instructionWithInfo, 0)
}

func addInstructionWithInfo(rout int, instr s_ssa.Instruction, elem trace.Element) *instructionWithInfo {
	switch instr := instr.(type) {
	case *s_ssa.InstructionAlloc, *s_ssa.InstructionMakeChan:
		elem := elem.(*trace.ElementAlloc)
		resources := make(map[*trace.Resource]struct{})
		if r, ok := blocked[elem]; ok {
			resources[r] = struct{}{}
		}
		return addPathInstr(rout, instr, resources)
	case *s_ssa.InstructionStore:
		ssaVar := findDefOfSSAVar(rout, instr.Term())
		if ssaVar == nil {
			log.Debug("NIL: ", instr.String())
		}
		return addPathInstr(rout, instr, ssaVar.resource)
	case *s_ssa.InstructionUnOp:
		re := regexp.MustCompile(`^\*f\d+$`) // *f...
		ssaVar := findDefOfSSAVar(rout, instr.Term())
		if re.MatchString(instr.Term()) {
			return addPathInstr(rout, instr, ssaVar.resource)
		}
	}

	return nil
}

func addPathInstr(rout int, inst s_ssa.Instruction, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := pathPerRoutine[rout]; !ok {
		newPathPerRoutine(rout)
	}

	v := inst.Variable()

	newElem := instructionWithInfo{inst, resources, v}

	pathPerRoutine[rout] = append(pathPerRoutine[rout], &newElem)

	return &newElem
}

func addPathParam(rout int, v string, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := pathPerRoutine[rout]; !ok {
		newPathPerRoutine(rout)
	}

	newElem := instructionWithInfo{nil, resources, v}

	pathPerRoutine[rout] = append(pathPerRoutine[rout], &newElem)

	return &newElem
}
