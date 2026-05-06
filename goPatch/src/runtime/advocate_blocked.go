// Copyright (c) 2025 Erik Kassubek
//
// File: advocate_partial_deadlock.go
// Brief: Detect partial deadlocks while running
//
// Author: Erik Kassubek
// Created: 2025-08-01
//
// License: BSD-3-Clause

package runtime

import (
	"unsafe"
)

var CollectPartialDeadlockInfo = false
var DeadlockInfoHaveRef = make(map[uintptr][]bool) // pointer to parked operation -> list of routines with reference to this

var BlockedConcurrencyReasons = []WaitReason{
	WaitReasonChanReceiveNilChan,
	WaitReasonChanSendNilChan,
	WaitReasonSelect,
	WaitReasonSelectNoCases,
	WaitReasonChanReceive,
	WaitReasonChanSend,
	WaitReasonSyncCondWait,
	WaitReasonSyncMutexLock,
	WaitReasonSyncRWMutexRLock,
	WaitReasonSyncRWMutexLock,
	WaitReasonSyncWaitGroupWait,
}

type AdvocateG struct {
	rout *g
}

func (self *AdvocateG) ParkForeverReplay() bool {
	return self.rout.advocateRoutineInfo.parkForeverReplay
}

func (self *AdvocateG) ParkPos() string {
	return self.rout.advocateRoutineInfo.parkPos
}

func (self *AdvocateG) SetParkPos(pos string) {
	self.rout.advocateRoutineInfo.parkPos = pos
}

func (self *AdvocateG) ParkOn() []unsafe.Pointer {
	return self.rout.advocateRoutineInfo.parkOn
}

func (self *AdvocateG) Id() uint64 {
	return self.rout.advocateRoutineInfo.id
}

func (self *AdvocateG) GoId() uint64 {
	return self.rout.goid
}

func (self *AdvocateG) GetWaitReason() WaitReason {
	return self.rout.waitreason
}

func ForEachAdvocateG(fn func(adGp *AdvocateG)) {
	forEachG(func(gp *g) {
		fn(&AdvocateG{gp})
	})
}

// StorePark stores in a routine, a pointer to the last concurrency element,
// on which the routine parked
//
// Parameter:
//   - p unsafe.Pointer: pointer to the chan, (rw)mutex, wait group or conditional variable
//   - skip int: caller skip
//   - replay bool: park is forever park due to replay
func StorePark(p unsafe.Pointer, skip int, replay bool) {
	currentGoRoutineInfo().parkOn = []unsafe.Pointer{p}
	currentGoRoutineInfo().parkPos = posFromCaller(skip)
	currentGoRoutineInfo().parkForeverReplay = replay
}

// StorePark stores in a routine, a pointers to the channels involved in a
// select on which a routine parked.
// Do not call if the select has a default.
//
// Parameter:
//   - cas0 *scase: cas0 from the select implementation
//   - order0 *uint16: order0 from the select implementation
//   - ncases int: number of cases in the select (nsends+nrecvs from the select implementation)
//   - skip int: caller skip
func StoreParkSelect(cas0 *scase, order0 *uint16, ncases int, skip int) {
	cas1 := (*[1 << 16]scase)(unsafe.Pointer(cas0))

	scases := cas1[:ncases:ncases]

	currentGoRoutineInfo().parkOn = []unsafe.Pointer{}

	for _, scase := range scases {
		currentGoRoutineInfo().parkOn = append(currentGoRoutineInfo().parkOn, unsafe.Pointer(scase.c))
	}
	currentGoRoutineInfo().parkPos = posFromCaller(skip)
}

// noDeadlockSelect checks for a blocked element, if it is blocked in a select,
// and if so if all cases in the select have no running routines
//
// Parameter:
//   - opID uintptr: the element to check
//
// Returns:
//   - bool: true if the op is in a routine, where another case has channel
//     with a reference in a running routine, false if it is not blocked in
//     a select or if the select has another live reference
// func noDeadlockSelect(opID uintptr) bool {
// 	for _, ref := range waitingRef[opID] {
// 		g := routinesByID[uint64(ref)]

// 		// TODO: this should not happen, but does
// 		if g == nil {
// 			continue
// 		}
// 		if g.waitreason != waitReasonSelect {
// 			continue
// 		}

// 		for _, r := range g.advocateRoutineInfo.parkOn {
// 			if len(aliveRef[uintptr(r)]) > 0 {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// GetWaitingReasonString takes a waitReason of a routine and returns a
// string representation
//
// Parameter:
//   - wr waitReason: the wait reason enum value
//
// Returns:
//   - string: the string representation of wr
func GetWaitingReasonString(wr WaitReason) string {
	switch wr {
	case WaitReasonChanReceiveNilChan:
		return "chan:recvOnNil"
	case WaitReasonChanSendNilChan:
		return "chan:sendOnNil"
	case WaitReasonSelect:
		return "select:select"
	case WaitReasonSelectNoCases:
		return "select:withoutCases"
	case WaitReasonChanReceive:
		return "chan:revc"
	case WaitReasonChanSend:
		return "chan:send"
	case WaitReasonSyncCondWait:
		return "cond:wait"
	case WaitReasonSyncMutexLock:
		return "mutex:lock"
	case WaitReasonSyncRWMutexRLock:
		return "rwmutex:rlock"
	case WaitReasonSyncRWMutexLock:
		return "rwmutex:lock"
	case WaitReasonSyncWaitGroupWait:
		return "waitGroup:wait"
	}
	return "unknown:unknown"
}

func ReadyStatusWaiting(gp *AdvocateG) bool {
	return readgstatus(gp.rout) == _Gwaiting
}
