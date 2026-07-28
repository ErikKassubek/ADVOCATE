// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: For a given alloc, get all SSA variables that are equal to the allocated resource when following the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/analysis/a_base"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"

	"golang.org/x/tools/go/ssa"
)

type ssaPos struct {
	f      *s_ssa.Function
	b      *s_ssa.Block
	i      s_ssa.Instruction
	instID int
}

func newSsaPosFunc(f *s_ssa.Function) ssaPos {
	b := f.Blocks()[0]
	i := b.Instrs()[0]
	return ssaPos{f: f, b: b, i: i, instID: 0}
}

func newSsaPosFuncBlock(f *s_ssa.Function, blockID int) ssaPos {
	b := f.Blocks()[blockID]
	i := b.Instrs()[0]
	return ssaPos{f: f, b: b, i: i, instID: 0}
}

func (this *ssaPos) Blocks() []*s_ssa.Block {
	return this.f.Blocks()
}

func (this *ssaPos) Instrs() []s_ssa.Instruction {
	return this.b.Instrs()
}

func (this *ssaPos) Nil() bool {
	return this.f == nil
}

func (this *ssaPos) String() string {
	f := this.f.Name()
	if f == "" {
		f = "init"
	}
	return f + " : " + this.i.String()
}

func (this *ssaPos) Next() ssaPos {
	instID := this.instID + 1
	i := this.b.Instrs()[instID]
	res := ssaPos{f: this.f, b: this.b, i: i, instID: instID}
	return res
}

var nextPerRout = make(map[int]ssaPos)
var jumpBackPos = make(map[int]*types.Stack[ssaPos])
var closures = make(map[string]*s_ssa.Function)

func determineResouceToSSAAtTermination() {
	trIter := a_base.MainTrace.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]
	for _, inst := range b.Instrs() {
		if inst.InTrace() {
			nextPerRout[1] = ssaPos{f: f, b: b, i: inst}
			break
		}
		log.Debug("SKIP-> ", inst.String())
	}

	jumpBackPos[1] = types.NewStack[ssaPos]()

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		elemF, ok := elem.(*trace.ElementFunc)
		if ok && elemF.GetName() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			for _, inst := range b.Instrs() {
				if inst.InTrace() {
					nextPerRout[1] = ssaPos{f: f, b: b, i: inst}
					break
				}
				log.Debug("SKIP: ", inst.String())
			}
			continue
		}

		switch elem.(type) {
		case *trace.ElementReplay, *trace.ElementRoutineEnd:
			continue
		}

		routine := elem.Routine()

		// log.Debug(elem.StringDebug())
		res := passSsaAtPos(elem, nextPerRout[routine], routine)
		if !res.Nil() {
			nextPerRout[routine] = res
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func passSsaAtPos(elem trace.Element, start ssaPos, rout int) ssaPos {
	if start.Nil() {
		return ssaPos{nil, nil, nil, 0}
	}

	next := parseInstruction(elem, start, rout)

	if next.Nil() {
		return next
	}

	return skipNonRelevant(next, rout)
}

func parseInstruction(elem trace.Element, pos ssaPos, rout int) ssaPos {
	log.Debug("RUN -> ", pos.String())

	switch inst := pos.i.(type) {
	case *s_ssa.InstructionAlloc, *s_ssa.InstructionMakeChan:
		if _, ok := blocked[elem.(*trace.ElementAlloc)]; ok {
			// TODO: store alloc
			// TODO: alloc not correct if copied
			log.Debug("IS BLOCKED ALLOC")
		}
		pos = pos.Next()
	case *s_ssa.InstructionJump:
		pos = newSsaPosFuncBlock(pos.f, inst.To())
	case *s_ssa.InstructionCall:
		f := inst.GetFunc(data.Ssa())
		if f != nil {
			pos = newSsaPosFunc(f)
			jumpBackPos[rout].Push(pos.Next())
		} else {
			pos = pos.Next()
		}

	case *s_ssa.InstructionReturn:
		pos = jumpBackPos[rout].Pop()
	case *s_ssa.InstructionIf:
		elem := elem.(*trace.ElementControllFlow)
		switch elem.Type(true) {
		case trace.ControllIf:
			pos = followIfChain(pos, inst, elem.ChosenCase())
		case trace.ControllSwitch:
			pos = followSwitchChain(pos, inst, elem.ChosenCase())
		}
	case *s_ssa.InstructionGo:
		parseGo(elem, pos, inst)
		jumpBackPos[elem.ObjID()] = types.NewStack[ssaPos]()
		pos = pos.Next()
	default:
		pos = pos.Next()
	}

	return pos
}

func followIfChain(pos ssaPos, inst s_ssa.Instruction, chosen int) ssaPos {
	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return pos
	}

	if chosen == 0 {
		pos.b = pos.Blocks()[instrIf.If()]
		pos.instID = 0
		pos.i = pos.b.Instrs()[0]
		return pos
	}

	pos.b = pos.Blocks()[instrIf.Else()]
	pos.instID = 0
	pos.i = pos.b.Instrs()[0]
	return followIfChain(pos, pos.i, chosen-1)
}

func followSwitchChain(pos ssaPos, inst s_ssa.Instruction, chosen int) ssaPos {
	if _, ok := inst.(*s_ssa.InstructionBinOp); ok {
		next := pos.Next()
		return followSwitchChain(next, next.i, chosen)
	}

	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return pos
	}

	if chosen == 0 {
		pos.b = pos.Blocks()[instrIf.If()]
		pos.instID = 0
		pos.i = pos.b.Instrs()[0]
		return pos
	}

	pos.b = pos.Blocks()[instrIf.Else()]
	pos.instID = 0
	pos.i = pos.b.Instrs()[0]
	return followSwitchChain(pos, pos.i, chosen-1)
}

func parseGo(elem trace.Element, pos ssaPos, inst *s_ssa.InstructionGo) {
	v := inst.Inst().(*ssa.Go)

	var fName string

	switch v := v.Call.Value.(type) {
	case *ssa.MakeClosure:
		fName = v.Fn.(*ssa.Function).String()

	case *ssa.Function:
		fName = v.String()
	}

	f := getSSAFuncFromName(fName)

	nextPerRout[elem.ObjID()] = newSsaPosFunc(f)
}

func skipNonRelevant(pos ssaPos, rout int) ssaPos {
	for p := pos; !p.Nil(); {
		inst := p.i
		if !inst.Relevant() {
			p = p.Next()
			continue
		}

		if inst.InTrace() {

			return p
		}

		p = parseInstruction(nil, p, rout)
	}

	return ssaPos{nil, nil, nil, 0}
}
