package main

import (
	"sync"
	"testing"
)

func TestTryLock(t *testing.T) {
	var x, y sync.Mutex

	go func() {
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
	}()

	go func() {
		y.Lock()
		println("TryLock succeeded:", x.TryLock())
		x.Unlock()
		y.Unlock()
	}()
}
