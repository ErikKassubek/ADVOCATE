// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_time.go
// Brief: Set of functions using time
//
// Author: Erik Kassubek
// Created: 2024-12-04
//
// License: BSD-3-Clause

package runtime

func sleep(seconds float64) {
	timeSleep(sToNs(seconds))
}

func sToNs(seconds float64) int64 {
	return int64(seconds * 1e9)
}

func hasTimePast(startNs int64, durationS int64) bool {
	durationNano := durationS * 1e9
	return nanotime()-startNs > durationNano
}

func currentTime() int64 {
	return nanotime()
}

// The following is mostly a copy and adaption of functions in time/sleep.go

type afterTimer struct {
	C <-chan struct{}
	r *timeTimer
}

func newTimer2(d int64) *afterTimer {
	c := make(chan struct{}, 1)
	t := &afterTimer{
		C: c,
		r: newTimer(when(d), 0, sendTime, c, nil),
	}
	return t
}

func sendTime(c any, seq uintptr, delay int64) {
	select {
	case c.(chan struct{}) <- struct{}{}:
	default:
	}
}

func after(d int64) <-chan struct{} {
	return newTimer2(d).C
}

func when(nano int64) int64 {
	if nano <= 0 {
		return currentTime()
	}
	t := currentTime() + int64(nano)
	if t < 0 {
		// N.B. runtimeNano() and d are always positive, so addition
		// (including overflow) will never result in t == 0.
		t = 1<<63 - 1 // math.MaxInt64
	}
	return t
}

// ADVOCATE-FILE-END
