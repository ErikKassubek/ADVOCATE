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
	"go/token"
	"strings"

	"golang.org/x/tools/go/ssa"
)

var data *static.Data
var blocking = newBlockData()

type BlockingData struct {
	nextPerRout           map[int]s_ssa.Instruction
	jumpBackPos           map[int]*types.Stack[s_ssa.Instruction]
	closures              map[string]*s_ssa.Function
	lastClosure           map[int][]*instructionWithInfo
	globalVars            map[string]*instructionWithInfo
	pathPerRoutine        map[int]*types.Stack[path]
	lastBlockIdPerRoutine map[int]int
	chanBuffer            map[int]*types.Stack[*instructionWithInfo]
	returnVariables       map[int]*types.Stack[*s_ssa.InstructionCall]

	funcCallToSSAFunc map[*trace.ElementFunc]*s_ssa.Function
	blocked           map[trace.Element][]trace.Resource
	blockedResources  map[int]trace.Resource

	maxRoutId int
}

func newBlockData() *BlockingData {
	return &BlockingData{
		nextPerRout:           make(map[int]s_ssa.Instruction),
		jumpBackPos:           make(map[int]*types.Stack[s_ssa.Instruction]),
		closures:              make(map[string]*s_ssa.Function),
		lastClosure:           make(map[int][]*instructionWithInfo),
		globalVars:            make(map[string]*instructionWithInfo),
		pathPerRoutine:        make(map[int]*types.Stack[path]),
		funcCallToSSAFunc:     make(map[*trace.ElementFunc]*s_ssa.Function),
		blocked:               make(map[trace.Element][]trace.Resource),
		blockedResources:      make(map[int]trace.Resource),
		lastBlockIdPerRoutine: make(map[int]int),
		chanBuffer:            make(map[int]*types.Stack[*instructionWithInfo]),
		returnVariables:       make(map[int]*types.Stack[*s_ssa.InstructionCall]),
		maxRoutId:             0,
	}
}

type path []*instructionWithInfo

func (p path) last() *instructionWithInfo {
	return p[len(p)-1]
}

type instructionWithInfo struct {
	Inst     s_ssa.Instruction
	Resource []map[int]trace.Resource // make list to deal with extract.
	Variable string
	Parents  []*instructionWithInfo
}

func newIWI(inst s_ssa.Instruction, res []map[int]trace.Resource, par []*instructionWithInfo) *instructionWithInfo {
	return &instructionWithInfo{inst, res, inst.Variable(), par}
}

func newIwiFromIwi(inst s_ssa.Instruction, iwi *instructionWithInfo) *instructionWithInfo {
	return &instructionWithInfo{inst, iwi.Resource, inst.Variable(), iwi.Parents}
}

func newIWI1(inst s_ssa.Instruction, res []map[int]trace.Resource) *instructionWithInfo {
	return &instructionWithInfo{inst, res, inst.Variable(), make([]*instructionWithInfo, 0)}
}

func newIWI2(inst s_ssa.Instruction) *instructionWithInfo {
	return &instructionWithInfo{inst, make([]map[int]trace.Resource, 0), inst.Variable(), make([]*instructionWithInfo, 0)}
}

func newIWI3(inst s_ssa.Instruction, v string, res []map[int]trace.Resource) *instructionWithInfo {
	return &instructionWithInfo{inst, res, v, make([]*instructionWithInfo, 0)}
}

func (self *instructionWithInfo) Merge(other *instructionWithInfo) *instructionWithInfo {
	maxLen := max(len(self.Resource), len(other.Resource))
	res := make([]map[int]trace.Resource, maxLen)

	for i := 0; i < maxLen; i++ {
		res[i] = make(map[int]trace.Resource)
		if i < len(self.Resource) {
			for key, reso := range self.Resource[i] {
				res[i][key] = reso
			}
		}
		if i < len(other.Resource) {
			for key, reso := range other.Resource[i] {
				res[i][key] = reso
			}
		}
	}

	parents := make([]*instructionWithInfo, 0)

	for _, par := range self.Parents {
		parents = append(parents, par)
	}

	for _, par := range other.Parents {
		parents = append(parents, par)
	}

	return &instructionWithInfo{self.Inst, res, self.Variable, parents}
}

func (self *instructionWithInfo) GetResources() map[int]trace.Resource {
	res := self.Resource[0]

	for _, parent := range self.Parents {
		resPar := parent.GetResources()
		for _, par := range resPar {
			res[par.Id()] = par
		}
	}

	return res
}

func compatible(iwi *instructionWithInfo, elem trace.Element) (bool, *trace.Resource) {
	if iwi == nil || iwi.Resource == nil {
		return false, nil
	}

	for _, res := range iwi.Resource { // should be only one element, but better to be sure
		if _, ok := res[elem.ResourceID()]; !ok { // not the same object
			return false, nil
		}

		r := res[elem.ResourceID()]

		switch elem := elem.(type) {
		case *trace.ElementChannel:
			switch elem.Type(true) {
			case trace.ChannelRecv:
				return iwi.Inst.Class() == s_ssa.Ic_send, &r
			case trace.ChannelSend:
				if i, ok := iwi.Inst.Inst().(*ssa.UnOp); ok {
					return i.Op == token.ARROW, &r
				}
			}
		case *trace.ElementSelect:
			for _, c := range elem.GetCases() {
				if _, ok := res[c.ResourceID()]; !ok { // not the same object
					continue
				}
				switch c.Type(true) {
				case trace.ChannelRecv:
					return iwi.Inst.Class() == s_ssa.Ic_send, &r
				case trace.ChannelSend:
					if i, ok := iwi.Inst.Inst().(*ssa.UnOp); ok {
						return i.Op == token.ARROW, &r
					}
				}
			}
		case *trace.ElementMutex:
			if iwi.Inst.HasMutex() {
				if elem.Type(true) == trace.MutexLock && (strings.Contains(iwi.Inst.Term(), "(*sync.Mutex).Unlock(") || strings.Contains(iwi.Inst.Term(), "(*sync.RWMutex).Unlock(")) {
					return true, &r
				} else if elem.Type(true) == trace.MutexRLock && strings.Contains(iwi.Inst.Term(), "(*sync.RWMutex).RUnlock(") {
					return true, &r
				}
			}
		case *trace.ElementCond:
			return iwi.Inst.HasCond() && (strings.Contains(iwi.Inst.Term(), "(*sync.Cond).Signal(") || strings.Contains(iwi.Inst.Term(), "(*sync.Cond).Broadcast(")), &r
		case *trace.ElementWait:
			return iwi.Inst.HasWG() && strings.Contains(iwi.Inst.Term(), "(*sync.WaitGroup).Done("), &r
		}

	}

	return false, nil
}

func fmtInstRes(resource map[int]trace.Resource) []map[int]trace.Resource {
	if resource == nil {
		return make([]map[int]trace.Resource, 0)
	}

	return []map[int]trace.Resource{resource}
}

func (self *instructionWithInfo) sameResource(inst *instructionWithInfo) bool {
	if len(self.Resource) != len(inst.Resource) {
		return false
	}

	for i := 0; i < len(self.Resource); i++ {
		res1 := self.Resource[i]
		res2 := inst.Resource[i]

		if len(res1) != len(res2) {
			return false
		}

		for k := range res1 {
			if _, ok := res2[k]; !ok {
				return false
			}
		}

	}
	return true

}

func (self *BlockingData) NewFuncStack(rout int, inst *s_ssa.InstructionCall) {
	res := make([]*instructionWithInfo, 0)
	if inst != nil && inst.Variable() != "" {
		res = append(res, &instructionWithInfo{Variable: inst.Variable()})
	}

	self.pathPerRoutine[rout].Push(res)
	self.returnVariables[rout].Push(inst)
}

func (self *BlockingData) ReturnStack(rout int) *s_ssa.InstructionCall {
	self.pathPerRoutine[rout].Pop()
	return self.returnVariables[rout].Pop()
}

func (self *BlockingData) NewPathPerRoutine(rout int) {
	self.pathPerRoutine[rout] = types.NewStack[path]()
	self.returnVariables[rout] = &types.Stack[*s_ssa.InstructionCall]{}
}

func addPathInstr(rout int, iwi *instructionWithInfo) *instructionWithInfo {
	if _, ok := blocking.pathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	top := blocking.pathPerRoutine[rout].Pop()
	top = append(top, iwi)
	blocking.pathPerRoutine[rout].Push(top)

	return iwi
}

func addPathParam(rout int, v string, resources []map[int]trace.Resource) *instructionWithInfo {
	if _, ok := blocking.pathPerRoutine[rout]; !ok {
		blocking.NewPathPerRoutine(rout)
	}

	newElem := &instructionWithInfo{nil, resources, v, nil}

	top := blocking.pathPerRoutine[rout].Pop()
	top = append(top, newElem)
	blocking.pathPerRoutine[rout].Push(top)

	return newElem
}

func getDecOfSSAVar(rout int, v string) *instructionWithInfo {
	if strings.Contains(v, ":") {
		return &instructionWithInfo{}
	}

	ppr := blocking.pathPerRoutine[rout].Peek()

	v = strings.TrimPrefix(v, "*")
	v = strings.TrimPrefix(v, "<-")

	for i := len(ppr) - 1; i >= 0; i-- {
		if ppr[i].Variable == v {
			return ppr[i]
		}
	}

	// global variables
	if b, ok := blocking.globalVars[v]; ok {
		return b
	}

	log.Errorf("Unable to find declaration of %s in rout %d", v, rout)

	return nil
}
