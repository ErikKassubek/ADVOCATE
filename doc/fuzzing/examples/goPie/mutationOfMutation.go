package main

import (
	"testing"
	"time"
)

// The original GoPie implementation will choose some scheduling chains after the first run
// This chain will then be mutated over and over again.
// If a mutation was executed and uncovered a new execution path, the original
// GoPie implementation will not chose new scheduling chains from this
// new code to mutated (probably because with the limited replay, it cannot be
// guaranteed/is fairly unlikely, that a run on such a mutation would reach the
// mutated are)
// In our improved version, we also choose new chains to mutate from the already
// mutated executions.
// Lets assume that in this example, in the first execution, the send of 0
// was executed before the send of 1, leading to the receive receiving 0, and
// therefore not executing the if statement.
// Let now assume, that a mutation is executed, where the order of the send
// if flipped, leading to the code in the if block being executed, but in such a
// way, that the send on d happens before the close, therefore not triggering the
// send on closed (especially if based on the timing, the send on closed is very
// unlikely)
// GoPie will now continue to only mutate the originally chosen chain. This
// means, the bug could only be detected if the send and close on d would
// at some point be added to the chain by an substitute or augment operations
// and then then mutated in such a way, that there order is reversed. Especially if
// there are other concurrent operations between the if and the potential bug,
// this may need a lot of mutations.
// By also choosing new mutations chains from already mutated runs, our improved
// version may choose a scheduling chain closer to the new bug, therefore
// discovering it requiring fewer runs.

func TestGoPieMutOfMut(_ *testing.T) {
	c := make(chan int, 2)
	d := make(chan int)

	go func() {
		c <- 0
	}()

	go func() {
		c <- 1
	}()

	if a := <-c; a == 1 {
		time.Sleep(100 * time.Millisecond) // other code, simulated by sleep, may contains other concurrency operations

		go func() {
			// some code, simulated by sleep
			time.Sleep(300 * time.Millisecond)

			close(d)
		}()

		// some code, simulated by sleep
		time.Sleep(100 * time.Millisecond)
		d <- 1

	}
}
