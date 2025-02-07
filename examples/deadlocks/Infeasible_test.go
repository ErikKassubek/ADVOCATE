package main

import (
	"sync"
	"testing"
)

func TestInfeasible(t *testing.T) {
	var x, y sync.Mutex

	x.Lock()
	y.Lock()
	y.Unlock()
	x.Unlock()

	go func() {
		y.Lock()
		x.Lock() // This SHOULD NOT produce a deadlock since the other critical section has already been executed
		x.Unlock()
		y.Unlock()
	}()
}
