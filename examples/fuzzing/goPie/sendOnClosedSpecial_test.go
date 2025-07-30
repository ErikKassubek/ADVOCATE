package main

import (
	"advocate"
	"testing"
	"time"
)

func TestSendOnClosed(t *testing.T) {
	// ======= Preamble Start =======
  advocate.InitFuzzing("/home/erik/Uni/Advocate/ADVOCATE/examples/fuzzing/goPie/fuzzingTraces/fuzzingTrace_1", 30)
  defer advocate.FinishFuzzing()
  // ======= Preamble End =======

	c := make(chan int)

	go func() {
		println("TEST1")
		c <- 1
		println("TEST2")
	}()

	go func() {
		println("TEST3")
		<-c
		println("TEST4")
	}()

	println("TEST5")
	time.Sleep(100 * time.Millisecond)
	println("TEST6")
	// panic("A")
	close(c)

	println("TEST7")
	time.Sleep(300 * time.Millisecond)
	println("TEST8")
}
