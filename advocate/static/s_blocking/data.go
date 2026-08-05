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
	res := make([]*instructionWithInfo, 0)
	if inst != "" {
		res = append(res, &instructionWithInfo{Variable: inst})
	}

	if rout == 3 {
		log.Debug("Push 1 ", rout, " ", inst)
	}
	self.PathPerRoutine[rout].Push(res)
}

func (self *BlockingData) ReturnStack(rout int, ret []*instructionWithInfo) {
	if rout == 3 {
		log.Debug("POP 1")
	}
	self.PathPerRoutine[rout].Pop()
}

func (self *BlockingData) NewPathPerRoutine(rout int) {
	self.PathPerRoutine[rout] = types.NewStack[[]*instructionWithInfo]()
}

func addPathInstr(rout int, inst s_ssa.Instruction, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	v := inst.Variable()

	newElem := &instructionWithInfo{inst, resources, v}

	if rout == 3 {
		log.Debug("POP 2: before")
		for _, a := range blocking.PathPerRoutine[rout].Peek() {
			log.Debug(a.Variable)
		}
	}

	top := blocking.PathPerRoutine[rout].Pop()
	top = append(top, newElem)
	if rout == 3 {
		log.Debug("Push 2")
	}
	blocking.PathPerRoutine[rout].Push(top)
	if rout == 3 {
		log.Debug("POP 2: after")
		for _, a := range blocking.PathPerRoutine[rout].Peek() {
			log.Debug(a.Variable)
		}
	}

	return newElem
}

func addPathParam(rout int, v string, resources map[*trace.Resource]struct{}) *instructionWithInfo {
	if _, ok := blocking.PathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	newElem := &instructionWithInfo{nil, resources, v}

	if rout == 3 {
		log.Debug("ADD: ", newElem.Variable, " ", rout)

		log.Debug("POP 3")
	}
	top := blocking.PathPerRoutine[rout].Pop()
	top = append(top, newElem)
	if rout == 3 {
		log.Debug("Push 3")
	}
	blocking.PathPerRoutine[rout].Push(top)

	return newElem
}

func findDecOfSSAVar(rout int, v string) *instructionWithInfo {
	if strings.Contains(v, ":") {
		return &instructionWithInfo{}
	}

	ppr := blocking.PathPerRoutine[rout].Peek()

	v = strings.TrimPrefix(v, "*")

	if rout == 3 {
		for _, vari := range ppr {
			log.Debug("V: ", vari.Variable, " ", rout, " ", v)
		}
	}

	for i := len(ppr) - 1; i >= 0; i-- {
		if ppr[i].Variable == v {
			return ppr[i]
		}
	}

	if b, ok := blocking.GlobalVars[v]; ok {
		return b
	}

	log.Errorf("Unable to find declaration of %s in rout %d", v, rout)

	return nil
}
