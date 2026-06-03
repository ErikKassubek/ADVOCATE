package main

import (
	"sync"
	"testing"
)

func TestReadWriteGuard(t *testing.T) {
	var x, y sync.Mutex
	var guard sync.RWMutex

	go func() {
		guard.RLock() // This SHOULD produce a deadlock since R Locks do not block each other so the guard does not work
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
		guard.RUnlock()
	}()

	guard.RLock()
	y.Lock()
	x.Lock()
	x.Unlock()
	y.Unlock()
	guard.RUnlock()

}
