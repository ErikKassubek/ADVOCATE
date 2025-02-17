package main

import (
	"advocate"
	"sync"
	"testing"
)

func TestReadWriteMultiInst(t *testing.T) {
	// ======= Preamble Start =======
  advocate.InitReplay("3", false, 20, true)
  defer advocate.FinishReplay()
  // ======= Preamble End =======
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
