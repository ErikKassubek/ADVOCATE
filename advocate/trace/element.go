// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/element.go
// Brief: Interface for all trace element types
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"advocate/utils/consts"
	"fmt"
)

// ========================================================
// MARK: Element
// ========================================================

// Element is an interface for the elements in a trace
type Element interface {
	ID() int
	setID(ID int)
	ObjID() int

	T(t timeType) int
	SetT(t timeType, time int)
	SetTWithoutNotExecuted(tSort int)
	Committed() bool

	Pos() string
	File() string
	Line() int

	Routine() int
	TraceIndex() (int, int)

	Type(operation bool) OperationType

	IsEqual(elem Element) bool
	IsSameElement(elem Element) bool

	String() string

	Function() *ElementFunc // TODO: GLOBAL

	Vc(weak a_clock.VcType, vc *a_clock.VectorClock)
	GetVC(weak a_clock.VcType) *a_clock.VectorClock

	NumberConcurrent(weak, sameElem bool) int
	SetNumberConcurrent(c int, weak, sameElem bool)

	ReplayID() string

	Copy(mapping map[int]Element, keep bool) Element

	IsValid() bool
}

func IsOp(elem Element) bool {
	switch elem.(type) {
	case *ElementAlloc, *ElementReplay, *ElementRoutineEnd:
		return false
	}

	return true
}

// ========================================================
// MARK: Operation
// ========================================================

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

	Controll       OperationType = "I"
	ControllIf     OperationType = "II"
	ControllSwitch OperationType = "IS"

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

// ========================================================
// MARK: Time
// ========================================================

type timeType int

const (
	Request timeType = iota
	Commit
	Both = iota
	Sorting
)

// ========================================================
// MARK: Concurrent
// ========================================================

type concInfo struct {
	vc                       *a_clock.VectorClock
	wVc                      *a_clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

func newConcInfo() *concInfo {
	return &concInfo{nil, nil, -1, -1, -1, -1}
}

func (this *concInfo) copy() *concInfo {
	return &concInfo{
		this.vc.Copy(),
		this.wVc.Copy(),
		this.numberConcurrent,
		this.numberConcurrentWeak,
		this.numberConcurrentSame,
		this.numberConcurrentWeakSame,
	}
}

func (this *concInfo) getVC(weak a_clock.VcType) *a_clock.VectorClock {
	if weak == a_clock.Weak {
		return this.wVc
	}
	return this.vc
}

func (this *concInfo) setVC(weak a_clock.VcType, vc *a_clock.VectorClock) {
	if weak == a_clock.Weak {
		this.wVc = vc
		return
	}
	this.vc = vc
}

func (this *concInfo) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return this.numberConcurrentWeakSame
		}
		return this.numberConcurrentWeak
	}
	if sameElem {
		return this.numberConcurrentSame
	}
	return this.numberConcurrent
}

func (this *concInfo) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			this.numberConcurrentWeakSame = c
		} else {
			this.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			this.numberConcurrentSame = c
		} else {
			this.numberConcurrent = c
		}
	}
}

// ========================================================
// MARK: Position
// ========================================================

type position struct {
	file string
	line int
}

func newPosition(file string, line int) position {
	return position{file, line}
}

func (this position) copy() position {
	return position{
		this.file, this.line,
	}
}

func (this position) toString() string {
	return fmt.Sprintf("%s%s%d", this.file, consts.PosSep, this.line)
}
