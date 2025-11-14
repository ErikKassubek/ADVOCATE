// advocate/Examples/Examples_Simple/MixedDeadlock/mixed_deadlock_test.go

/*
------------------------------------------------------------
Mixed Deadlock Test Matrix (MDS-2)
------------------------------------------------------------

Each test corresponds to one theoretical case:
- MD2-1 : both inside CS  (symmetric)
- MD2-2 : sender inside, receiver after CS (lock→channel)
- MD2-3 : sender after CS, receiver inside (channel→lock)
- MD-Close : close–receive dependency

Variants:
- U: unbuffered channel
- B: buffered channel

LockType:
- READ/READ
- READ/WRITE

*/

package main

import (
	"sync"
	"testing"
	"time"
)

// Helper
func run2(a, b func()) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); a() }()
	go func() { defer wg.Done(); b() }()
	wg.Wait()
}

// ------------------------------------------------------------
// MD2-1: Both sender and receiver in CS
// ------------------------------------------------------------

// MD2-1B: Buffered Variant
func TestMixedDeadlock_MD2_1B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	sender := func() {
		m.Lock()
		c <- 1 // send inside CS
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// ------------------------------------------------------------
// MD2-2: Sender inside CS, Receiver with PCS
// ------------------------------------------------------------

// MD2-2U: Unbuffered Variant
func TestMixedDeadlock_MD2_2U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let receiver complete PCS
		m.Lock()
		c <- 1 // unbuffered send inside CS (blocks until rcv)
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}

// MD2-2B: Buffered Variant
func TestMixedDeadlock_MD2_2B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let receiver complete PCS (not necessary, always non-blocking)
		m.Lock()
		c <- 1 // buffered send inside CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // receive after PCS
	}

	run2(sender, receiver)
}

// ------------------------------------------------------------
// MD2-3: Sender with PCS, Receiver inside CS
// ------------------------------------------------------------

// MD2-3U: Unbuffered Variant
func TestMixedDeadlock_MD2_3U(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// MD2-3B: Buffered Variant
func TestMixedDeadlock_MD2_3B(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // send after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond) // sleep to let sender complete PCS
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// ------------------------------------------------------------
// CLOSE TESTS: MDX-Y-CLOSE VARIANTS
// ------------------------------------------------------------

// MD2-1B: Buffered Variant
func TestMixedDeadlock_MD2_1_CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // buffered

	sender := func() {
		m.Lock()
		close(c) // close inside CS
		m.Unlock()
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS
		m.Unlock()
	}

	run2(sender, receiver)
}

// MD-CloseU: Unbuffered Variant (Mirror of MD-2-2U)
func TestMixedDeadlock_MD_2_2CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		time.Sleep(50 * time.Millisecond) // let receiver finish first
		m.Lock()
		close(c) // close in CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // recv with PCS
	}

	run2(receiver, closer)
}

// MD-CloseB: Buffered Variant (Mirror of MD-2-2B)
func TestMixedDeadlock_MD_2_2CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1) // unbuffered

	closer := func() {
		time.Sleep(50 * time.Millisecond) // let receiver finish first
		m.Lock()
		close(c) // close in CS
		m.Unlock()
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // recv with PCS
	}

	run2(receiver, closer)
}

// MD-CloseU: Unbuffered Variant (Mirror of MD-2-3U)
func TestMixedDeadlock_MD_2_3CloseU(t *testing.T) {
	var m sync.Mutex
	c := make(chan int) // unbuffered

	closer := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		close(c) // close after PCS
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS (blocked until close)
		m.Unlock()
	}

	run2(receiver, closer)
}

// MD-CloseB: Buffered Variant (Mirror of MD-2-3B)
func TestMixedDeadlock_MD_2_3CloseB(t *testing.T) {
	var m sync.Mutex
	c := make(chan int, 1)

	closer := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		close(c)
	}

	receiver := func() {
		time.Sleep(50 * time.Millisecond)
		m.Lock()
		<-c // receive inside CS (blocked until close)
		m.Unlock()
	}

	run2(receiver, closer)
}

// ------------------------------------------------------------
// LOCKTYPE TESTS: READ/READ
// ------------------------------------------------------------

// READ/READ
func TestMixedDeadlock_MD_Read(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

	reader_1 := func() {
		rw.RLock()
		<-c
		rw.RUnlock()
	}

	reader_2 := func() {
		rw.RLock()
		c <- 1
		rw.RUnlock()
	}

	run2(reader_1, reader_2)
}

// ------------------------------------------------------------
// LOCKTYPE TESTS: READ/WRITE MD-Cases
// ------------------------------------------------------------

// READ/WRTIE MD2-1B
func TestMixedDeadlock_MD_2_1B_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	writer := func() {
		rw.Lock()
		c <- 1 // send inside CS
		rw.Unlock()
	}

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender finish PCS
		rw.RLock()
		<-c // receive in CS
		rw.RUnlock()
	}

	run2(reader, writer)
}

// READ/WRTIE MD2-2U
func TestMixedDeadlock_MD_2_2U_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

	writer := func() {
		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
		rw.Lock()
		c <- 1 // send inside CS
		rw.Unlock()
	}

	reader := func() {
		rw.RLock()
		time.Sleep(10 * time.Millisecond)
		rw.RUnlock() // PCS
		<-c          // receive after PCS
	}

	run2(reader, writer)
}

// READ/WRTIE MD2-2B
func TestMixedDeadlock_MD_2_2B_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	writer := func() {
		time.Sleep(50 * time.Millisecond) // let receiver finish PCS
		rw.Lock()
		c <- 1 // send inside CS
		rw.Unlock()
	}

	reader := func() {
		rw.RLock()
		time.Sleep(10 * time.Millisecond)
		rw.RUnlock() // PCS
		<-c          // receive after PCS
	}

	run2(reader, writer)
}

// READ/WRTIE MD2-3U
func TestMixedDeadlock_MD_2_3_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int)

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender finish PCS
		rw.RLock()
		<-c // receive inside CS
		rw.RUnlock()
	}

	writer := func() {
		rw.Lock()
		time.Sleep(10 * time.Millisecond)
		rw.Unlock() // PCS
		c <- 1      // send after PCS
	}

	run2(reader, writer)
}

// READ/WRTIE MD2-3B
func TestMixedDeadlock_MD_2_3B_RW(t *testing.T) {
	var rw sync.RWMutex
	c := make(chan int, 1)

	reader := func() {
		time.Sleep(50 * time.Millisecond) // let sender finish PCS
		rw.RLock()
		<-c // receive inside CS
		rw.RUnlock()
	}

	writer := func() {
		rw.Lock()
		time.Sleep(10 * time.Millisecond)
		rw.Unlock() // PCS
		c <- 1      // send after PCS
	}

	run2(reader, writer)
}

// ------------------------------------------------------------
// FALSE POSITIVE TESTS
// ------------------------------------------------------------

func TestMixedDeadlock_No_MD_BeforeCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		c <- 1 // after PCS
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	receiver := func() {
		<-c // after PCS
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_No_MD_AfterPCS(t *testing.T) {
	var m sync.Mutex
	c := make(chan int)

	sender := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		c <- 1 // after PCS
	}

	receiver := func() {
		m.Lock()
		time.Sleep(10 * time.Millisecond)
		m.Unlock()
		<-c // after PCS
	}

	run2(sender, receiver)
}

func TestMixedDeadlock_No_MD_DifferentLocks(t *testing.T) {
	var m1, m2 sync.Mutex
	c := make(chan int)

	sender := func() {
		m1.Lock()
		c <- 1
		m1.Unlock()
	}

	receiver := func() {
		m2.Lock()
		<-c
		m2.Unlock()
	}

	run2(sender, receiver)
}
