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
)

func determineResouceToSSAAtTermination() {
	trIter := a_base.MainTrace.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]
	for _, inst := range b.Instrs() {
		if inst.InTrace() {
			blocking.nextPerRout[1] = inst
			break
		}
	}

	blocking.jumpBackPos[1] = types.NewStack[s_ssa.Instruction]()
	blocking.NewPathPerRoutine(1)

	lastWasFork := make(map[int]bool)

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		if elemF, ok := elem.(*trace.ElementFunc); ok && elemF.Name() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			for _, inst := range b.Instrs() {
				if inst.InTrace() {
					blocking.nextPerRout[1] = inst
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

		next := parseInstructions(elem, blocking.nextPerRout[routine], routine)
		if next != nil {
			blocking.nextPerRout[routine] = next
		} else {
			delete(blocking.nextPerRout, routine)
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func parseInstructions(elem trace.Element, inst s_ssa.Instruction, rout int) s_ssa.Instruction {
	if inst == nil {
		return nil
	}

	next := parseInstruction(inst, rout, elem)

	if next == nil {
		return next
	}

	return skipNonRelevant(next, rout)
}

func parseInstruction(inst s_ssa.Instruction, rout int, elem trace.Element) s_ssa.Instruction {
	if elem != nil && !elem.Committed() {
		return inst.Next()
	}

	next, info := parse(inst, rout, elem)

	infoStr := "<NIL>"
	if info != nil && len(info.Resource) != 0 {
		infoStr = ""
		for r := range info.Resource {
			infoStr += fmt.Sprint(r.Alloc().ObjID())
		}
	}

	if elem != nil {
		log.Debug("ELEM1 $ ", inst.StringInfo(), " $ ", elem.StringDebug(), " $ ", infoStr)
	} else {
		log.Debug("ELEM2 $ ", inst.StringInfo(), " $ ", infoStr)
	}

	return next
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

		p = parseInstruction(p, rout, nil)

		blocking.lastBlockIdPerRoutine[rout] = inst.Inst().Block().Index

	}

	return nil
}

func parse(inst s_ssa.Instruction, rout int, elem trace.Element) (s_ssa.Instruction, *instructionWithInfo) {
	if inst == nil {
		return nil, nil
	}

	switch inst := inst.(type) {
	case *s_ssa.InstructionAlloc:
		return ParseAlloc(inst, rout, elem)
	case *s_ssa.InstructionCall:
		return ParseCall(inst, rout, elem)
	case *s_ssa.InstructionExtract:
		return ParseExtract(inst, rout, elem)
	case *s_ssa.InstructionField:
		return ParseField(inst, rout, elem)
	case *s_ssa.InstructionFieldAddr:
		return ParseFieldAddr(inst, rout, elem)
	case *s_ssa.InstructionGo:
		return ParseGo(inst, rout, elem)
	case *s_ssa.InstructionIf:
		return ParseIf(inst, rout, elem)
	case *s_ssa.InstructionIndex:
		return ParseIndex(inst, rout, elem)
	case *s_ssa.InstructionIndexAddr:
		return ParseIndexAddr(inst, rout, elem)
	case *s_ssa.InstructionJump:
		return ParseJump(inst, rout, elem)
	case *s_ssa.InstructionLookup:
		return ParseLookup(inst, rout, elem)
	case *s_ssa.InstructionMakeChan:
		return ParseMakeChan(inst, rout, elem)
	case *s_ssa.InstructionMakeClosure:
		return ParseMakeClosure(inst, rout, elem)
	case *s_ssa.InstructionMakeInterface:
		return ParseMakeInterface(inst, rout, elem)
	case *s_ssa.InstructionMakeMap:
		return ParseMakeMap(inst, rout, elem)
	case *s_ssa.InstructionMakeSlice:
		return ParseMakeSlice(inst, rout, elem)
	case *s_ssa.InstructionMapUpdate:
		return ParseMapUpdate(inst, rout, elem)
	case *s_ssa.InstructionNext:
		return ParseNext(inst, rout, elem)
	case *s_ssa.InstructionPhi:
		return ParsePhi(inst, rout, elem)
	case *s_ssa.InstructionRange:
		return ParseRange(inst, rout, elem)
	case *s_ssa.InstructionReturn:
		return ParseReturn(inst, rout, elem)
	case *s_ssa.InstructionRunDefers:
		return ParseRunDefer(inst, rout, elem)
	case *s_ssa.InstructionSelect:
		return ParseSelect(inst, rout, elem)
	case *s_ssa.InstructionSend:
		return ParseSend(inst, rout, elem)
	case *s_ssa.InstructionSlice:
		return ParseSlice(inst, rout, elem)
	case *s_ssa.InstructionStore:
		return ParseStore(inst, rout, elem)
	case *s_ssa.InstructionUnOp:
		return ParseUnOp(inst, rout, elem)
	default:
		return inst.Next(), nil
	}
}
