package main

import (
	"sync"
	"testing"
)

func TestReadWriteReadonly(t *testing.T) {
	var x, y sync.RWMutex

	go func() {
		x.RLock() // This SHOULD NOT produce a deadlock
		y.RLock()
		y.RUnlock()
		x.RUnlock()
	}()

	y.RLock()
	x.RLock()
	x.RUnlock()
	y.RUnlock()
}
