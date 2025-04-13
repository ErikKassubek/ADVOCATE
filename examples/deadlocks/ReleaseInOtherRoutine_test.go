package main

import (
	"sync"
	"testing"
)

func TestReleaseInOtherRoutine(t *testing.T) {
	var x, y sync.Mutex

	go func() {
		y.Lock()
		x.Lock()
		x.Unlock()
	}()

	go func() {
		y.Unlock()
	}()

	x.Lock()
	y.Lock()
	x.Unlock()
}
