package main

import (
	"sync"
	"testing"
)

func TestInInfeasible(t *testing.T) {
	var x, y sync.Mutex

	x.Lock()
	y.Lock()
	y.Unlock()
	x.Unlock()

	go func() {
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
	}()

	x.Lock()
	y.Lock() // This SHOULD produce a deadlock since we now have two concurrent routines
	y.Unlock()
	x.Unlock()
}
