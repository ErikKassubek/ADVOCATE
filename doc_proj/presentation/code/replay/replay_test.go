package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestReplay(_ *testing.T) {
	var mu sync.Mutex
	c := make(chan int, 1)
	d := make(chan int, 1)

	go func() {
		mu.Lock()
		fmt.Println("A")
		mu.Unlock()
	}()

	go func() {
		mu.Lock()
		fmt.Println("B")
		mu.Unlock()
	}()

	go func() {
		c <- 1
		fmt.Println("C")
	}()

	go func() {
		d <- 1
		fmt.Println("D")
	}()

	go func() {
		<-d
		fmt.Println("E")
	}()

	time.Sleep(time.Second)
}
