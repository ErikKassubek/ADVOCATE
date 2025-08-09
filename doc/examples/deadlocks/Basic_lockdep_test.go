package main

import (
	"sync"
	"testing"
)

func TestBasicLockdep(t *testing.T) {

	var x, y, z sync.Mutex

	// Starte Go Routine T2
	go func() { // Entstehende Lock Dependency
		x.Lock()   // D = (T2, x, {})          Lockset T2 danach: {x}
		y.Lock()   // D = (T2, y, {x, z})      Lockset T2 danach: {x, y, z}
		z.Lock()   // D = (T2, z, {x})      	Lockset T2 danach: {x, z}
		z.Unlock() //                          Lockset T2 danach: {x, y}
		y.Unlock() //                          Lockset T2 danach: {x}
		z.Lock()   // D = (T2, z, {x})         Lockset T2 danach: {x, z}
		x.Unlock() //                          Lockset T2 danach: {z}
		z.Unlock() //                          Lockset T2 danach: {}
	}()

	// Main Routine ist T1
	z.Lock()   // D = (T1, z, {})          Lockset T1 danach: {z}
	x.Lock()   // D = (T1, x, {z})         Lockset T1 danach: {z, x}
	x.Unlock() //                          Lockset T1 danach: {z}
	z.Unlock() //                          Lockset T1 danach: {}

}
