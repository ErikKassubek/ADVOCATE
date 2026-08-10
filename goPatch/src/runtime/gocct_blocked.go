// Copyright (c) 2025 Erik Kassubek
//
// File: gocct_partial_deadlock.go
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

type GoCCTG struct {
	rout *g
}

func (self *GoCCTG) parkForeverReplay() bool {
	return self.rout.gocctRoutineInfo.parkForeverReplay
}

func (self *GoCCTG) isNil() bool {
	return self.rout.gocctRoutineInfo == nil
}

func (self *GoCCTG) id() uint64 {
	return self.rout.gocctRoutineInfo.id
}

func (self *GoCCTG) goId() uint64 {
	return self.rout.goid
}

func (self *GoCCTG) getForkPos() string {
	return self.rout.gocctRoutineInfo.GetForkPos()
}

func (self *GoCCTG) setOAT(oat []uint64) {
	self.rout.gocctRoutineInfo.oat = oat
}

func ForEachGoCCTG(fn func(adGp *GoCCTG)) {
	forEachG(func(gp *g) {
		fn(&GoCCTG{gp})
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

// isRoutineWaitingOnConcurrency determines if an gocct routine is blocked on a concurreny operations
// Returns:
//   - bool: true blocked on concurrency op, false otherwise
func (self *GoCCTG) LastTraceElem() traceElem {
	if self.isNil() {
		return nil
	}

	l := len(self.rout.gocctRoutineInfo.Trace)

	if l == 0 {
		return nil
	}

	return self.rout.gocctRoutineInfo.Trace[l-1]
}
