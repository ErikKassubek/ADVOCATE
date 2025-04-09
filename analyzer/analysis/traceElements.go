// Copyright (c) 2024 Erik Kassubek
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

const (
	ObjectTypeAtomic     = "A"
	ObjectTypeChannel    = "C"
	ObjectTypeCond       = "D"
	ObjectTypeFork       = "R"
	ObjectTypeMutex      = "M"
	ObjectTypeNew        = "N"
	ObjectTypeOnce       = "O"
	ObjectTypeReplay     = "R"
	ObjectTypeRoutineEnd = "E"
	ObjectTypeSelect     = "S"
	ObjectTypeWait       = "W"
)

// Interface for trace elements
type TraceElement interface {
	GetID() int
	GetTPre() int
	GetTSort() int
	GetTPost() int
	GetPos() string
	GetFile() string
	GetLine() int
	GetReplayID() string
	GetObjType(operation bool) string
	GetTID() string
	GetRoutine() int
	IsEqual(elem TraceElement) bool
	GetTraceIndex() (int, int)
	SetTPre(tPre int)
	SetTSort(tSort int)
	SetTWithoutNotExecuted(tSort int)
	SetT(time int)
	ToString() string
	updateVectorClock()
	GetVC() *clock.VectorClock
	GetVCWmHB() *clock.VectorClock
	Copy() TraceElement
	AddRel1(elem TraceElement, pos int)
	AddRel2(elem TraceElement)
	GetRel1() []TraceElement
	GetRel2() []TraceElement
}
