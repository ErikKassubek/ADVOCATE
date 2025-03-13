// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sync provides basic synchronization primitives such as mutual
// exclusion locks. Other than the [Once] and [WaitGroup] types, most are intended
// for use by low-level library routines. Higher-level synchronization is
// better done via channels and communication.
//
// Values containing the types defined in this package should not be copied.
package sync

import (
	isync "internal/sync"

	// ADVOCATE-START
	"runtime"
	// ADVOCATE-END
)

// A Mutex is a mutual exclusion lock.
// The zero value for a Mutex is an unlocked mutex.
//
// A Mutex must not be copied after first use.
//
// In the terminology of [the Go memory model],
// the n'th call to [Mutex.Unlock] “synchronizes before” the m'th call to [Mutex.Lock]
// for any n < m.
// A successful call to [Mutex.TryLock] is equivalent to a call to Lock.
// A failed call to TryLock does not establish any “synchronizes before”
// relation at all.
//
// [the Go memory model]: https://go.dev/ref/mem
type Mutex struct {
	_ noCopy

	mu isync.Mutex

	// ADVOCATE-START
	id uint64
	// ADVOCATE-END
}

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
	Lock()
	Unlock()
}

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *Mutex) Lock() {
	// ADVOCATE-START
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
	// ADVOCATE-END

	m.mu.Lock()

	// ADVOCATE-START
	runtime.AdvocateMutexPost(advocateIndex)
	//ADVOCATE-END
}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (m *Mutex) TryLock() bool {
	// ADVOCATE-START
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

	res := m.mu.TryLock()

	runtime.AdvocatePostTry(advocateIndex, res)

	return res
	// ADVOCATE-END
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
//
// A locked [Mutex] is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *Mutex) Unlock() {
	// ADVOCATE-START
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
	// ADVOCATE-END

	m.mu.Unlock()

	// ADVOCATE-START
	runtime.AdvocateMutexPost(advocateIndex)
	// ADVOCATE-END
}
