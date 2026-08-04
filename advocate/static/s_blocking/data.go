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
	PathPerRoutine        map[int]*types.Stack[[]*instructionWithInfo]
	LastBlockIdPerRoutine map[int]int
	LastReturn            map[int]*instructionWithInfo

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
		PathPerRoutine:        make(map[int]*types.Stack[[]*instructionWithInfo]),
		FuncCallToSSAFunc:     make(map[*trace.ElementFunc]*s_ssa.Function),
		LastReturn:            make(map[int]*instructionWithInfo),
		Blocked:               make(map[int]*trace.Resource),
		LastBlockIdPerRoutine: make(map[int]int),
	}
}

type instructionWithInfo struct {
	Inst     s_ssa.Instruction
	Resource map[*trace.Resource]struct{}
	Variable string
}

func (self *BlockingData) NewFuncStack(rout int, inst string) {
	res := make([]*instructionWithInfo, 1)
	if inst != "" {
		res[0] = &instructionWithInfo{Variable: inst}
	}

	self.PathPerRoutine[rout].Push(res)
}

func (self *BlockingData) ReturnStack(rout int, ret []*instructionWithInfo) {
	// insts := self.PathPerRoutine[rout].Pop()

	// addPathParam(rout, insts[0].Variable, ret[i].Resource)
}

func (self *BlockingData) NewPathPerRoutine(rout int) {
	self.PathPerRoutine[rout] = types.NewStack[[]*instructionWithInfo]()
}

func addPathInstr(rout int, inst s_ssa.Instruction, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	v := inst.Variable()

	newElem := instructionWithInfo{inst, resources, v}

	top := blocking.PathPerRoutine[rout].Pop()
	top = append(top, &newElem)
	blocking.PathPerRoutine[rout].Push(top)

	return &newElem
}

func addPathParam(rout int, v string, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	newElem := instructionWithInfo{nil, resources, v}

	top := blocking.PathPerRoutine[rout].Pop()
	top = append(top, &newElem)
	blocking.PathPerRoutine[rout].Push(top)

	return &newElem
}

func findDefOfSSAVar(rout int, v string, global bool) *instructionWithInfo {
	if strings.Contains(v, ":") {
		return &instructionWithInfo{}
	}

	ppr := blocking.PathPerRoutine[rout].Peek()

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
