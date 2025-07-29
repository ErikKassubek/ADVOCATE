package main

import (
	"advocate"
	"fmt"
	"sync/atomic"
	"testing"
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

func TestGoPieAtomic(_ *testing.T) {
	// ======= Preamble Start =======
  advocate.InitFuzzing("/home/erik/Uni/Advocate/ADVOCATE/examples/fuzzing/goPie/fuzzingTraces/fuzzingTrace_7", 30)
  defer advocate.FinishFuzzing()
  // ======= Preamble End =======
	a := atomic.Int32{}
	c := make(chan int)

	fmt.Println("1")

	go func() {
		// some code
		fmt.Println("2")

		i := 0
		for n := 0; n < 10000; n++ {
			i *= 2
			i /= 2
		}

		fmt.Println("3")
		a.Store(1)

		fmt.Println("4")
	}()

	// some code
	fmt.Println("5")

	if a.Load() == 1 {
		fmt.Println("7")
		close(c)
		c <- 1
	}

	fmt.Println("8")

	time.Sleep(500 * time.Millisecond)
}
