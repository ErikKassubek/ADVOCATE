package main

import (
	"sync/atomic"
	"time"
)

// The following code is an example in which GoPie is helpful.
// In it, whether the code part containing the bug is triggered, depends
// on the order of the atomic store and load.
// If in the program, the Load is executed in most cases before the store
// (simulated by sleep), the code part containing the bug is not triggered,
// and no dynamic analysis will be able to detect the bug.
// Be reordering the store and load, such that the store is executed before
// the load, the code can be detected.

func codeWithBug() {
	e := make(chan int)
	close(e)
	e <- 1
}

func main() {
	a := atomic.Int32{}

	go func() {
		// some code
		time.Sleep(400 * time.Millisecond)

		a.Store(1)
	}()

	// some code
	time.Sleep(200 * time.Millisecond)

	if a.Load() == 1 {
		codeWithBug()
	}
}
