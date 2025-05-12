package main

import "time"

// The following code is an example in which GFuzz is helpful.
// It consists of a select with one case an a default.
// The case (triggered by the send on c) calls code which contains a send on closed channel bug.
// Let's assume, that it is very likely, but not guaranteed, that the
// select, and therefore the default case is executed before the send on c.
// In a dynamic analysis, we can only analyze program parts that are actually run.
// Since in this program executing the default is much more likely, the dynamic
// analysis will probably not be able to detect the unlikely, but still possible
// bug that arises when the c case is triggered.
// By forcing the program to execute the c case as the preferred case,
// GFuzz is able to aid in detecting this bug.

func codeWithBug() {
	e := make(chan int)
	close(e)
	e <- 1
}

func main() {
	c := make(chan int)

	go func() {
		// some code
		time.Sleep(300 * time.Millisecond)
		c <- 1
	}()

	select {
	case <-c:
		codeWithBug()
	default:
	}
}
