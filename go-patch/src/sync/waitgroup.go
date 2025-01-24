// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"internal/race"
	// ADVOCATE-CHANGE-START
	"runtime"
	// ADVOCATE-CHANGE-END
	"sync/atomic"
	"unsafe"
)

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for. Then each of the goroutines
// runs and calls Done when finished. At the same time,
// Wait can be used to block until all goroutines have finished.
//
// A WaitGroup must not be copied after first use.
//
// In the terminology of the Go memory model, a call to Done
// “synchronizes before” the return of any Wait call that it unblocks.
type WaitGroup struct {
	noCopy noCopy

	state atomic.Uint64 // high 32 bits are counter, low 32 bits are waiter count.
	sema  uint32

	// ADVOCATE-CHANGE-START
	id uint64 // id for the waitgroup
	// ADVOCATE-CHANGE-END
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
//
// Note that calls with a positive delta that occur when the counter is zero
// must happen before a Wait. Calls with a negative delta, or calls with a
// positive delta that start when the counter is greater than zero, may happen
// at any time.
// Typically this means the calls to Add should execute before the statement
// creating the goroutine or other event to be waited for.
// If a WaitGroup is reused to wait for several independent sets of events,
// new Add calls must happen after all previous Wait calls have returned.
// See the WaitGroup example.
func (wg *WaitGroup) Add(delta int) {
	// ADVOCATE-CHANGE-START
	skip := 3
	if delta > 0 {
		skip = 2
	}
	wait, ch, chAck := runtime.WaitForReplay(runtime.OperationWaitgroupAddDone, skip, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		<-ch
	}
	// ADVOCATE-CHANGE-END

	if race.Enabled {
		if delta < 0 {
			// Synchronize decrements with Wait.
			race.ReleaseMerge(unsafe.Pointer(wg))
		}
		race.Disable()
		defer race.Enable()
	}
	state := wg.state.Add(uint64(delta) << 32)
	v := int32(state >> 32)
	w := uint32(state)

	// ADVOCATE-CHANGE-START
	// Waitgroups don't need to be initialized in default go code. Because
	// go does not have constructors, the only way to initialize a wg
	// is directly in it's functions. If the id of the wg is the default
	// value, it is set to a new, unique object id
	if wg.id == 0 {
		wg.id = runtime.GetAdvocateObjectID()
	}
	// Record the add or done of a wait group in the routine's trace.
	// If delta > 0, it is an add, if it's -1, it's a done.
	// The add or done cannot fait without crashing the program. Add and done
	// do not block the program. Therefore it is not possible, that it is
	// called but not finished (except if it panics). Therefore it is not
	// necessary to record a post event.
	runtime.AdvocateWaitGroupAdd(wg.id, delta, v)
	// ADVOCATE-CHANGE-END

	if race.Enabled && delta > 0 && v == int32(delta) {
		// The first increment must be synchronized with Wait.
		// Need to model this as a read, because there can be
		// several concurrent wg.counter transitions from 0.
		race.Read(unsafe.Pointer(&wg.sema))
	}
	if v < 0 {
		panic("sync: negative WaitGroup counter")
	}
	if w != 0 && delta > 0 && v == int32(delta) {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	if v > 0 || w == 0 {
		return
	}
	// This goroutine has set counter to 0 when waiters > 0.
	// Now there can't be concurrent mutations of state:
	// - Adds must not happen concurrently with Wait,
	// - Wait does not increment waiters if it sees counter == 0.
	// Still do a cheap sanity check to detect WaitGroup misuse.
	if wg.state.Load() != state {
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	// Reset waiters count to 0.
	wg.state.Store(0)
	for ; w != 0; w-- {
		runtime_Semrelease(&wg.sema, false, 0)
	}
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	// ADVOCATE-CHANGE-START
	wait, ch, chAck := runtime.WaitForReplay(runtime.OperationWaitgroupWait, 2, true)
	if wait {
		defer func() { chAck <- struct{}{} }()
		replayElem := <-ch
		if replayElem.Blocked {
			if wg.id == 0 {
				wg.id = runtime.GetAdvocateObjectID()
			}
			_ = runtime.AdvocateWaitGroupWaitPre(wg.id)
			runtime.BlockForever()
		}
	}
	// ADVOCATE-CHANGE-END

	if race.Enabled {
		race.Disable()
	}

	// ADVOCATE-CHANGE-START
	// Waitgroups don't need to be initialized in default go code. Because
	// go does not have constructors, the only way to initialize a wg
	// is directly in it's functions. If the id of the wg is the default
	// value, it is set to a new, unique object id
	if wg.id == 0 {
		wg.id = runtime.GetAdvocateObjectID()
	}

	// Record the wait of a wait group in the routine's trace.
	// The wait will run until the waitgroup counte is zero. Therefor it
	// blocks the routine and it is nessesary to record the successful
	// finish of the wait with a post.
	advocateIndex := runtime.AdvocateWaitGroupWaitPre(wg.id)
	// ADVOCATE-CHANGE-END

	for {
		state := wg.state.Load()
		v := int32(state >> 32)
		w := uint32(state)
		if v == 0 {
			// Counter is 0, no need to wait.
			if race.Enabled {
				race.Enable()
				race.Acquire(unsafe.Pointer(wg))
			}
			// ADVOCATE-CHANGE-START
			runtime.AdvocateWaitGroupPost(advocateIndex)
			//ADVOCATE-CHANGE-END
			return
		}
		// Increment waiters count.
		if wg.state.CompareAndSwap(state, state+1) {
			if race.Enabled && w == 0 {
				// Wait must be synchronized with the first Add.
				// Need to model this is as a write to race with the read in Add.
				// As a consequence, can do the write only for the first waiter,
				// otherwise concurrent Waits will race with each other.
				race.Write(unsafe.Pointer(&wg.sema))
			}
			runtime_Semacquire(&wg.sema)
			if wg.state.Load() != 0 {
				panic("sync: WaitGroup is reused before previous Wait has returned")
			}
			if race.Enabled {
				race.Enable()
				race.Acquire(unsafe.Pointer(wg))
			}
			// ADVOCATE-CHANGE-START
			runtime.AdvocateWaitGroupPost(advocateIndex)
			//ADVOCATE-CHANGE-END
			return
		}
	}
}
