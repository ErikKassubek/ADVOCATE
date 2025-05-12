package main

import "time"

// Based on the original GFuzz, we build some improvements based on our HB
// analysis. This mainly focuses on reducing the number of impossible runs.
// In the given example, the select contains a loot of cases, but only 2 of them
// (c and d) have concurrent partners, according to the HB analysis.
// The original GFuzz implementation would still treat all of the cases the same
// selecting them as the preferred case with the same probability.
// To get better results, especially since in many cases it is not possible to
// execute all possible combinations of preferred cases (path explosion problem)
// we focus on cases, which according to the HB analysis have possible partner.
// This reduces the number of fuzzing runs, where the program is not able
// to execute the preferred case.
// In some cases (not shown here), it may be possible for a select case to
// have a possible partner in the fuzzing run, even though the HB analysis
// did not indicate it, e.g. if a differently executed select case in a
// previous case, leads to the execution of such a partner. We therefor
// do select cases without an indicated partner, but with a smaller probability
// that cases, that have a possible partner

func codeWithBug() {
	e := make(chan int)
	close(e)
	e <- 1
}

func main() {
	c := make(chan int)
	d := make(chan int)
	e := make(chan int)
	f := make(chan int)
	g := make(chan int)
	h := make(chan int)
	i := make(chan int)
	j := make(chan int)

	go func() {
		// some code
		time.Sleep(300 * time.Millisecond)

		c <- 1
	}()

	go func() {
		// some code that in most cases takes longer than the code in the other routine
		time.Sleep(500 * time.Millisecond)

		<-d
	}()

	select {
	case <-c:
	case <-d:
		codeWithBug()
	case <-e:
	case <-f:
	case <-g:
	case <-h:
	case <-i:
	case <-j:
	}
}
