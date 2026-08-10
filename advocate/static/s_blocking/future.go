// Copyright (c) 2026 Erik Kassubek
//
// File: future.go
// Brief: Starting from the end of the execution, continue all possible paths to determine if
//        it could lead to the blocking operation beeing released.
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_blocking

import (
	"advocate/static/static/s_ssa"
	"advocate/trace"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ssa"
)

func future() {
	var iwi *instructionWithInfo
	canBeReleased := make([]trace.Element, 0)

	// TODO: finish instruction
	for len(blocking.nextPerRout) != 0 {
		for rout, next := range blocking.nextPerRout {
			blocking.nextPerRout[rout], iwi = parse(next, rout, nil)
			if elem := isBlockingRelease(iwi); len(elem) != 0 {
				canBeReleased = append(canBeReleased, elem...)
			}
		}

	}

	// TODO: report blocking bug
}

func isBlockingRelease(iwi *instructionWithInfo) []trace.Element {
	elemToDelete := make([]trace.Element, 0)
	for elem, _ := range blocking.blocked {
		if compatible(iwi, elem) {
			elemToDelete = append(elemToDelete, elem)
		}
	}

	for _, e := range elemToDelete {
		delete(blocking.blocked, e)
	}

	return elemToDelete
}

func compatible(iwi *instructionWithInfo, elem trace.Element) bool {
	for _, res := range iwi.Resource { // should be only one element, but better to be sure
		if _, ok := res[elem.ObjID()]; !ok { // not the same object
			return false
		}

		switch elem := elem.(type) {
		case *trace.ElementChannel:
			switch elem.Type(true) {
			case trace.ChannelRecv:
				return iwi.Inst.Class() == s_ssa.Ic_send
			case trace.ChannelSend:
				if i, ok := iwi.Inst.Inst().(*ssa.UnOp); ok {
					return i.Op == token.ARROW
				}
			}
		case *trace.ElementSelect:
			for _, c := range elem.GetCases() {
				if _, ok := res[c.ObjID()]; !ok { // not the same object
					continue
				}
				switch c.Type(true) {
				case trace.ChannelRecv:
					return iwi.Inst.Class() == s_ssa.Ic_send
				case trace.ChannelSend:
					if i, ok := iwi.Inst.Inst().(*ssa.UnOp); ok {
						return i.Op == token.ARROW
					}
				}
			}
		case *trace.ElementMutex:
			if iwi.Inst.HasMutex() {
				if elem.Type(true) == trace.MutexLock && (strings.Contains(iwi.Inst.Term(), "(*sync.Mutex).Unlock(") || strings.Contains(iwi.Inst.Term(), "(*sync.RWMutex).Unlock(")) {
					return true
				} else if elem.Type(true) == trace.MutexRLock && strings.Contains(iwi.Inst.Term(), "(*sync.RWMutex).RUnlock(") {
					return true
				}
			}
		case *trace.ElementCond:
			return iwi.Inst.HasCond() && (strings.Contains(iwi.Inst.Term(), "(*sync.Cond).Signal(") || strings.Contains(iwi.Inst.Term(), "(*sync.Cond).Broadcast("))
		case *trace.ElementWait:
			return iwi.Inst.HasWG() && strings.Contains(iwi.Inst.Term(), "(*sync.WaitGroup).Done(")
		}

	}

	return false
}
