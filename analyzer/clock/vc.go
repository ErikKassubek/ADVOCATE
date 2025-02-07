// Copyrigth (c) 2024 Erik Kassubek
//
// File: vc.go
// Brief: Struct and functions of vector clocks vc
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package clock

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
)

/*
 * vectorClock is a vector clock
 * Fields:
 *   size (int): The size of the vector clock
 *   clock ([]int): The vector clock
 */
type VectorClock struct {
	size  int
	clock map[int]int
}

/*
 * Create a new vector clock
 * Args:
 *   size (int): The size of the vector clock
 * Returns:
 *   (vectorClock): The new vector clock
 */
func NewVectorClock(size int) VectorClock {
	if size < 0 {
		size = 0
	}
	c := make(map[int]int)
	for i := 1; i <= size; i++ {
		c[i] = 0
	}

	return VectorClock{
		size:  size,
		clock: c,
	}
}

/*
 * Create a new vector clock and set it
 * Args:
 *   size (int): The size of the vector clock
 *   cl (map[int]int): The vector clock
 */
func NewVectorClockSet(size int, cl map[int]int) VectorClock {
	clock := NewVectorClock(size)

	if cl == nil {
		return clock
	}

	if size < 0 {
		size = 0
	}

	for i := 1; i <= size; i++ {
		if _, ok := cl[i]; !ok {
			clock.clock[i] = 0
		} else {
			clock.clock[i] = cl[i]
		}
	}

	return clock
}

/*
 * Get the size of the vector clock
 * Returns:
 *   (int): The size of the vector clock
 */
func (vc VectorClock) GetSize() int {
	return vc.size
}

/*
 * Get the vector clock
 * Returns:
 *   (map[int]int): The vector clock
 */
func (vc VectorClock) GetClock() map[int]int {
	return vc.clock
}

/*
 * Get a string representation of the vector clock
 * Returns:
 *   (string): The string representation of the vector clock
 */
func (vc VectorClock) ToString() string {
	str := "["
	for i := 1; i <= vc.size; i++ {
		str += fmt.Sprint(vc.clock[i])
		if i <= vc.size-1 {
			str += ", "
		}
	}
	str += "]"
	return str
}

/*
 * Increment the vector clock at the given position
 * Args:
 *   routine (int): The routine to increment
 * Returns:
 *   (vectorClock): The vector clock
 */
func (vc VectorClock) Inc(routine int) VectorClock {
	if routine > vc.size {
		return vc
	}

	if vc.clock == nil {
		vc.clock = make(map[int]int)
	}

	vc.clock[routine]++
	return vc
}

/*
 * Update the vector clock with the received vector clock
 * Args:
 *   rec (vectorClock): The received vector clock
 * Returns:
 *   (vectorClock): The new vector clock
 */
func (vc VectorClock) Sync(rec VectorClock) VectorClock {
	if vc.size == 0 && rec.size == 0 {
		_, file, line, _ := runtime.Caller(1)
		log.Print("Sync of empty vector clocks: " + file + ":" + strconv.Itoa(line))
	}

	if vc.size == 0 {
		vc = NewVectorClock(rec.size)
	}

	if rec.size == 0 {
		return vc.Copy()
	}

	copy := rec.Copy()
	for i := 1; i <= vc.size; i++ {
		if vc.clock[i] > copy.clock[i] {
			copy.clock[i] = vc.clock[i]
		}
	}

	return copy
}

/*
 * Create a copy of the vector clock
 * Returns:
 *   (vectorClock): The copy of the vector clock
 */
func (vc VectorClock) Copy() VectorClock {
	newVc := NewVectorClock(vc.size)
	for i := 1; i <= vc.size; i++ {
		newVc.clock[i] = vc.clock[i]
	}
	return newVc
}

/*
 * Check if the the arg vc2 is equal to the vc
 */
func (vc VectorClock) IsEqual(vc2 VectorClock) bool {
	if vc.size != vc2.size {
		return false
	}

	for i := 1; i <= vc.size; i++ {
		if vc.clock[i] != vc2.clock[i] {
			return false
		}
	}

	return true
}

func IsMapVcEqual(v1 map[int]VectorClock, v2 map[int]VectorClock) bool {
	if len(v1) != len(v2) {
		return false
	}

	for k, vc1 := range v1 {
		vc2, ok := v2[k]
		if !ok || !vc1.IsEqual(vc2) {
			return false
		}
	}

	return true
}
