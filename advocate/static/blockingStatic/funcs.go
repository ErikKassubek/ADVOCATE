// Copyright (c) 2026 Erik Kassubek
//
// File: funcs.go
// Brief: enums for all relevant functions
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blockingStatic

import (
	"fmt"
	"go/ast"
)

type funcName string

const (
	unknownFunc funcName = "<unknown>"
	makeFunc    funcName = "<make>"

	makeChan  funcName = "<chan:make>"
	chanSend  funcName = "<chan:send>"
	chanRecv  funcName = "<chan:recv>"
	chanClose funcName = "<chan:close>"

	mutexLock     funcName = "<mutex:lock>"
	mutexRLock    funcName = "<mutex:rlock>"
	mutexTryLock  funcName = "<mutex:trylock>"
	mutexTryRLock funcName = "<mutex:tryrlock>"
	mutexUnlock   funcName = "<mutex:unlock>"
	mutexRUnlock  funcName = "<mutex:runlock>"

	condWait      funcName = "<cond:wait>"
	condBroadcast funcName = "<cond:broadcast>"
	condSignal    funcName = "<cond:signal>"

	wgWait funcName = "<wait:wait>"
	wgAdd  funcName = "<wait:add>"
	wgDone funcName = "<wait:done>"
	wgGo   funcName = "<wait:go>"

	// TODO: list all
)

type objName int

const (
	unknownObj objName = iota

	mutex
	channel
	condVar
	wg
)

// func isCompatibleFunc(a, b funcName) bool {
// 	// a should be less then
// 	if int(a) > int(b) {
// 		a, b = b, a
// 	}

// 	switch a {
// 	case chanSend:
// 		return b == chanRecv
// 	case chanRecv:
// 		return b == chanClose
// 	case mutexLock, mutexTryLock:
// 		return b == mutexUnlock
// 	case mutexRLock, mutexTryRLock:
// 		return b == mutexRUnlock
// 	case condWait:
// 		return b == condBroadcast || b == condSignal
// 	case wgWait:
// 		return b == wgDone
// 	}

// 	return false
// }

func (self *staticData) getName(id ast.Expr) string {
	if id == nil {
		return "NIL"
	}

	switch e := id.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return self.getName(e.X) + "." + e.Sel.Name
	case *ast.FuncLit:
		return "FuncLit"
	case *ast.CallExpr:
		return self.getName(e.Fun)
	default:
		panic(fmt.Sprintf("Unknown expr type %T", e))
	}
}
