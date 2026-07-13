// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/element.go
// Brief: Interface for all trace element types
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
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

	Replay   OperationType = "R"
	ReplayOP OperationType = "RR"

	Select   OperationType = "S"
	SelectOp OperationType = "SS"

	Wait     OperationType = "W"
	WaitAdd  OperationType = "WA"
	WaitDone OperationType = "WD"
	WaitWait OperationType = "WW"

	Func       OperationType = "F"
	FuncCall   OperationType = "FC"
	FuncReturn OperationType = "FR"

	UnknownOperation OperationType = "XX"
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
	case Func, FuncCall, FuncReturn:
		return Func
	default:
		return None
	}
}

type timeType int

const (
	Request timeType = iota
	Commit
	Both = iota
	Sorting
)

// Element is an interface for the elements in a trace
type Element interface {
	GetID() int
	setID(ID int)
	GetObjId() int

	GetT(t timeType) int
	SetT(t timeType, time int)
	SetTWithoutNotExecuted(tSort int)
	Committed() bool

	GetPos() string
	GetFile() string
	GetLine() int

	GetRoutine() int
	GetTraceIndex() (int, int)

	GetType(operation bool) OperationType

	IsEqual(elem Element) bool // TODO: fix
	IsSameElement(elem Element) bool

	ToString() string

	SetVc(weak a_clock.VcType, vc *a_clock.VectorClock)
	GetVC(weak a_clock.VcType) *a_clock.VectorClock

	GetNumberConcurrent(weak, sameElem bool) int
	SetNumberConcurrent(c int, weak, sameElem bool)

	GetReplayID() string
	GetTID() string

	Copy(mapping map[string]Element, keep bool) Element

	IsValid() bool
}

func IsOp(elem Element) bool {
	switch elem.(type) {
	case *ElementAlloc, *ElementReplay, *ElementRoutineEnd:
		return false
	}

	return true
}
