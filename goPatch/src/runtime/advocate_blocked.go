// Copyright (c) 2026 Erik Kassubek
//
// File: advocate_partial_deadlock.go
// Brief: Detect partial deadlocks while running
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package runtime

var blockedConcurrencyReasons = []waitReason{
	waitReasonChanReceiveNilChan,
	waitReasonChanSendNilChan,
	waitReasonSelect,
	waitReasonSelectNoCases,
	waitReasonChanReceive,
	waitReasonChanSend,
	waitReasonSyncCondWait,
	waitReasonSyncMutexLock,
	waitReasonSyncRWMutexRLock,
	waitReasonSyncRWMutexLock,
	waitReasonSyncWaitGroupWait,
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
func GetWaitingReasonString(wr waitReason) string {
	switch wr {
	case waitReasonChanReceiveNilChan:
		return "chan:recvOnNil"
	case waitReasonChanSendNilChan:
		return "chan:sendOnNil"
	case waitReasonSelect:
		return "select:select"
	case waitReasonSelectNoCases:
		return "select:withoutCases"
	case waitReasonChanReceive:
		return "chan:revc"
	case waitReasonChanSend:
		return "chan:send"
	case waitReasonSyncCondWait:
		return "cond:wait"
	case waitReasonSyncMutexLock:
		return "mutex:lock"
	case waitReasonSyncRWMutexRLock:
		return "rwmutex:rlock"
	case waitReasonSyncRWMutexLock:
		return "rwmutex:lock"
	case waitReasonSyncWaitGroupWait:
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
