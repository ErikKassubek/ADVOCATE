// Copyright (c) 2026 Erik Kassubek
//
// File: ssaData.go
// Brief: Blocking Data
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/trace"
	"advocate/utils/types"
)

type BlockingData struct {
	NextPerRout    map[int]Instruction
	JumpBackPos    map[int]*types.Stack[Instruction]
	Closures       map[string]*Function
	LastClosure    map[int][]*InstructionWithInfo
	GlobalVars     map[string]*InstructionWithInfo
	PathPerRoutine map[int][]*InstructionWithInfo

	FuncCallToSSAFunc map[*trace.ElementFunc]*Function
	Blocked           map[int]*trace.Resource
}

func newBlockData() *BlockingData {
	return &BlockingData{
		NextPerRout:       make(map[int]Instruction),
		JumpBackPos:       make(map[int]*types.Stack[Instruction]),
		Closures:          make(map[string]*Function),
		LastClosure:       make(map[int][]*InstructionWithInfo),
		GlobalVars:        make(map[string]*InstructionWithInfo),
		PathPerRoutine:    make(map[int][]*InstructionWithInfo),
		FuncCallToSSAFunc: make(map[*trace.ElementFunc]*Function),
		Blocked:           make(map[int]*trace.Resource),
	}
}

type InstructionWithInfo struct {
	Inst     Instruction
	Resource map[*trace.Resource]struct{}
	Variable string
}

func (self *BlockingData) NewPathPerRoutine(rout int) {
	self.PathPerRoutine[rout] = make([]*InstructionWithInfo, 0)
}

func addPathInstr(blocking *BlockingData, rout int, inst Instruction, resources map[*trace.Resource]struct{}) *InstructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	v := inst.Variable()

	newElem := InstructionWithInfo{inst, resources, v}

	blocking.PathPerRoutine[rout] = append(blocking.PathPerRoutine[rout], &newElem)

	return &newElem
}
