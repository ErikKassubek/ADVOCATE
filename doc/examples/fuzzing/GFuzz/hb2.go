package main

import "testing"

// In this example we show how the change of the score function for GFuzz
// may lead to more interesting runs.
// Lets assume, that for all the channels in the program except for x and y,
// send operations that are concurrent to the selects, are available.
// Lets assume, we have two runs during the fuzzing, one where the c case
// is executed and one where the d case is executed. The original GFuzz
// will score both of those executions with the same score, creating the same
// number of mutations.
// But we can see, based on our HB analysis, that in the case where d is executed,
// the selection of a preferred case in the embedded select will fail with a
// higher probability, for the select in the c case, since the HB information
// does detect possible partners for e and f but not for x and y.
// It is therefore sensible to create more mutations based on the run where c
// was executed/was the preferred case than for the run with d.
// By modifying the score function to take the number of select cases with
// possible partners into account, we limit the number select cases where the
// execution of a preferred case is not possible.

func TestHB2(_ *testing.T) {
	c := make(chan int)
	d := make(chan int)
	e := make(chan int)
	f := make(chan int)

	x := make(chan int)
	y := make(chan int)

	go func() {
		// sends of c, d, e, f
	}()

	select {
	case <-c:
		select {
		case <-e:
		case <-f:
		default:
		}
	case <-d:
		select {
		case <-x:
		case <-y:
		default:
		}
	}

}
