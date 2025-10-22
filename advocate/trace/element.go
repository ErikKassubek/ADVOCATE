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

import (
	"advocate/analysis/hb/clock"
)

// Values for possible primitive types and functions
type ObjectType string

const (
	None ObjectType = ""

	Atomic            ObjectType = "A"
	AtomicLoad        ObjectType = "AL"
	AtomicStore       ObjectType = "AS"
	AtomicAdd         ObjectType = "AA"
	AtomicAnd         ObjectType = "AN"
	AtomicOr          ObjectType = "AO"
	AtomicSwap        ObjectType = "AW"
	AtomicCompAndSwap ObjectType = "AC"

	Channel      ObjectType = "C"
	ChannelSend  ObjectType = "CS"
	ChannelRecv  ObjectType = "CR"
	ChannelClose ObjectType = "CC"

	Cond          ObjectType = "D"
	CondWait      ObjectType = "DW"
	CondSignal    ObjectType = "DS"
	CondBroadcast ObjectType = "DB"

	Fork   ObjectType = "G"
	ForkOp ObjectType = "GG"

	End        ObjectType = "E"
	EndRoutine ObjectType = "EG"

	Mutex         ObjectType = "M"
	MutexLock     ObjectType = "ML"
	MutexRLock    ObjectType = "MR"
	MutexTryLock  ObjectType = "MT"
	MutexTryRLock ObjectType = "MY"
	MutexUnlock   ObjectType = "MU"
	MutexRUnlock  ObjectType = "MN"

	New        ObjectType = "N"
	NewAtomic  ObjectType = "NA"
	NewChannel ObjectType = "NC"
	NewCond    ObjectType = "ND"
	NewMutex   ObjectType = "NM"
	NewOnce    ObjectType = "NO"
	NewWait    ObjectType = "NW"

	Once     ObjectType = "O"
	OnceSuc  ObjectType = "OS"
	OnceFail ObjectType = "OF"

	Replay   ObjectType = "X"
	ReplayOP ObjectType = "XR"

	Select   ObjectType = "S"
	SelectOp ObjectType = "SS"

	Wait     ObjectType = "W"
	WaitAdd  ObjectType = "WA"
	WaitDone ObjectType = "WD"
	WaitWait ObjectType = "WW"
)

// Element is an interface for the elements in a trace
type Element interface {
	GetID() int
	GetTPre() int
	GetTSort() int
	GetTPost() int
	GetPos() string
	GetFile() string
	GetLine() int
	GetReplayID() string
	GetType(operation bool) ObjectType
	GetTID() string
	GetRoutine() int
	IsEqual(elem Element) bool
	IsSameElement(elem Element) bool
	GetTraceIndex() (int, int)
	SetTPre(tPre int)
	SetTSort(tSort int)
	SetTWithoutNotExecuted(tSort int)
	SetT(time int)
	ToString() string
	SetVc(vc *clock.VectorClock)
	SetWVc(vc *clock.VectorClock)
	GetVC() *clock.VectorClock
	GetWVC() *clock.VectorClock
	Copy(mapping map[string]Element) Element
	setTraceID(ID int)
	GetTraceID() int
	GetNumberConcurrent(weak, sameElem bool) int
	SetNumberConcurrent(c int, weak, sameElem bool)
	GetElemMin() (ElemMin, bool)
}
