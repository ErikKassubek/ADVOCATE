// Copyright (c) 2025 Erik Kassubek
//
// File: memory.go
// Brief: Cancel analysis when not enough memory
//
// Author: Erik Kassubek
// Created: 2025-03-03
//
// License: BSD-3-Clause

package memory

import (
	"analyzer/utils"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

var (
	wasCanceled    atomic.Bool
	wasCanceledRam atomic.Bool
)

func MemorySupervisor() {
	// Get the memory stats
	v, err := mem.VirtualMemory()
	if err != nil {
		utils.LogErrorf("Error getting memory info: %v", err)
	}

	// Get the swap stats
	s, err := mem.SwapMemory()
	if err != nil {
		utils.LogErrorf("Error getting swap info: %v", err)
	}

	thresholdRAM := uint64(float64(v.Total) * 0.02)
	thresholdSwap := uint64(1000 * 1024 * 1024) // 1GB

	startSwap := s.Used

	for {
		// Get the memory stats
		v, err = mem.VirtualMemory()
		if err != nil {
			utils.LogErrorf("Error getting memory info: %v", err)
		}

		// Get the swap stats
		s, err = mem.SwapMemory()
		if err != nil {
			utils.LogErrorf("Error getting swap info: %v", err)
		}

		// cancel if available RAM is below the threshold or the used swap is above the threshhold
		if v.Available < thresholdRAM {
			notEnoughSpace()
			return
		}

		if s.Used > thresholdSwap+startSwap {
			notEnoughSpace()
			return
		}

		// Sleep for a while before checking again
		time.Sleep(1 * time.Second)
	}
}

func notEnoughSpace() {
	wasCanceled.Store(true)
	wasCanceledRam.Store(true)
	utils.LogError("Not enough RAM")
}

func CheckCanceled() (bool, bool) {
	return wasCanceled.Load(), wasCanceledRam.Load()
}

func WasCanceled() bool {
	return wasCanceled.Load()
}

func WasCanceledRAM() bool {
	return wasCanceledRam.Load()
}

func Reset() {
	wasCanceled.Store(false)
	wasCanceledRam.Store(false)
}

func Cancel() {
	wasCanceled.Store(true)
}
