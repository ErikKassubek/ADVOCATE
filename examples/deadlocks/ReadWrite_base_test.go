package main

import (
	"sync"
	"testing"
)

func TestReadWriteBase(t *testing.T) {
	var x, y sync.RWMutex

	go func() { // This SHOULD produce a deadlock
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
	}()

	go func() {
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
	}()
}
