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
	"unsafe"

	// OOSC-START
	"runtime"
	// OOSC-END
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

	// OOSC-START
	id uint64 // id for the mutex
	// OOSC-END
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
	// OOSC-START
	wait, ch, chAck, _ := runtime.WaitForReplay(runtime.OperationMutexLock, runtime.CallerSkipMutex, false)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if m.id == 0 {
			m.id = runtime.GetOoscObjectID()
		}
		if replayElem.Blocked {
			_ = runtime.OoscMutexPre(m.id, runtime.OperationMutexLock)
			runtime.StorePark(unsafe.Pointer(m), runtime.CallerSkipMutex, true)
			runtime.BlockForever()
		}
	}

	runtime.FuzzingFlowWait(runtime.CallerSkipMutex)

	// Mutexe don't need to be initialized in default go code. Because
	// go does not have constructors, the only way to initialize a mutex
	// is directly in the lock function. If the id of the channel is the default
	// value, it is set to a new, unique object id.
	if m.id == 0 {
		m.id = runtime.GetOoscObjectID()
	}

	// OoscMutexPre records, that a routine tries to lock a mutex.
	// OoscPost is called, if the mutex was locked successfully.
	// In this case, the Lock event in the trace is updated to include
	// this information. ooscIndex is used for OoscPost to find the
	// pre event.
	ooscIndex := runtime.OoscMutexPre(m.id, runtime.OperationMutexLock)
	// OOSC-END

	// OOSC-START
	runtime.StorePark(unsafe.Pointer(m), runtime.CallerSkipMutex, false)
	// OOSC-END

	m.mu.Lock()

	// OOSC-START
	runtime.OoscMutexPost(ooscIndex, true)
	//OOSC-END
}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (m *Mutex) TryLock() bool {
	// OOSC-START
	wait, ch, chAck, _ := runtime.WaitForReplay(runtime.OperationMutexTryLock, runtime.CallerSkipMutex, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if replayElem.Blocked {
			if m.id == 0 {
				m.id = runtime.GetOoscObjectID()
			}
			_ = runtime.OoscMutexPre(m.id, runtime.OperationMutexTryLock)
			runtime.StorePark(unsafe.Pointer(m), runtime.CallerSkipMutex, true)
			runtime.BlockForever()
		}
	}

	runtime.FuzzingFlowWait(runtime.CallerSkipMutex)

	// Mutexe don't need to be initialized in default go code. Because
	// go does not have constructors, the only way to initialize a mutex
	// is directly in the lock function. If the id of the channel is the default
	// value, it is set to a new, unique object id
	if m.id == 0 {
		m.id = runtime.GetOoscObjectID()
	}

	// OoscMutexPre records, that a routine tries to lock a mutex.
	// ooscIndex is used for OoscMutexPost to find the pre event.
	ooscIndex := runtime.OoscMutexPre(m.id, runtime.OperationMutexTryLock)
	// OOSC-END

	res := m.mu.TryLock()

	runtime.OoscMutexPost(ooscIndex, res)

	return res
	// OOSC-END
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
//
// A locked [Mutex] is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *Mutex) Unlock() {
	// OOSC-START
	wait, ch, chAck, _ := runtime.WaitForReplay(runtime.OperationMutexUnlock, runtime.CallerSkipMutex, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if replayElem.Blocked {
			if m.id == 0 {
				m.id = runtime.GetOoscObjectID()
			}
			_ = runtime.OoscMutexPre(m.id, runtime.OperationMutexUnlock)
			runtime.StorePark(unsafe.Pointer(m), runtime.CallerSkipMutex, true)
			runtime.BlockForever()
		}
	}
	// OoscMutexPre is used to record the unlocking of a mutex.
	// OoscPost records the successful unlocking of a mutex.
	// For non rw mutexe, the unlock cannot fail. Therefore it is not
	// strictly necessary to record the post for the unlocking of a mutex.
	// For rw mutexes, the unlock can fail (e.g. unlock after rlock). Therefore
	// in this case it is nessesary to record the post for the unlocking of an
	// rw mutex.
	// Here the post is seperatly recorded to easy the implementation for
	// the rw mutexes.
	ooscIndex := runtime.OoscMutexPre(m.id, runtime.OperationMutexUnlock)
	// OOSC-END

	m.mu.Unlock()

	// OOSC-START
	runtime.OoscMutexPost(ooscIndex, true)
	// OOSC-END
}
