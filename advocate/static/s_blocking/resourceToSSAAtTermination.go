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

var nextPerRout = make(map[int]s_ssa.SsaPos)
var jumpBackPos = make(map[int]*types.Stack[s_ssa.SsaPos])
var closures = make(map[string]*s_ssa.Function)

func determineResouceToSSAAtTermination() {
	trIter := a_base.MainTrace.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]
	for id, inst := range b.Instrs() {
		if inst.InTrace() {
			nextPerRout[1] = s_ssa.NewSsaPos(f, b, inst, id)
			break
		}
		log.Debug("SKIP-> ", inst.String())
	}

	jumpBackPos[1] = types.NewStack[s_ssa.SsaPos]()

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		elemF, ok := elem.(*trace.ElementFunc)
		if ok && elemF.GetName() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			for id, inst := range b.Instrs() {
				if inst.InTrace() {
					nextPerRout[1] = s_ssa.NewSsaPos(f, b, inst, id)
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
		} else {
			delete(nextPerRout, routine)
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func passSsaAtPos(elem trace.Element, start s_ssa.SsaPos, rout int) s_ssa.SsaPos {
	if start.Nil() {
		return s_ssa.NewNilSsaPos()
	}

	next := parseInstruction(elem, start, rout)

	if next.Nil() {
		return next
	}

	return skipNonRelevant(next, rout)
}

func parseInstruction(elem trace.Element, pos s_ssa.SsaPos, rout int) s_ssa.SsaPos {
	log.Debug("RUN -> ", pos.String())

	switch inst := pos.I.(type) {
	case *s_ssa.InstructionAlloc, *s_ssa.InstructionMakeChan:
		if _, ok := blocked[elem.(*trace.ElementAlloc)]; ok {
			// TODO: store alloc
			// TODO: alloc not correct if copied
			log.Debug("IS BLOCKED ALLOC")
		}
		pos = pos.Next()
	case *s_ssa.InstructionJump:
		pos = s_ssa.NewSsaPosFuncBlock(pos.F, inst.To())
	case *s_ssa.InstructionCall:
		f := inst.GetFunc(data.Ssa())
		if f != nil {
			jumpBackPos[rout].Push(pos.Next())
			pos = s_ssa.NewSsaPosFunc(f)
		} else {
			pos = pos.Next()
		}

	case *s_ssa.InstructionReturn:
		pos = jumpBackPos[rout].Pop()
		if pos.Nil() {
			return s_ssa.NewNilSsaPos()
		}
	case *s_ssa.InstructionIf:
		elem := elem.(*trace.ElementControllFlow)
		switch elem.Type(true) {
		case trace.ControllIf:
			pos = followIfChain(pos, inst, elem.ChosenCase())
		case trace.ControllSwitch:
			pos = followSwitchChain(pos, inst, elem.ChosenCase())
		}
	case *s_ssa.InstructionGo:
		parseGo(elem, inst)
		jumpBackPos[elem.ObjID()] = types.NewStack[s_ssa.SsaPos]()
		pos = pos.Next()
	default:
		pos = pos.Next()
	}

	return pos
}

func followIfChain(pos s_ssa.SsaPos, inst s_ssa.Instruction, chosen int) s_ssa.SsaPos {
	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return pos
	}

	if chosen == 0 {
		pos.NewBlock(pos.Blocks()[instrIf.If()])
		return pos
	}

	pos.NewBlock(pos.Blocks()[instrIf.Else()])
	return followIfChain(pos, pos.I, chosen-1)
}

func followSwitchChain(pos s_ssa.SsaPos, inst s_ssa.Instruction, chosen int) s_ssa.SsaPos {
	if _, ok := inst.(*s_ssa.InstructionBinOp); ok {
		next := pos.Next()
		return followSwitchChain(next, next.I, chosen)
	}

	instrIf, ok := inst.(*s_ssa.InstructionIf)
	if !ok {
		return pos
	}

	if chosen == 0 {
		pos.NewBlock(pos.Blocks()[instrIf.If()])
		return pos
	}

	pos.NewBlock(pos.Blocks()[instrIf.Else()])
	return followSwitchChain(pos, pos.I, chosen-1)
}

func parseGo(elem trace.Element, inst *s_ssa.InstructionGo) {
	v := inst.Inst().(*ssa.Go)

	var fName string

	switch v := v.Call.Value.(type) {
	case *ssa.MakeClosure:
		fName = v.Fn.(*ssa.Function).String()

	case *ssa.Function:
		fName = v.String()
	}

	f := getSSAFuncFromName(fName)

	nextPerRout[elem.ObjID()] = s_ssa.NewSsaPosFunc(f)
}

func skipNonRelevant(pos s_ssa.SsaPos, rout int) s_ssa.SsaPos {
	for p := pos; !p.Nil(); {
		inst := p.I
		if !inst.Relevant() {
			p = p.Next()
			continue
		}

		if inst.InTrace() {

			return p
		}

		p = parseInstruction(nil, p, rout)
	}

	return s_ssa.NewNilSsaPos()
}
