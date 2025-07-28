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

// Timer is a timer that can be started and stopped
//
// Parameter:
//   - startTime time.Time: time of the last start
//   - elapsedTime time.Duration: total elapsed time
//   - running bool: true if running, false if stopped
type Timer struct {
	startTime time.Time
	elapsed   time.Duration
	running   bool
}

// Start a timer
func (t *Timer) Start() {
	if t.running {
		return
	}

	t.startTime = time.Now()
	t.running = true
}

// Stop a timer
func (t *Timer) Stop() {
	if !t.running {
		return
	}
	t.elapsed += time.Since(t.startTime)
	t.running = false
}

// GetTime returns the elapsed time of the timer
//
// Returns:
//   - time.Duration: current elapsed time of timer
func (t *Timer) GetTime() time.Duration {
	if t.running {
		return t.elapsed + time.Since(t.startTime)
	}
	return t.elapsed
}

// Reset the timer
func (t *Timer) Reset() {
	t.running = false
	t.elapsed = time.Duration(0)
}

// IsRunning returns if the timer is currently running
func (t *Timer) IsRunning() bool {
	return t.running
}
