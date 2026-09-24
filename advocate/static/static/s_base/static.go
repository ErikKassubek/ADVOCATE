// Copyright (c) 2026 Erik Kassubek
//
// File: static.go
// Brief: Base package for static analysis
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_base

type FuncName string

const (
	UnknownFunc FuncName = "<unknown>"
	MakeFunc    FuncName = "<make>"

	MakeChan      FuncName = "<chan:make>"
	ChanSend      FuncName = "<chan:send>"
	ChanRecv      FuncName = "<chan:recv>"
	ChanClose     FuncName = "<chan:close>"
	MutexLock     FuncName = "<mutex:lock>"
	MutexRLock    FuncName = "<mutex:rlock>"
	MutexTryLock  FuncName = "<mutex:trylock>"
	MutexTryRLock FuncName = "<mutex:tryrlock>"
	MutexUnlock   FuncName = "<mutex:unlock>"
	MutexRUnlock  FuncName = "<mutex:runlock>"

	CondWait      FuncName = "<cond:wait>"
	CondBroadcast FuncName = "<cond:broadcast>"
	CondSignal    FuncName = "<cond:signal>"

	WgWait FuncName = "<wait:wait>"
	WgAdd  FuncName = "<wait:add>"
	WgDone FuncName = "<wait:done>"
	WgGo   FuncName = "<wait:go>"
)

type ObjName string

const (
	UnknownObj ObjName = "unknown"

	Mutex   ObjName = "mutex"
	Channel ObjName = "chan"
	CondVar ObjName = "condVar"
	Wg      ObjName = "waitGroup"
)
