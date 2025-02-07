package main

import (
	"sync"
	"testing"
)

func TestBasic(t *testing.T) {
	var x, y sync.Mutex

	go func() {
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
	}()

	go func() {
		y.Lock()
		x.Lock() // this SHOULD produce a deadlock
		x.Unlock()
		y.Unlock()
	}()
}
