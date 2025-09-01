package main

import (
	"sync"
	"testing"
	"time"
)

// The following code is an example in which Flow fuzzing is helpful.
// It consists of a once, that is called twice. Only on of the functions
// in the o.Do is executed, namely the one in the first executed Do.
// In this example, we assume that one od the Do (the one without the bug
// in its function), is much more likely to be executed first. This means, that
// the code containing the bug is very unlikely to be executed during an analysis.
// No matter how good an dynamic analysis is, To be able to detect a bug, it must
// execute the code containing the bug (in the given example the bug is always
// triggered if the function is called, but this may not be the case).
// By delaying the executed Do, we can get the program to execute the
// Do containing the bug, making it possible for the analysis to detect the
// bug.

func TestOnce(_ *testing.T) {
	var o sync.Once

	go func() {
		// some code
		time.Sleep(500 * time.Millisecond)

		o.Do(func() { panic("CODE WITH PANIC") })
	}()

	// some code
	time.Sleep(300 * time.Millisecond)
	o.Do(func() { time.Sleep(100 * time.Millisecond) })
}
