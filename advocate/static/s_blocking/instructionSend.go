// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionSend.go
// Brief: Send Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/code"
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"advocate/utils/log"
	"advocate/utils/types"
	"strings"
)

var sendForward = make(map[trace.Resource]map[trace.Resource]bool)

func instInfoSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element, forward bool) *instructionWithInfo {
	log.Debug(inst.Instruction().Chan.Name())
	iwiSender := getDecOfSSAVar(rout, inst.Instruction().Chan.Name())

	if elem != nil {
		if _, ok := blocking.chanBuffer[elem.ResourceID()]; !ok {
			blocking.chanBuffer[elem.ResourceID()] = types.NewStack[*instructionWithInfo]()
		}
		blocking.chanBuffer[elem.ResourceID()].Push(iwiSender)
	}

	if forward {
		// TODO: make correct

		log.Debug("FORWARD SEND")
		for _, resSend := range iwiSender.Resource[0] {
			// chan type contains concurrency primitive
			log.Debug(resSend.Alloc())
			pos := resSend.Alloc().Pos()
			l, err := code.GetLineContent(pos)
			if err != nil {
				log.Error(err)
				continue
			}

			if !isChanConc(l) {
				continue
			}

			val := inst.Instruction().X.Name()
			iwiVal := getDecOfSSAVar(rout, val)

			for _, resVal := range iwiVal.Resource[0] {
				if _, ok := sendForward[resSend]; !ok {
					sendForward[resSend] = make(map[trace.Resource]bool)
				}
				sendForward[resSend][resVal] = true
			}

		}
	}

	if iwiSender != nil {
		return addPathInstr(rout, inst, iwiSender.Resource)
	}
	return addPathInstr(rout, inst, nil)
}

func ParseSend(inst *s_ssa.InstructionSend, rout int, elem trace.Element, forward bool) (s_ssa.Instruction, *instructionWithInfo) {
	info := instInfoSend(inst, rout, elem, forward)

	if elem != nil && !elem.Committed() {
		return nil, info
	}

	return inst.Next(), info
}

// isChanConc takes a line containing a channel make and determines,
// if the channel type is a concurrency primitive, i.e.,
// make(chan chan int), make(chan sync.Mutex), ...
// TODO: what about structs that contain conc primitives
func isChanConc(line string) bool {
	t := strings.TrimSuffix(line, ")")

	if t == line {
		return false
	}

	ct := strings.Split(t, "make(chan")

	if len(ct) <= 1 {
		return false
	}

	ctVal := ct[len(ct)-1]

	chanType := strings.TrimPrefix(ctVal, "chan")

	return strings.Contains(chanType, "chan") ||
		strings.Contains(chanType, "Mutex") ||
		strings.Contains(chanType, "RWMutex") ||
		strings.Contains(chanType, "WaitGroup") ||
		strings.Contains(chanType, "Once") ||
		strings.Contains(chanType, "Cond")
}
