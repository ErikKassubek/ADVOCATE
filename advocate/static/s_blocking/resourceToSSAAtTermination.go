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
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
)

func determineResouceToSSAAtTermination(res map[*trace.Resource]struct{}) {
	passFunction(data.Ssa().InitFunc())
	passFunction(data.Ssa().MainFunc())
}

func passFunction(f *s_ssa.Function) {
	blocks := f.Blocks()

	currentBlock := blocks[0]

	for currentBlock != nil {
		insts := currentBlock.Instrs()

		for _, inst := range insts {
			if !inst.Relevant() {
				continue
			}

			log.Debug(inst.String())

			switch inst := inst.(type) {
			case *s_ssa.InstructionJump:
				currentBlock = blocks[inst.To()]
			case *s_ssa.InstructionCall:
				f := inst.GetFunc(data.Ssa())
				passFunction(f)
			case *s_ssa.InstructionReturn:
				currentBlock = nil
			case *s_ssa.InstructionIf:
				// TODO: for now we always go into if. Implement forward or check from trace
				currentBlock = blocks[inst.If()]
			}
		}
	}
}
