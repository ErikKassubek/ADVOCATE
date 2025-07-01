// Copyright (c) 2024 Erik Kassubek
//
// File: traceElements.go
// Brief: Interface for all trace element types
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import "advocate/analysis/clock"

// Values possible primitive types
const (
	ObjectTypeAtomic     = "A"
	ObjectTypeChannel    = "C"
	ObjectTypeCond       = "D"
	ObjectTypeFork       = "G"
	ObjectTypeMutex      = "M"
	ObjectTypeNew        = "N"
	ObjectTypeOnce       = "O"
	ObjectTypeReplay     = "X"
	ObjectTypeRoutineEnd = "E"
	ObjectTypeSelect     = "S"
	ObjectTypeWait       = "W"
)

// TraceElement is an interface for the elements in a trace
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
	SetVc(vc *clock.VectorClock)
	SetWVc(vc *clock.VectorClock)
	GetVC() *clock.VectorClock
	GetWVc() *clock.VectorClock
	Copy() TraceElement
	setTraceID(ID int)
	GetTraceID() int
	AddChild(elem TraceElement)
	GetChildren() []TraceElement
	GetParents() []TraceElement
}
