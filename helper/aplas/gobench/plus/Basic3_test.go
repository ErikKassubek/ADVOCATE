package main

import (
	"sync"
	"testing"
)

func TestBasic3(t *testing.T) {
	var x, y, z sync.Mutex

	go func() {
		x.Lock()
		y.Lock()
		y.Unlock()
		x.Unlock()
	}()

	go func() {
		y.Lock()
		z.Lock()
		z.Unlock()
		y.Unlock()
	}()

	z.Lock()
	x.Lock() // this SHOULD produce a deadlock
	x.Unlock()
	z.Unlock()

}
