// Copyright (c) 2024 Erik Kassubek
//
// File: clock.go
// Brief: Struct and functions of vector clocks vc
//
// Author: Erik Kassubek
// Created: 2023-07-25
//
// License: BSD-3-Clause

package clock

import (
	"advocate/utils/log"
	"fmt"
	"runtime"
	"strconv"
)

// VectorClock is a vector clock
// Fields:
//
//   - size int: The size of the vector clock
//   - clock []int: The vector clock
type VectorClock struct {
	size  int
	clock map[uint32]uint32
}

// NewVectorClock creates and returns a new, empty vector clock
//
// Parameter:
//   - size int: The size of the vector clock
//
// Returns:
//   - *VectorClock: The new vector clock
func NewVectorClock(size int) *VectorClock {
	if size < 0 {
		size = 0
	}
	c := make(map[uint32]uint32)
	return &VectorClock{
		size:  size,
		clock: c,
	}
}

// NewVectorClockSet creates a new vector clock and set it
//
// Parameter:
//   - size int: The size of the vector clock
//   - cl (map[uint32]iunt)32: The vector clock
//
// Returns:
//   - *VectorClock: A Pointer to the new vector clock
func NewVectorClockSet(size int, cl map[uint32]uint32) *VectorClock {
	vc := NewVectorClock(size)

	if cl == nil {
		return vc
	}

	if size < 0 {
		size = 0
	}

	for rout, val := range cl {
		if rout > uint32(size) {
			continue
		}
		vc.clock[rout] = val
	}

	return vc
}

// GetSize returns the size of the vector clock
//
// Returns:
//   - int: The size of the vector clock
func (vc VectorClock) GetSize() int {
	return int(vc.size)
}

// GetValue returns the value of the vector clock at a given index
//
// Parameter:
//   - index int: the index to get the value for
//
// Returns:
//   - uint32: the value at the given index
func (vc *VectorClock) GetValue(index int) uint32 {
	if index > vc.size {
		return 0
	}

	if val, ok := vc.clock[uint32(index)]; ok {
		return val
	}
	return 0
}

// SetValue sets a value of the vector clock at a given index
//
// Parameter:
//   - index int: the index to set the value for
//   - value uint32: the new value
func (vc *VectorClock) SetValue(index int, value uint32) {
	vc.clock[uint32(index)] = value
}

// GetClock returns the vector clock
//
// Returns:
//   - map[uint32]uint32: The vector clock
func (vc *VectorClock) GetClock() map[uint32]uint32 {
	return vc.clock
}

// ToString returns a string representation of the vector clock
//
// Returns:
//   - string: The string representation of the vector clock
func (vc *VectorClock) ToString() string {
	str := "["
	for i := 1; i <= vc.size; i++ {
		str += fmt.Sprint(vc.GetValue(i))
		if i <= vc.size-1 {
			str += ", "
		}
	}
	str += "]"
	return str
}

// Inc increments the vector clock at the given position
//
// Parameter:
//   - routine int: The routine to increment
func (vc *VectorClock) Inc(routine int) {
	if vc == nil {
		return
	}

	if routine > int(vc.size) {
		return
	}

	if vc.clock == nil {
		vc.clock = make(map[uint32]uint32)
	}

	vc.clock[uint32(routine)]++
}

// Sync updates the vector clock with the received vector clock
//
// Parameter:
//   - rec *VectorClock: The received vector clock
//
// Returns:
//   - *VectorClock: The synced vc (not a copy)
func (vc *VectorClock) Sync(rec *VectorClock) *VectorClock {
	if vc == nil {
		vc = rec.Copy()
		return vc
	}

	if rec == nil {
		return vc
	}

	if vc.size == 0 && rec.size == 0 {
		_, file, line, _ := runtime.Caller(1)
		log.Error("Sync of empty vector clocks: " + file + ":" + strconv.Itoa(line))
	}

	if vc.size == 0 {
		vc = NewVectorClock(rec.size)
	}

	if rec.size == 0 {
		return vc
	}

	for i := 1; i <= vc.size; i++ {
		if rec.GetValue(i) > vc.GetValue(i) {
			vc.SetValue(i, rec.GetValue(i))
		}
	}

	return vc
}

// Copy creates a copy of the vector clock
//
// Returns:
//   - *VectorClock: The copy of the vector clock
func (vc *VectorClock) Copy() *VectorClock {
	if vc == nil {
		return nil
	}

	newVc := NewVectorClock(vc.size)
	for rout, val := range vc.clock {
		newVc.clock[rout] = val
	}
	return newVc
}

// IsEqual checks if the the parameter vc2 is equal to the vc
func (vc *VectorClock) IsEqual(vc2 *VectorClock) bool {
	if vc.size != vc2.size {
		return false
	}

	for i := 1; i <= vc.size; i++ {
		if vc.GetValue(i) != vc2.GetValue(i) {
			return false
		}
	}

	return true
}

// IsMapVcEqual determines if two maps of vector clock are equal, meaning for
// each key they have the same vector clock as vale
//
// Parameter:
//   - v1 map[int]*VectorClock: vector clock 1
//   - v2 map[int]*VectorClock: vector clock 2
//
// Returns:
//   - bool: true if they are equal, false otherwise
func IsMapVcEqual(v1 map[int]*VectorClock, v2 map[int]*VectorClock) bool {
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
