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
	return this.i.String()
}

func (this *ssaPos) Next() ssaPos {
	instID := this.instID + 1
	i := this.Instrs()[instID]
	return ssaPos{f: this.f, b: this.b, i: i, instID: instID}
}

var nextPerRout = make(map[int]ssaPos)
var jumpBackPos = make(map[int]*types.Stack[ssaPos])

func determineResouceToSSAAtTermination(res map[*trace.Resource]struct{}) {
	trIter := a_base.MainTrace.AsIterator()

	// get the first relevant value in init
	f := data.Ssa().InitFunc()
	b := f.Blocks()[1]
	for _, inst := range b.Instrs() {
		if inst.InTrace() {
			nextPerRout[1] = ssaPos{f: f, b: b, i: inst}
			break
		}
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
			}
			continue
		}

		routine := elem.Routine()

		log.Debug(elem.StringDebug())
		res := passFromSsaPos(nextPerRout[routine], routine)
		if res.Nil() {
			delete(nextPerRout, routine)
		} else {
			nextPerRout[routine] = res
		}
	}
}

// Iterate over SSA starting from start and stopping before end
func passFromSsaPos(start ssaPos, rout int) ssaPos {
	if start.Nil() {
		return ssaPos{nil, nil, nil, 0}
	}
	log.Debug("PASS: ", start.String())
	next := passInstruction(start, rout)

	if next.Nil() {
		return next
	}

	for i, inst := range next.b.Instrs() {
		if i <= next.instID {
			continue
		}
		log.Debug("INST: ", inst.String())

		if !inst.Relevant() {
			log.Debug("NOT RELEVANT: ", inst.String())
			continue
		}

		if inst.InTrace() {
			log.Debug("IN TRACE: ", inst.String())
			return ssaPos{f: next.f, b: next.b, i: inst, instID: i}
		}

		log.Debug(inst.String())
		return ssaPos{f: next.f, b: next.b, i: inst, instID: i}

	}

	return ssaPos{nil, nil, nil, 0}
}

func passInstruction(pos ssaPos, rout int) ssaPos {
	switch inst := pos.i.(type) {
	case *s_ssa.InstructionJump:
		pos.b = pos.Blocks()[inst.To()]
	case *s_ssa.InstructionCall:
		f := inst.GetFunc(data.Ssa())
		s := newSsaPosFunc(f)
		jumpBackPos[rout].Push(pos.Next())
		return passFromSsaPos(s, rout)
	case *s_ssa.InstructionReturn:
		pos = jumpBackPos[rout].Pop()
	case *s_ssa.InstructionIf:
		// TODO: for now we always go into if. Implement check from trace
		pos.b = pos.Blocks()[inst.If()]
	default:
		pos = pos.Next()
	}

	return pos
}
