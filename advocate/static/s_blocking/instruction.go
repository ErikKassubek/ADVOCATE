// Copyright (c) 2026 Erik Kassubek
//
// File: instruction.go
// Brief: One instructions and instruction paths
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ssa"
)

type path []*instructionWithInfo

func (p path) last() *instructionWithInfo {
	return p[len(p)-1]
}

type instructionWithInfo struct {
	Inst      s_ssa.Instruction
	Variable  string
	Resource  map[int]map[trace.Resource]bool // index (for field, return/extrace, ...) ->  resource
	Reference map[int][]*instructionWithInfo  // index (for field, return/extract, ...) -> []parents
}

func newIWI(inst s_ssa.Instruction, res map[int]map[trace.Resource]bool, par map[int][]*instructionWithInfo) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), res, par}
}

func newIwiFromIwi(inst s_ssa.Instruction, iwi *instructionWithInfo) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), iwi.Resource, iwi.Reference}
}

func newIwiFromIwiIndex(inst s_ssa.Instruction, iwi *instructionWithInfo, index int) *instructionWithInfo {
	newRes := make(map[int]map[trace.Resource]bool)
	newRes[index] = iwi.Resource[index]

	newRef := make(map[int][]*instructionWithInfo)
	newRef[index] = iwi.Reference[index]

	return &instructionWithInfo{inst, inst.Variable(), newRes, newRef}
}

func newIWI1(inst s_ssa.Instruction, res map[int]map[trace.Resource]bool) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), res, make(map[int][]*instructionWithInfo)}
}

func newIWI2(inst s_ssa.Instruction) *instructionWithInfo {
	return &instructionWithInfo{inst, inst.Variable(), make(map[int]map[trace.Resource]bool), make(map[int][]*instructionWithInfo)}
}

func newIWI3(inst s_ssa.Instruction, v string, res map[int]map[trace.Resource]bool) *instructionWithInfo {
	return &instructionWithInfo{inst, v, res, make(map[int][]*instructionWithInfo)}
}

func newIWI4(inst s_ssa.Instruction, n int) *instructionWithInfo {
	log.Debug("IWI4: ", inst, n)

	resources := make(map[int]map[trace.Resource]bool)
	references := make(map[int][]*instructionWithInfo)

	for i := 0; i < n; i++ {
		resources[i] = map[trace.Resource]bool{}
		references[i] = make([]*instructionWithInfo, 0)
	}

	return &instructionWithInfo{inst, inst.Variable(), resources, references}
}

func (self *instructionWithInfo) Merge(other *instructionWithInfo) *instructionWithInfo {
	maxLen := max(len(self.Resource), len(other.Resource))
	new_resources := make(map[int]map[trace.Resource]bool, maxLen)

	for i, resource := range self.Resource {
		if _, ok := new_resources[i]; !ok {
			new_resources[i] = make(map[trace.Resource]bool)
		}

		for res := range resource {
			new_resources[i][res] = true
		}
	}

	for i, resource := range other.Resource {
		if _, ok := new_resources[i]; !ok {
			new_resources[i] = make(map[trace.Resource]bool)
		}

		for res := range resource {
			new_resources[i][res] = true
		}
	}

	new_parents := make(map[int][]*instructionWithInfo, 0)

	for i, parents := range self.Reference {
		new_parents[i] = parents
	}

	for i, parents := range other.Reference {
		if _, ok := new_parents[i]; !ok {
			new_parents[i] = make([]*instructionWithInfo, 0)
		}

		new_parents[i] = append(new_parents[i], parents...)
	}

	return &instructionWithInfo{self.Inst, self.Variable, new_resources, new_parents}
}

func (self *instructionWithInfo) GetResources() map[trace.Resource]bool {
	visited := make(map[*instructionWithInfo]bool)
	return self.GetResourcesRec(visited)
}

func (self *instructionWithInfo) GetResourcesRec(visited map[*instructionWithInfo]bool) map[trace.Resource]bool {
	result := make(map[trace.Resource]bool)

	visited[self] = true

	for _, resMap := range self.Resource {
		for res := range resMap {
			result[res] = true
		}
	}

	for _, parents := range self.Reference {
		for _, parent := range parents {
			if _, ok := visited[parent]; ok {
				continue
			}
			for res := range parent.GetResourcesRec(visited) {
				result[res] = true
			}
		}
	}

	return result
}

func (self *instructionWithInfo) GetResourcesIndex(index int) map[trace.Resource]bool {
	result := self.Resource[index]

	for _, parent := range self.Reference[index] {
		for res := range parent.GetResources() {
			result[res] = true
		}
	}

	return result
}

func (self *instructionWithInfo) GetResourcesSlice() []map[trace.Resource]bool {
	if self == nil {
		return []map[trace.Resource]bool{}
	}
	res := make([]map[trace.Resource]bool, len(self.Resource))

	for i := range res {
		res[i] = self.GetResourcesIndex(i)
	}

	return res
}

func compatible(iwi *instructionWithInfo, elem trace.Element) (bool, *trace.Resource) {
	if iwi == nil || iwi.Resource == nil {
		return false, nil
	}

	for _, res := range iwi.Resource { // should be only one element, but better to be sure
		if _, ok := res[elem.Resource()]; !ok { // not the same object
			return false, nil
		}

		r := elem.Resource()

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
				if _, ok := res[c.Resource()]; !ok { // not the same object
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

func (self *instructionWithInfo) String() string {
	res := self.Inst.String() + " | "

	for index, resources := range self.Resource {
		res += fmt.Sprintf(" %d <", index)
		for resource := range resources {
			res += fmt.Sprint(resource.Id(), " ")
		}
		res += "> <"
		references := self.Reference[index]
		for _, ref := range references {
			res += fmt.Sprint(ref.Inst, " ")
		}
		res += "> | "
	}

	return res
}

func fmtInstRes(resource map[trace.Resource]bool) map[int]map[trace.Resource]bool {
	if resource == nil {
		return make(map[int]map[trace.Resource]bool, 0)
	}

	res := make(map[int]map[trace.Resource]bool, 1)
	res[0] = resource

	return res
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

func setReference(iwi1 *instructionWithInfo, index1 int, iwi2 *instructionWithInfo, index2 int) {
	iwi1.Reference[index1] = append(iwi1.Reference[index1], iwi2.Reference[index2]...)
	iwi2.Reference[index2] = append(iwi2.Reference[index2], iwi1.Reference[index1]...)
}
