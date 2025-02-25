// Copyright (c) 2024 Erik Kassubek
//
// File: timeMeasurement.go
// Brief: Timer to measure the times
//
// Author: Erik Kassubek
// Created: 2024-10-02
//
// License: BSD-3-Clause

package timer

import (
	"time"
)

type Timer struct {
	startTime time.Time
	elapsed   time.Duration
	running   bool
}

func (t *Timer) Start() {
	if t.running {
		return
	}

	t.startTime = time.Now()
	t.running = true
}

func (t *Timer) Stop() {
	if !t.running {
		return
	}
	t.elapsed += time.Since(t.startTime)
	t.running = false
	return
}

func (t *Timer) GetTime() time.Duration {
	if t.running {
		return t.elapsed + time.Since(t.startTime)
	}
	return t.elapsed
}

func (t *Timer) Reset() {
	t.running = false
	t.elapsed = time.Duration(0)
}
