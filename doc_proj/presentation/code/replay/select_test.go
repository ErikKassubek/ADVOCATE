package main

import (
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	c := make(chan int)

	go func() {
		c <- 1
	}()

	select {
	case <-c:
		println("Case 1")
	case <-c:
		println("Case 2")
	}

	time.Sleep(time.Second)
}
