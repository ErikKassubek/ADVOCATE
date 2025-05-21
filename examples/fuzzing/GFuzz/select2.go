package main

import (
	"testing"
)

// The following code is an example in which GFuzz is helpful.
// It consists of two select with two possible path.
// Only if the first select chooses the d case and the second select chooses
// the c channel, a bug is triggered. By choosing cases at random, this is
// a chance of 1/4. It is easy to see how this could be scaled up to
// get even smaller probabilities.
// Using GFuzz we try out many (in this case probably all) combinations
// of select cases, allowing us to find those bugs.

func TestSelect2(_ *testing.T) {
	c := make(chan int)
	d := make(chan int)
	e := make(chan int)

	go func() {
		c <- 1
		c <- 1
	}()

	go func() {
		d <- 1
		d <- 1
	}()

	select {
	case <-c:
	case <-d:
		close(e)
	}

	select {
	case <-c:
		e <- 1
	case <-d:
	}

}
