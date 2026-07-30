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
	"fmt"

	"golang.org/x/tools/go/ssa"
)

var nextPerRout = make(map[int]s_ssa.SsaPos)
var jumpBackPos = make(map[int]*types.Stack[s_ssa.SsaPos])
var closures = make(map[string]*s_ssa.Function)
var lastClosure = make(map[int][]*instructionWithInfo)

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
		// log.Debug("SKIP-> ", inst.String())
	}

	jumpBackPos[1] = types.NewStack[s_ssa.SsaPos]()

	lastWasFork := make(map[int]bool)

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		elemF, ok := elem.(*trace.ElementFunc)
		if ok && elemF.Name() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			for id, inst := range b.Instrs() {
				if inst.InTrace() {
					nextPerRout[1] = s_ssa.NewSsaPos(f, b, inst, id)
					break
				}
				// log.Debug("SKIP-> ", inst.String())
			}
			continue
		}

		routine := elem.Routine()
		if _, ok := elem.(*trace.ElementFunc); ok && lastWasFork[routine] {
			lastWasFork[routine] = false
			continue
		}

		switch elem.(type) {
		case *trace.ElementReplay, *trace.ElementRoutineEnd:
			continue
		case *trace.ElementFork:
			lastWasFork[elem.ObjID()] = true
		default:
			lastWasFork[routine] = false
		}

		next := passSsaAtPos(elem, nextPerRout[routine], routine)
		if !next.Nil() {
			nextPerRout[routine] = next
		} else {
			delete(nextPerRout, routine)
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func passSsaAtPos(elem trace.Element, pos s_ssa.SsaPos, rout int) s_ssa.SsaPos {
	if pos.Nil() {
		return s_ssa.NewNilSsaPos()
	}

	next := parseInstruction(elem, pos, rout)

	if next.Nil() {
		return next
	}

	return skipNonRelevant(next, rout)
}

func parseInstruction(elem trace.Element, pos s_ssa.SsaPos, rout int) s_ssa.SsaPos {
	info := addInstructionWithInfo(rout, pos.I, elem)

	infoStr := "<NIL>"
	if info != nil {
		infoStr = ""
		for r := range info.resource {
			infoStr += fmt.Sprint(r.Alloc().ObjID())
		}
	}
	log.Debug("ELEM  -> ", elem.StringDebug(), " -> ", pos.I.String(), " -> ", infoStr)

	switch inst := pos.I.(type) {
	case *s_ssa.InstructionAlloc, *s_ssa.InstructionMakeChan:
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
		elem, ok := elem.(*trace.ElementControllFlow)
		if !ok {
			panic("A")
			log.Debug(elem.IsValid())
			log.Debug("INVALID: ", elem.String())
		}
		switch elem.Type(true) {
		case trace.ControllIf:
			pos = followIfChain(pos, inst, elem.ChosenCase())
		case trace.ControllSwitch:
			pos = followSwitchChain(pos, inst, elem.ChosenCase())
		}
	case *s_ssa.InstructionMakeClosure:
		bindings := inst.Inst().(*ssa.MakeClosure).Bindings
		lastClosure[rout] = make([]*instructionWithInfo, len(bindings))
		for i, b := range bindings {
			lastClosure[rout][i] = findDefOfSSAVar(rout, b.Name())
		}
	case *s_ssa.InstructionGo:
		f := parseGo(rout, elem, inst)
		jumpBackPos[elem.ObjID()] = types.NewStack[s_ssa.SsaPos]()

		// we skip the func call in this case. For this case, perform it here
		parseNewFunc(rout, f)

		pos = pos.Next()
	default:
		pos = pos.Next()
	}

	return pos
}

func skipNonRelevant(pos s_ssa.SsaPos, rout int) s_ssa.SsaPos {
	for p := pos; !p.Nil(); {
		inst := p.I
		if !inst.Relevant() {
			// log.Debug("SKIPR -> ", inst.String())
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
