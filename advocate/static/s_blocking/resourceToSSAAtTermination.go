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

var nextPerRout = make(map[int]s_ssa.Instruction)
var jumpBackPos = make(map[int]*types.Stack[s_ssa.Instruction])
var closures = make(map[string]*s_ssa.Function)
var lastClosure = make(map[int][]*instructionWithInfo)
var globalVars = make(map[string]*instructionWithInfo)

func determineResouceToSSAAtTermination() {
	trIter := a_base.MainTrace.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]
	for _, inst := range b.Instrs() {
		if inst.InTrace() {
			nextPerRout[1] = inst
			break
		}
	}

	jumpBackPos[1] = types.NewStack[s_ssa.Instruction]()

	lastWasFork := make(map[int]bool)

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		if elemF, ok := elem.(*trace.ElementFunc); ok && elemF.Name() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			for _, inst := range b.Instrs() {
				if inst.InTrace() {
					nextPerRout[1] = inst
					break
				}
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

		next := parseInstructions(elem, nextPerRout[routine], routine)
		if next != nil {
			nextPerRout[routine] = next
		} else {
			delete(nextPerRout, routine)
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func parseInstructions(elem trace.Element, inst s_ssa.Instruction, rout int) s_ssa.Instruction {
	if inst == nil {
		return nil
	}

	next := parseInstruction(elem, inst, rout)

	if next == nil {
		return next
	}

	return skipNonRelevant(next, rout)
}

func parseInstruction(elem trace.Element, inst s_ssa.Instruction, rout int) s_ssa.Instruction {
	info := addInstructionWithInfo(rout, inst, elem)

	infoStr := "<NIL>"
	if info != nil && len(info.resource) != 0 {
		infoStr = ""
		for r := range info.resource {
			infoStr += fmt.Sprint(r.Alloc().ObjID())
		}
	}

	if elem != nil {
		log.Debug("ELEM1 -> ", inst.String(), " -> ", elem.StringDebug(), " -> ", infoStr)
	} else {
		log.Debug("ELEM2 -> ", inst.String(), " -> ", infoStr)
	}

	switch i := inst.(type) {
	case *s_ssa.InstructionAlloc, *s_ssa.InstructionMakeChan:
		inst = inst.Next()
	case *s_ssa.InstructionJump:
		inst = s_ssa.NewSsaPosFuncBlock(inst.Function(), i.To())
	case *s_ssa.InstructionCall:
		f := i.GetFunc(data.Ssa())
		if f != nil {
			jumpBackPos[rout].Push(inst.Next())
			inst = s_ssa.NewSsaPosFunc(f)
		} else {
			inst = inst.Next()
		}

	case *s_ssa.InstructionReturn:
		inst = jumpBackPos[rout].Pop()
		if inst == nil {
			return nil
		}
	case *s_ssa.InstructionIf:
		elem := elem.(*trace.ElementControllFlow)
		switch elem.Type(true) {
		case trace.ControllIf:
			inst = followIfChain(inst, elem.ChosenCase())
		case trace.ControllSwitch:
			inst = followSwitchChain(inst, elem.ChosenCase())
		}
	case *s_ssa.InstructionMakeClosure:
		bindings := inst.Inst().(*ssa.MakeClosure).Bindings
		lastClosure[rout] = make([]*instructionWithInfo, len(bindings))
		for i, b := range bindings {
			lastClosure[rout][i] = findDefOfSSAVar(rout, b.Name(), false)
		}
		inst = inst.Next()
	case *s_ssa.InstructionGo:
		f := parseGo(rout, elem, i)
		jumpBackPos[elem.ObjID()] = types.NewStack[s_ssa.Instruction]()

		// we skip the func call in this case. For this case, perform it here
		parseNewFunc(rout, elem.ObjID(), f)

		inst = inst.Next()
	default:
		inst = inst.Next()
	}

	return inst
}

func skipNonRelevant(inst s_ssa.Instruction, rout int) s_ssa.Instruction {
	for p := inst; p != nil; {
		if !p.Relevant() {
			p = p.Next()
			continue
		}

		if p.InTrace() {
			return p
		}

		p = parseInstruction(nil, p, rout)
	}

	return nil
}
