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

var blockedConcurrencyReasons = []WaitReason{
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

func (self *AdvocateG) parkForeverReplay() bool {
	return self.rout.advocateRoutineInfo.parkForeverReplay
}

func (self *AdvocateG) parkPos() string {
	return self.rout.advocateRoutineInfo.parkPos
}

func (self *AdvocateG) setParkPos(pos string) {
	self.rout.advocateRoutineInfo.parkPos = pos
}

func (self *AdvocateG) parkInfo() []park {
	return self.rout.advocateRoutineInfo.parkObj
}

func (self *AdvocateG) isNil() bool {
	return self.rout.advocateRoutineInfo == nil
}

func (self *AdvocateG) id() uint64 {
	return self.rout.advocateRoutineInfo.id
}

func (self *AdvocateG) goId() uint64 {
	return self.rout.goid
}

func (self *AdvocateG) getWaitReason() WaitReason {
	return self.rout.waitreason
}

func (self *AdvocateG) getForkPos() string {
	return self.rout.advocateRoutineInfo.GetForkPos()
}

func (self *AdvocateG) setOAT(oat []uint64) {
	self.rout.advocateRoutineInfo.oat = oat
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
//   - op ...Operations: opertion types waiting on, only multiple in select
func StorePark(p unsafe.Pointer, skip int, replay bool, op Operation, id uint64) {
	parkObj := park{
		addr: p,
		id:   id,
		op:   op,
	}

	currentGoRoutineInfo().parkObj = []park{parkObj}
	currentGoRoutineInfo().parkPos = posFromCaller(skip)
	currentGoRoutineInfo().parkForeverReplay = replay
}

// StorePark stores in a routine, a pointers to the channels involved in a
// select on which a routine parked.
// Do not call if the select has a default.
//
// Parameter:
//   - cas0: cases of the select
//   - nsends: number of send cases
//   - ncases: total number of non default cases
//   - skip int: caller skip
func StoreParkSelect(cas0 *scase, nsends int, ncases int, skip int) {
	cas1 := (*[1 << 16]scase)(unsafe.Pointer(cas0))
	scases := cas1[:ncases:ncases]

	currentGoRoutineInfo().parkObj = make([]park, 0)

	currentGoRoutineInfo().parkPos = posFromCaller(skip)

	for casi := 0; casi < ncases; casi++ {
		cas := &scases[casi]
		c := cas.c

		if c == nil { // ignore nil cases
			continue
		}

		chanOp := OperationChannelRecv
		if casi < nsends {
			chanOp = OperationChannelSend
		}

		parkObj := park{
			addr: unsafe.Pointer(cas.c),
			id:   c.id,
			op:   chanOp,
		}

		currentGoRoutineInfo().parkObj = append(currentGoRoutineInfo().parkObj, parkObj)
	}
}

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

func (gp *AdvocateG) ReadyStatusWaiting() bool {
	return readgstatus(gp.rout) == _Gwaiting
}

// isRoutineWaitingOnConcurrency determines if an advocate routine is blocked on a concurreny operations
// Returns:
//   - bool: true blocked on concurrency op, false otherwise
func (gp *AdvocateG) isRoutineWaitingOnConcurrency() bool {
	if !gp.ReadyStatusWaiting() {
		return false
	}

	if !IsInSlice(blockedConcurrencyReasons, gp.getWaitReason()) {
		return false
	}

	return true
}
