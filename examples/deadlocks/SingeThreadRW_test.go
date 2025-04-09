package main

import (
	"sync"
	"testing"
)

func TestSingeThreadRW(t *testing.T) {
	var x sync.RWMutex

	x.Lock()
	x.RLock() // Deadlock
	x.RUnlock()
	x.Unlock()
}
