package main

import (
	"sync"
	"testing"
)

func TestBasicLooped(t *testing.T) {
	var x, y sync.Mutex
	repeat := 3

	go func() {
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
	}()

	for i := 0; i < repeat; i++ {
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
	}
}
