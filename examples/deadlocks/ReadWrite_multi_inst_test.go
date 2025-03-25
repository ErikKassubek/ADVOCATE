package main

import (
	"sync"
	"testing"
)

func TestReadWriteMultiInst(t *testing.T) {
	var x, y sync.RWMutex

	go func() {
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
		y.RLock()
		x.Lock()
		x.Unlock()
		y.RUnlock()
	}()

	x.Lock()
	y.Lock()
	y.Unlock()
	x.Unlock()
}
