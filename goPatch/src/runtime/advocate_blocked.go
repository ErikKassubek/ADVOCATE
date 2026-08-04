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

func (self *AdvocateG) isNil() bool {
	return self.rout.advocateRoutineInfo == nil
}

func (self *AdvocateG) id() uint64 {
	return self.rout.advocateRoutineInfo.id
}

func (self *AdvocateG) goId() uint64 {
	return self.rout.goid
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

// isRoutineWaitingOnConcurrency determines if an advocate routine is blocked on a concurreny operations
// Returns:
//   - bool: true blocked on concurrency op, false otherwise
func (self *AdvocateG) LastTraceElem() traceElem {
	if self.isNil() {
		return nil
	}

	l := len(self.rout.advocateRoutineInfo.Trace)

	if l == 0 {
		return nil
	}

	return self.rout.advocateRoutineInfo.Trace[l-1]
}
