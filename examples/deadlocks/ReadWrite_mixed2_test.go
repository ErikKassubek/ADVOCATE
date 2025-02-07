package main

import (
	"sync"
	"testing"
)

func TestReadWriteMixed2(t *testing.T) {
	var x, y sync.RWMutex

	go func() {
		x.RLock()
		y.RLock()
		y.RUnlock()
		x.RUnlock()
	}()

	go func() {
		y.Lock()
		x.Lock() // SHOULD create a deadlock
		x.Unlock()
		y.Unlock()
	}()
}
