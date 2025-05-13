package main

import (
	"testing"
	"time"
)

// The following code is an example in which GFuzz is helpful.
// It consists of a select with two possible path.
// One of the paths (c) contains no bug.
// The path (d) calls code which contains a send on closed channel bug.
// Both cases of the select have a possible communication partner, that could
// trigger the case.
// Let's assume, that it is very likely, but not guaranteed, that the send
// on c is before the send on d (simulated with sleep of different length).
// In a dynamic analysis, we can only analyze program parts that are actually run.
// Since in this program executing the c case is much more likely, the dynamic
// analysis will probably not be able to detect the unlikely, but still possible
// bug that arises when the d case is triggered.
// By forcing the program to execute the d case as the preferred case,
// GFuzz is able to aid in detecting this bug.

func TestSelect(_ *testing.T) {
	c := make(chan int)
	d := make(chan int)

	go func() {
		// some code
		time.Sleep(300 * time.Millisecond)

		c <- 1
	}()

	go func() {
		// some code that in most cases takes longer than the code in the other routine
		time.Sleep(500 * time.Millisecond)

		<-d
	}()

	select {
	case <-c:
	case <-d:
		panic("CODE WITH PANIC")
	}
}
