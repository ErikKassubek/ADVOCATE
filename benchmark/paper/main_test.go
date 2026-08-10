package main

import (
	 _ "gocct"
	"fmt"
	"testing"
	"time"
)

func TestBlkBug(_ *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- calc()
	}()

	select {
	case x := <-ch:
		fmt.Println(x)
	case <-time.After(1 * time.Second):
		fmt.Println("Timeout")
	}
}

func TestPanicBug(_ *testing.T) {
	ch := make(chan int)
	go func() { ch <- 1 }()
	go func() {
		for {
			x, more := <-ch
			if more {
				consume(x)
			} else {
				break
			}
		}
	}()
	close(ch)
}

// helper
// func calc() int {
// 	time.Sleep(500 * time.Millisecond)
// 	return 1
// }

// func consume(x int) {
// 	fmt.Println(x)
// }
