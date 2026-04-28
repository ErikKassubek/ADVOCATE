// Copyright (c) 2026 Erik Kassubek
//
// File: funcs.go
// Brief: enums for all relevant functions
//
// Author: Erik Kassubek
// Created: 2026-04-28
//
// License: BSD-3-Clause

package blocking

type funcs int

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

	// TODO: list all
)

// func (self *staticData) isCompatible(a, b *ast.Ident) bool {
// 	if !self.sameVar(a, b) {
// 		return false
// 	}

// 	funcA, err := getFuncsFromIdent(a)
// 	if err != nil {
// 		return false
// 	}

// 	funcB, err := getFuncsFromIdent(b)
// 	if err != nil {
// 		return false
// 	}
// 	return isCompatibleFunc(funcA, funcB)
// }

func isCompatibleFunc(a, b funcs) bool {

	// TODO: implement
	return false
}
