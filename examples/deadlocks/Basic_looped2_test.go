package main

import (
	"sync"
	"testing"
)

func TestBasicLooped2(t *testing.T) {
	var x, y sync.Mutex
	repeat := 3

	go func() {
		for i := 0; i < repeat; i++ {
			x.Lock()
			y.Lock()
			y.Unlock()
			x.Unlock()
		}
	}()

	go func() {
		for i := 0; i < repeat; i++ {
			y.Lock()
			x.Lock()
			x.Unlock()
			y.Unlock()
		}
	}()
}
