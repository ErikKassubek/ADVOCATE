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

	y.Lock()
	var suc = x.TryLock()
	println("TryLock succeeded:", suc)
	if suc {
		x.Unlock()
	}
	y.Unlock()
}
