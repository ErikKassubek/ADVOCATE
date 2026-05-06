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

import "go/ast"

type funcName int

const (
	unknown = iota

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

func (self *staticData) isMutex(id *ast.Ident) bool {
	named, ok := self.getNamed(id)

	return ok && named.Obj().Pkg().Path() == "sync" &&
		(named.Obj().Name() == "Mutex" || named.Obj().Name() == "RWMutex")
}

func (self *staticData) isCondVar(id *ast.Ident) bool {
	named, ok := self.getNamed(id)

	return ok && named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "Cond"
}

func (self *staticData) isWaitGroup(id *ast.Ident) bool {
	named, ok := self.getNamed(id)

	return ok && named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "WaitGroup"
}
