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
	"go/ast"
)

type funcName int

const (
	unknownFunc funcName = iota

	chanSend
	chanRecv
	chanClose

	mutexLock
	mutexRLock
	mutexTryLock
	mutexTryRLock
	mutexUnlock
	mutexRUnlock

	condWait
	condBroadcast
	condSignal

	wgWait
	wgAdd
	wgDone
	wgGo

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

func isCompatibleFunc(a, b funcName) bool {
	// a should be less then
	if int(a) > int(b) {
		a, b = b, a
	}

	switch a {
	case chanSend:
		return b == chanRecv
	case chanRecv:
		return b == chanClose
	case mutexLock, mutexTryLock:
		return b == mutexUnlock
	case mutexRLock, mutexTryRLock:
		return b == mutexRUnlock
	case condWait:
		return b == condBroadcast || b == condSignal
	case wgWait:
		return b == wgDone
	}

	return false
}

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
	default:
		panic("Unknown expr type")
	}
}
