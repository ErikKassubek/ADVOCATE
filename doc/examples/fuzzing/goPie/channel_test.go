package main

import (
	"testing"
)

// The following code is an example in which GoPie is helpful.
// It consists of an unbuffered channel with 2 sends and once receive operations.
// This means, that only one of the sends will be able to execute.
// Because of the surrounding code (simulated by sleep), send 1 is much more
// likely to communicate with the receive, but it is not impossible, that send 2
// may be first. Only if send 2 is first, the code with the bug can be executed
// and the bug detected.
// By flipping the order of the sends, giving send 2
// the opportunity to execute, we give the analysis the chance to
// run and detect the possible bug.

// The same effect can be reached if all the sends are replaced by receives and
// the receive is replaced by a send.

func TestGoPieChannel(t *testing.T) {
	c := make(chan int)
	d := make(chan int)

	go func() {

		c <- 1 // send 1
	}()

	go func() {
		// some code
		// time.Sleep(100 * time.Millisecond)
		c <- 1 // send 2

		close(d)
		d <- 1
	}()

	<-c
	println("DONE")

	// time.Sleep(200 * time.Millisecond)
}
