// Copyright (c) 2026 Erik Kassubek
//
// File: ssaData.go
// Brief: Blocking Data
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"strings"
)

var data *static.Data
var blocking = newBlockData()

type BlockingData struct {
	NextPerRout           map[int]s_ssa.Instruction
	JumpBackPos           map[int]*types.Stack[s_ssa.Instruction]
	Closures              map[string]*s_ssa.Function
	LastClosure           map[int][]*instructionWithInfo
	GlobalVars            map[string]*instructionWithInfo
	PathPerRoutine        map[int][]*instructionWithInfo
	LastBlockIdPerRoutine map[int]int

	FuncCallToSSAFunc map[*trace.ElementFunc]*s_ssa.Function
	Blocked           map[int]*trace.Resource
}

func newBlockData() *BlockingData {
	return &BlockingData{
		NextPerRout:           make(map[int]s_ssa.Instruction),
		JumpBackPos:           make(map[int]*types.Stack[s_ssa.Instruction]),
		Closures:              make(map[string]*s_ssa.Function),
		LastClosure:           make(map[int][]*instructionWithInfo),
		GlobalVars:            make(map[string]*instructionWithInfo),
		PathPerRoutine:        make(map[int][]*instructionWithInfo),
		FuncCallToSSAFunc:     make(map[*trace.ElementFunc]*s_ssa.Function),
		Blocked:               make(map[int]*trace.Resource),
		LastBlockIdPerRoutine: make(map[int]int),
	}
}

type instructionWithInfo struct {
	Inst     s_ssa.Instruction
	Resource map[*trace.Resource]struct{}
	Variable string
}

func (self *BlockingData) NewPathPerRoutine(rout int) {
	self.PathPerRoutine[rout] = make([]*instructionWithInfo, 0)
}

func addPathInstr(rout int, inst s_ssa.Instruction, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	v := inst.Variable()

	newElem := instructionWithInfo{inst, resources, v}

	blocking.PathPerRoutine[rout] = append(blocking.PathPerRoutine[rout], &newElem)

	return &newElem
}

func addPathParam(rout int, v string, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	newElem := instructionWithInfo{nil, resources, v}

	blocking.PathPerRoutine[rout] = append(blocking.PathPerRoutine[rout], &newElem)

	return &newElem
}

func findDefOfSSAVar(rout int, v string, global bool) *instructionWithInfo {
	ppr := blocking.PathPerRoutine[rout]

	v = strings.TrimPrefix(v, "*")

	if global {
		return blocking.GlobalVars[v]
	}

	for i := len(ppr) - 1; i >= 0; i-- {
		if ppr[i].Variable == v {
			return ppr[i]
		}
	}

	log.Errorf("Unable to find definition of %s", v)

	return nil
}
