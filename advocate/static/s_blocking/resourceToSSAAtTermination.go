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
	blocking := data.Blocking()

	trIter := a_base.MainTrace.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]
	for _, inst := range b.Instrs() {
		if inst.InTrace() {
			blocking.NextPerRout[1] = inst
			break
		}
	}

	blocking.JumpBackPos[1] = types.NewStack[s_ssa.Instruction]()

	lastWasFork := make(map[int]bool)

	for elem := trIter.Next(); elem != nil; elem = trIter.Next() {
		// Set main as func
		if elemF, ok := elem.(*trace.ElementFunc); ok && elemF.Name() == "main.main" {
			f := data.Ssa().MainFunc()
			b := f.Blocks()[0]

			// skip non relevant instructions in main
			for _, inst := range b.Instrs() {
				if inst.InTrace() {
					blocking.NextPerRout[1] = inst
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

		next := parseInstructions(elem, blocking.NextPerRout[routine], routine)
		if next != nil {
			blocking.NextPerRout[routine] = next
		} else {
			delete(blocking.NextPerRout, routine)
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
	next, info := inst.Parse(data.Ssa(), rout, elem)

	infoStr := "<NIL>"
	if info != nil && len(info.Resource) != 0 {
		infoStr = ""
		for r := range info.Resource {
			infoStr += fmt.Sprint(r.Alloc().ObjID())
		}
	}

	if elem != nil {
		log.Debug("ELEM1 -> ", inst.String(), " -> ", elem.StringDebug(), " -> ", infoStr)
	} else {
		log.Debug("ELEM2 -> ", inst.String(), " -> ", infoStr)
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

		p = parseInstruction(nil, p, rout)
	}

	return nil
}
