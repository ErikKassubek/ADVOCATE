package main

import (
	"sync"
	"time"
)

var (
	m = sync.Mutex{}
	c = make(chan bool)
)

func R1() {
	for i := 0; i < 2; i++ {
		<-c
		m.Lock()
		m.Unlock()
	}
}

func R2() {
	m.Lock()
	c <- true
	m.Unlock()
}

func main() {
	go R1()
	go R2()
	time.Sleep(200 * time.Millisecond)
	R2()

	time.Sleep(500 * time.Millisecond)
}
