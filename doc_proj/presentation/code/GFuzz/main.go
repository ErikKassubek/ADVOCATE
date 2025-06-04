package main

import "time"

func main() {
	c := make(chan int, 2)
	d := make(chan int, 2)
	e := make(chan int, 1)

	go func() {
		c <- 1
		c <- 1
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		d <- 1
		d <- 1
	}()

	select {
	case <-c:
	case <-d:
		close(e)
	}

	select {
	case <-c:
	case <-d:
		e <- 1
	}
}