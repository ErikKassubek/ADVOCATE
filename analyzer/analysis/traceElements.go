// Copyrigth (c) 2024 Erik Kassubek
//
// File: traceElements.go
// Brief: Interface for all trace element types
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import "analyzer/clock"

// Interface for trace elements
type TraceElement interface {
	GetID() int
	GetTPre() int
	GetTSort() int
	GetTPost() int
	GetPos() string
	GetFile() string
	GetLine() int
	GetObjType() string
	GetTID() string
	GetRoutine() int
	SetTPre(tPre int)
	SetTSort(tSort int)
	SetTWithoutNotExecuted(tSort int)
	SetT(time int)
	ToString() string
	updateVectorClock()
	GetVC() clock.VectorClock
	GetVCWmHB() clock.VectorClock
	Copy() TraceElement
}
