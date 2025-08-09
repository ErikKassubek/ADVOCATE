package main

import "testing"

// This function does not contain a bug, but is an illustration on a program,
// that can lead to an impossible mutation using GoPie, which can
// be directly filtered out by the HB information in GoPieHB.
// The two send operations happen on the the same element but in different
// routines. For GoPie, which does not record fork operations, the two
// sends are in a Rel2 relationship. GoPie can therefore create a mutation,
// where the order of the send is changed.
// But it is easy to see, that this order cannot be executed, since the fork
// must happen after the first send. Therefore the second send must also
// happen after the first send.
// Using the HB information, this can be detected, and the mutation is
// discarded, without being run, either reducing the number of necessary
// runs or making time for more interesting runs.

func TestGoPieImpOrder(_ *testing.T) {
	c := make(chan int, 3)

	c <- 1 // send 1

	go func() { // fork
		c <- 1 // send 2
	}()
}
