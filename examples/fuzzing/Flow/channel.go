package main

import (
	"time"
)

// The following code is an example in which Flow fuzzing is helpful.
// It consists of an unbuffered channel with 2 sends and once receive operations.
// This means, that only one of the sends will be able to execute.
// Because of the surrounding code (simulated by sleep), send 1 is much more
// likely to communicate with the receive, but it is not impossible, that send 2
// may be first. Only if send 2 is first, the code with the bug can be executed
// and the bug detected.
// By delaying the successful channel communication of send 1, giving send 2
// the opportunity to execute, we give the analysis the chance to
// run and detect the possible bug.

// The same effect can be reached if all the sends are replaced by receives and
// the receive is replaced by a send.

func codeWithBug() {
	e := make(chan int)
	close(e)
	e <- 1
}

func main() {
	c := make(chan int)

	go func() {
		// some code
		time.Sleep(100 * time.Second)

		c <- 1 // send 1
	}()

	go func() {
		// some code
		time.Sleep(300 * time.Second)

		c <- 1 // send 2
		codeWithBug()
	}()

	<-c
}
