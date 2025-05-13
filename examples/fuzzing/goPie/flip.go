package main

import (
	"sync/atomic"
	"testing"
)

// This program illustrates, how the missing flip in the GoPie implementation
// may increase the number of mutation required to get find a panic.
// In the following code, lets assume that during the first recording,
// the marked send happened before the receive, and therefore the store
// on x happens before the load, not triggering the panic (in the original
// GFuzz, atomic operations are not recorded/mutated). Since they happen
// in different routines but on the same element, a flip mutation may
// change the order of the send and receive, triggering the panic.
// Without the flip, we would need multiple mutation steps (in this case
// it could be replaced with an abridge to remove the send, followed by an
// augment adding it back in after the receive). Since flip seems to be
// the most impactful mutation, it seems useful to have this mutation.
// In most cases, like the case given here, it is possible to get the same
// result with multiple mutation steps. But since we are in most cases
// not able to run all possible mutation, it seems unwise to not directly use
// a flip mutation.

func TestFlip(_ *testing.T) {
	c := make(chan int, 2) // buffered to allow non-blocking send
	var x atomic.Int32

	c <- 1

	go func() {
		c <- 1 // send
		x.Store(1)
	}()

	<-c // recv
	if x.Load() == 0 {
		panic("CODE WITH PANIC")
	}
}
