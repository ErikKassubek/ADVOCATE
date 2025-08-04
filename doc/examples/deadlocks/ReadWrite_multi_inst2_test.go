package main

import (
	"sync"
	"testing"
)

func TestReadWriteMultiInst2(t *testing.T) {
	var x, y sync.RWMutex

	go func() {
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
		y.Lock()
		x.RLock()
		x.RUnlock()
		y.Unlock()
	}()

	x.Lock()
	y.Lock()
	y.Unlock()
	x.Unlock()
}
