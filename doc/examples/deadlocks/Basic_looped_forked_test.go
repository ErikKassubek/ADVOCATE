package main

import (
	"sync"
	"testing"
	"time"
)

func TestBasicLoopedForked(t *testing.T) {
	var x, y sync.Mutex
	repeat := 3

	for i := 0; i < repeat; i++ {
		go func() {
			x.Lock()
			y.Lock()
			y.Unlock()
			x.Unlock()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	y.Lock()
	x.Lock()
	x.Unlock()
	y.Unlock()
}
