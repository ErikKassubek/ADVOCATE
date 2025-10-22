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
type OperationType string

const (
	None OperationType = ""

	Atomic            OperationType = "A"
	AtomicLoad        OperationType = "AL"
	AtomicStore       OperationType = "AS"
	AtomicAdd         OperationType = "AA"
	AtomicAnd         OperationType = "AN"
	AtomicOr          OperationType = "AO"
	AtomicSwap        OperationType = "AW"
	AtomicCompAndSwap OperationType = "AC"

	Channel      OperationType = "C"
	ChannelSend  OperationType = "CS"
	ChannelRecv  OperationType = "CR"
	ChannelClose OperationType = "CC"

	Cond          OperationType = "D"
	CondWait      OperationType = "DW"
	CondSignal    OperationType = "DS"
	CondBroadcast OperationType = "DB"

	Fork   OperationType = "G"
	ForkOp OperationType = "GG"

	End        OperationType = "E"
	EndRoutine OperationType = "EG"

	Mutex         OperationType = "M"
	MutexLock     OperationType = "ML"
	MutexRLock    OperationType = "MR"
	MutexTryLock  OperationType = "MT"
	MutexTryRLock OperationType = "MY"
	MutexUnlock   OperationType = "MU"
	MutexRUnlock  OperationType = "MN"

	New        OperationType = "N"
	NewAtomic  OperationType = "NA"
	NewChannel OperationType = "NC"
	NewCond    OperationType = "ND"
	NewMutex   OperationType = "NM"
	NewOnce    OperationType = "NO"
	NewWait    OperationType = "NW"

	Once     OperationType = "O"
	OnceSuc  OperationType = "OS"
	OnceFail OperationType = "OF"

	Replay   OperationType = "X"
	ReplayOP OperationType = "XR"

	Select   OperationType = "S"
	SelectOp OperationType = "SS"

	Wait     OperationType = "W"
	WaitAdd  OperationType = "WA"
	WaitDone OperationType = "WD"
	WaitWait OperationType = "WW"
)

// GetElemTypeFromObjectType returns the object type from the operation type
//
// Parameter:
//   - ob OperationType: the operation or object type
//
// Returns:
//   - OperationType: the corresponding object type
func GetElemTypeFromObjectType(ob OperationType) OperationType {
	switch ob {
	case Atomic, AtomicLoad, AtomicStore, AtomicAdd, AtomicAnd, AtomicOr, AtomicSwap, AtomicCompAndSwap:
		return Atomic
	case Channel, ChannelSend, ChannelRecv, ChannelClose:
		return Channel
	case Cond, CondWait, CondSignal, CondBroadcast:
		return Cond
	case Fork, ForkOp:
		return Fork
	case End, EndRoutine:
		return End
	case Mutex, MutexLock, MutexRLock, MutexTryLock, MutexUnlock, MutexRUnlock:
		return Mutex
	case Once, OnceSuc, OnceFail:
		return Once
	case Replay, ReplayOP:
		return Replay
	case Select, SelectOp:
		return Select
	case Wait, WaitAdd, WaitDone, WaitWait:
		return Wait
	default:
		return None
	}
}

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
	GetType(operation bool) OperationType
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
