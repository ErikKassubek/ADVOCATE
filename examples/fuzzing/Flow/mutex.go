package main

import (
	"sync"
	"testing"
	"time"
)

// The following code is an example in which Flow fuzzing is helpful.
// It consists of a mutex on which a TryLock is called. A TryLock only
// acquires a mutex, if the mutex is not locked already.
// In the given example, the TryLock operations is in most cases not successful
// since, given by the most likely execution times, it will be run while
// the mutex is held by the Lock. But it is still possible for the TryLock
// to be successful.
// If during a dynamic analysis, the TryLock is not, or never, successful,
// the analysis will never be able to execute the relevant code and detect
// the bug.
// By delaying the execution of the Lock operations, we may enable the TryLock
// to successfully acquire the mutex, making it possible for the analysis
// to detect the hidden bug.

func TestMutex(_ *testing.T) {
	m := sync.Mutex{}

	go func() {
		// some code
		time.Sleep(100 * time.Second)

		res := m.TryLock()
		if res {
			panic("CODE WITH PANIC")

			m.Unlock()
		}
	}()

	m.Lock()

	// some code
	time.Sleep(300 * time.Millisecond)

	m.Unlock()

}
