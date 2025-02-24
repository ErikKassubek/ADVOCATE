// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sync provides basic synchronization primitives such as mutual
// exclusion locks. Other than the Once and WaitGroup types, most are intended
// for use by low-level library routines. Higher-level synchronization is
// better done via channels and communication.
//
// Values containing the types defined in this package should not be copied.
package sync

import (
	"internal/race"
	// ADVOCATE-CHANGE-START
	"runtime"
	// ADVOCATE-CHANGE-END
	"sync/atomic"
	"unsafe"
)

// Provided by runtime via linkname.
func throw(string)
func fatal(string)

// A Mutex is a mutual exclusion lock.
// The zero value for a Mutex is an unlocked mutex.
//
// A Mutex must not be copied after first use.
//
// In the terminology of the Go memory model,
// the n'th call to Unlock “synchronizes before” the m'th call to Lock
// for any n < m.
// A successful call to TryLock is equivalent to a call to Lock.
// A failed call to TryLock does not establish any “synchronizes before”
// relation at all.
type Mutex struct {
	state int32
	sema  uint32

	// ADVOCATE-CHANGE-START
	id uint64 // id for the mutex
	// ADVOCATE-CHANGE-END
}

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
	Lock()
	Unlock()
}

const (
	mutexLocked = 1 << iota // mutex is locked
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota

	// Mutex fairness.
	//
	// Mutex can be in 2 modes of operations: normal and starvation.
	// In normal mode waiters are queued in FIFO order, but a woken up waiter
	// does not own the mutex and competes with new arriving goroutines over
	// the ownership. New arriving goroutines have an advantage -- they are
	// already running on CPU and there can be lots of them, so a woken up
	// waiter has good chances of losing. In such case it is queued at front
	// of the wait queue. If a waiter fails to acquire the mutex for more than 1ms,
	// it switches mutex to the starvation mode.
	//
	// In starvation mode ownership of the mutex is directly handed off from
	// the unlocking goroutine to the waiter at the front of the queue.
	// New arriving goroutines don't try to acquire the mutex even if it appears
	// to be unlocked, and don't try to spin. Instead they queue themselves at
	// the tail of the wait queue.
	//
	// If a waiter receives ownership of the mutex and sees that either
	// (1) it is the last waiter in the queue, or (2) it waited for less than 1 ms,
	// it switches mutex back to normal operation mode.
	//
	// Normal mode has considerably better performance as a goroutine can acquire
	// a mutex several times in a row even if there are blocked waiters.
	// Starvation mode is important to prevent pathological cases of tail latency.
	starvationThresholdNs = 1e6
)

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *Mutex) Lock() {
	// ADVOCATE-CHANGE-START
	wait, ch, chAck := runtime.WaitForReplay(runtime.OperationMutexLock, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if m.id == 0 {
			m.id = runtime.GetAdvocateObjectID()
		}
		if replayElem.Blocked {
			_ = runtime.AdvocateMutexLockPre(m.id, false, false)
			runtime.BlockForever()
		}
	}

	runtime.FuzzingFlowWait(2)

	// Mutexe don't need to be initialized in default go code. Because
	// go does not have constructors, the only way to initialize a mutex
	// is directly in the lock function. If the id of the channel is the default
	// value, it is set to a new, unique object id.
	if m.id == 0 {
		m.id = runtime.GetAdvocateObjectID()
	}

	// AdvocateMutexLockPre records, that a routine tries to lock a mutex.
	// AdvocatePost is called, if the mutex was locked successfully.
	// In this case, the Lock event in the trace is updated to include
	// this information. advocateIndex is used for AdvocatePost to find the
	// pre event.
	advocateIndex := runtime.AdvocateMutexLockPre(m.id, false, false)
	// ADVOCATE-CHANGE-END

	// Fast path: grab unlocked mutex.
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		// ADVOCATE-CHANGE-START
		runtime.AdvocateMutexPost(advocateIndex)
		//ADVOCATE-CHANGE-END
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		return
	}
	// Slow path (outlined so that the fast path can be inlined)
	m.lockSlow(advocateIndex)
}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (m *Mutex) TryLock() bool {
	// ADVOCATE-CHANGE-START
	wait, ch, chAck := runtime.WaitForReplay(runtime.OperationMutexTryLock, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if replayElem.Blocked {
			if m.id == 0 {
				m.id = runtime.GetAdvocateObjectID()
			}
			_ = runtime.AdvocateMutexLockTry(m.id, false, false)
			runtime.BlockForever()
		}
		// if !replayElem.Suc {
		// 	if m.id == 0 {
		// 		m.id = runtime.GetAdvocateObjectID()
		// 	}
		// 	advocateIndex := runtime.AdvocateMutexLockTry(m.id, false, false)
		// 	runtime.AdvocatePostTry(advocateIndex, false)
		// 	return false
		// }
	}

	runtime.FuzzingFlowWait(2)

	// Mutexe don't need to be initialized in default go code. Because
	// go does not have constructors, the only way to initialize a mutex
	// is directly in the lock function. If the id of the channel is the default
	// value, it is set to a new, unique object id
	if m.id == 0 {
		m.id = runtime.GetAdvocateObjectID()
	}

	// AdvocateMutexLockPre records, that a routine tries to lock a mutex.
	// advocateIndex is used for AdvocatePostTry to find the pre event.
	advocateIndex := runtime.AdvocateMutexLockTry(m.id, false, false)
	// ADVOCATE-CHANGE-END

	old := m.state
	if old&(mutexLocked|mutexStarving) != 0 {
		// ADVOCATE-CHANGE-START
		runtime.AdvocatePostTry(advocateIndex, false)
		// ADVOCATE-CHANGE-END
		return false
	}

	// There may be a goroutine waiting for the mutex, but we are
	// running now and can try to grab the mutex before that
	// goroutine wakes up.
	if !atomic.CompareAndSwapInt32(&m.state, old, old|mutexLocked) {
		// ADVOCATE-CHANGE-START
		// If the mutex was not locked successfully, AdvocatePostTry is called
		// to update the trace.
		runtime.AdvocatePostTry(advocateIndex, false)
		// ADVOCATE-CHANGE-END
		return false
	}

	if race.Enabled {
		race.Acquire(unsafe.Pointer(m))
	}
	// ADVOCATE-CHANGE-START
	// If the mutex was locked successfully, AdvocatePostTry is called
	// to update the trace.
	runtime.AdvocatePostTry(advocateIndex, true)
	// ADVOCATE-CHANGE-END
	return true
}

// ADVOCATE-CHANGE-START
func (m *Mutex) lockSlow(advocateIndex int) {
	// ADVOCATE-CHANGE-END
	var waitStartTime int64
	starving := false
	awoke := false
	iter := 0
	old := m.state
	for {
		// Don't spin in starvation mode, ownership is handed off to waiters
		// so we won't be able to acquire the mutex anyway.
		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
			// Active spinning makes sense.
			// Try to set mutexWoken flag to inform Unlock
			// to not wake other blocked goroutines.
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
			runtime_doSpin()
			iter++
			old = m.state
			continue
		}
		new := old
		// Don't try to acquire starving mutex, new arriving goroutines must queue.
		if old&mutexStarving == 0 {
			new |= mutexLocked
		}
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}
		// The current goroutine switches mutex to starvation mode.
		// But if the mutex is currently unlocked, don't do the switch.
		// Unlock expects that starving mutex has waiters, which will not
		// be true in this case.
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}
		if awoke {
			// The goroutine has been woken from sleep,
			// so we need to reset the flag in either case.
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
			if old&(mutexLocked|mutexStarving) == 0 {
				break // locked the mutex with CAS
			}
			// If we were already waiting before, queue at the front of the queue.
			queueLifo := waitStartTime != 0
			if waitStartTime == 0 {
				waitStartTime = runtime_nanotime()
			}
			runtime_SemacquireMutex(&m.sema, queueLifo, 1)
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
			old = m.state
			if old&mutexStarving != 0 {
				// If this goroutine was woken and mutex is in starvation mode,
				// ownership was handed off to us but mutex is in somewhat
				// inconsistent state: mutexLocked is not set and we are still
				// accounted as waiter. Fix that.
				if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
					throw("sync: inconsistent mutex state")
				}
				delta := int32(mutexLocked - 1<<mutexWaiterShift)
				if !starving || old>>mutexWaiterShift == 1 {
					// Exit starvation mode.
					// Critical to do it here and consider wait time.
					// Starvation mode is so inefficient, that two goroutines
					// can go lock-step infinitely once they switch mutex
					// to starvation mode.
					delta -= mutexStarving
				}
				atomic.AddInt32(&m.state, delta)
				break
			}
			awoke = true
			iter = 0
		} else {
			old = m.state
		}
	}

	// ADVOCATE-CHANGE-START
	runtime.AdvocateMutexPost(advocateIndex)
	//ADVOCATE-CHANGE-END
	if race.Enabled {
		race.Acquire(unsafe.Pointer(m))
	}
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *Mutex) Unlock() {
	// ADVOCATE-CHANGE-START
	wait, ch, chAck := runtime.WaitForReplay(runtime.OperationMutexUnlock, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if replayElem.Blocked {
			if m.id == 0 {
				m.id = runtime.GetAdvocateObjectID()
			}
			_ = runtime.AdvocateUnlockPre(m.id, false, false)
			runtime.BlockForever()
		}
	}
	// AdvocateUnlockPre is used to record the unlocking of a mutex.
	// AdvocatePost records the successful unlocking of a mutex.
	// For non rw mutexe, the unlock cannot fail. Therefore it is not
	// strictly necessary to record the post for the unlocking of a mutex.
	// For rw mutexes, the unlock can fail (e.g. unlock after rlock). Therefore
	// in this case it is nessesary to record the post for the unlocking of an
	// rw mutex.
	// Here the post is seperatly recorded to easy the implementation for
	// the rw mutexes.
	advocateIndex := runtime.AdvocateUnlockPre(m.id, false, false)
	runtime.AdvocateMutexPost(advocateIndex)
	// ADVOCATE-CHANGE-END

	if race.Enabled {
		_ = m.state
		race.Release(unsafe.Pointer(m))
	}

	// Fast path: drop lock bit.
	new := atomic.AddInt32(&m.state, -mutexLocked)
	if new != 0 {
		// Outlined slow path to allow inlining the fast path.
		// To hide unlockSlow during tracing we skip one extra frame when tracing GoUnblock.
		m.unlockSlow(new)
	}
}

func (m *Mutex) unlockSlow(new int32) {
	if (new+mutexLocked)&mutexLocked == 0 {
		// ADVOCATE-CHANGE-START
		fatal("sync: unlock of unlocked mutex")
		// ADVOCATE-CHANGE-END
	}
	if new&mutexStarving == 0 {
		old := new
		for {
			// If there are no waiters or a goroutine has already
			// been woken or grabbed the lock, no need to wake anyone.
			// In starvation mode ownership is directly handed off from unlocking
			// goroutine to the next waiter. We are not part of this chain,
			// since we did not observe mutexStarving when we unlocked the mutex above.
			// So get off the way.
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return
			}
			// Grab the right to wake someone.
			new = (old - 1<<mutexWaiterShift) | mutexWoken
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
				runtime_Semrelease(&m.sema, false, 1)
				return
			}
			old = m.state
		}
	} else {
		// Starving mode: handoff mutex ownership to the next waiter, and yield
		// our time slice so that the next waiter can start to run immediately.
		// Note: mutexLocked is not set, the waiter will set it after wakeup.
		// But mutex is still considered locked if mutexStarving is set,
		// so new coming goroutines won't acquire it.
		runtime_Semrelease(&m.sema, true, 1)
	}
}
