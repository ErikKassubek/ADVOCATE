package main

import (
	"sync"
	"testing"
)

func TestGuard(t *testing.T) {
	var x, y sync.Mutex
	var guard sync.Mutex

	go func() {
		guard.Lock() // This SHOULD NOT produce a deadlock
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
		guard.Unlock()
	}()

	go func() {
		guard.Lock()
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
		guard.Unlock()
	}()
}
