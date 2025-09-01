package main

import (
	"sync"
	"testing"
)

func TestBasicLoopedSynced(t *testing.T) {
	var x, y sync.Mutex
	repeat := 3
	c := make(chan int)

	go func() {
		for i := 0; i < repeat; i++ {
			<-c
			x.Lock()
			y.Lock()
			y.Unlock()
			x.Unlock()
		}
	}()

	for i := 0; i < repeat; i++ {
		y.Lock()
		x.Lock()
		x.Unlock()
		y.Unlock()
		c <- 1
	}
}
