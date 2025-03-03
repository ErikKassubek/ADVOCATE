// Copyright (c) 2025 Erik Kassubek
//
// File: memory.go
// Brief: Cancel analysis when not enough memory
//
// Author: Erik Kassubek
// Created: 2025-03-03
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/utils"
	"reflect"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/mem"
)

func memorySupervisor(cancel chan struct{}) {
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

	thresholdRAM := uint64(float64(v.Total) * 0.05)
	thresholdSwap := uint64(200 * 1024 * 1024) // 200mb

	startSwap := s.Used

	for {
		select {
		case <-cancel:
			return
		default:
		}

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
	utils.LogErrorf("Cancel analysis, not enough RAM")
	wasCanceled.Store(true)
	wasCanceledRam.Store(true)
}

func WasCanceled() (bool, bool) {
	return wasCanceled.Load(), wasCanceledRam.Load()
}

func LogSizes() {
	utils.LogError("Trace: ", getSizeInMB(traces))
	utils.LogError("CurrentIndex: ", getSizeInMB(currentIndex))

	utils.LogError("closeData: ", getSizeInMB(closeData))
	utils.LogError("lastSendRoutine: ", getSizeInMB(lastSendRoutine))
	utils.LogError("lastRecvRoutine: ", getSizeInMB(lastRecvRoutine))
	utils.LogError("hasSend: ", getSizeInMB(hasSend))
	utils.LogError("mostRecentSend: ", getSizeInMB(mostRecentSend))
	utils.LogError("hasReceived: ", getSizeInMB(hasReceived))
	utils.LogError("mostRecentReceive: ", getSizeInMB(mostRecentReceive))
	utils.LogError("bufferedVCs: ", getSizeInMB(bufferedVCs))
	utils.LogError("wgAdd: ", getSizeInMB(wgAdd))
	utils.LogError("wgDone: ", getSizeInMB(wgDone))
	utils.LogError("allLocks: ", getSizeInMB(allLocks))
	utils.LogError("allUnlocks: ", getSizeInMB(allUnlocks))
	utils.LogError("lockSet: ", getSizeInMB(lockSet))
	utils.LogError("mostRecentAcquire: ", getSizeInMB(mostRecentAcquire))
	utils.LogError("mostRecentAcquireTotal: ", getSizeInMB(mostRecentAcquireTotal))
	utils.LogError("relW: ", getSizeInMB(relW))
	utils.LogError("relR: ", getSizeInMB(relR))
	utils.LogError("leakingChannels: ", getSizeInMB(leakingChannels))
	utils.LogError("selectCases: ", getSizeInMB(selectCases))
	utils.LogError("allForks: ", getSizeInMB(allForks))
	utils.LogError("fuzzingFlowOnce: ", getSizeInMB(fuzzingFlowOnce))
	utils.LogError("fuzzingFlowMutex: ", getSizeInMB(fuzzingFlowMutex))
	utils.LogError("fuzzingFlowSend: ", getSizeInMB(fuzzingFlowSend))
	utils.LogError("fuzzingFlowRecv: ", getSizeInMB(fuzzingFlowRecv))
	utils.LogError("executedOnce: ", getSizeInMB(executedOnce))
	utils.LogError("fuzzingCounter: ", getSizeInMB(fuzzingCounter))
	utils.LogError("currentVCHb: ", getSizeInMB(currentVCHb))
	utils.LogError("currentVCWmhb: ", getSizeInMB(currentVCWmhb))
	utils.LogError("channelWithoutPartner: ", getSizeInMB(channelWithoutPartner))
	utils.LogError("currentState: ", getSizeInMB(currentState))
}

// GetSizeInMB recursively estimates the memory usage of a slice, map, or nested structures.
func getSizeInMB(data interface{}) float64 {
	visited := make(map[uintptr]bool) // Track visited pointers to prevent infinite loops
	sizeBytes := getSizeRecursive(reflect.ValueOf(data), visited)
	return float64(sizeBytes) / (1024 * 1024) // Convert to MB
}

// getSizeRecursive calculates the size of a value recursively.
func getSizeRecursive(val reflect.Value, visited map[uintptr]bool) int {
	if !val.IsValid() {
		return 0
	}

	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return 0
		}
		// Use Elem() to get the actual value inside the pointer/interface
		return getSizeRecursive(val.Elem(), visited)

	case reflect.Slice, reflect.Array:
		totalSize := int(unsafe.Sizeof(val.Interface())) // Slice header size
		for i := 0; i < val.Len(); i++ {
			totalSize += getSizeRecursive(val.Index(i), visited)
		}
		return totalSize

	case reflect.Map:
		if val.Len() == 0 {
			return 0
		}
		// Check if map is already visited
		ptr := val.Pointer()
		if ptr != 0 && visited[ptr] {
			return 0
		}
		visited[ptr] = true

		totalSize := int(unsafe.Sizeof(val.Interface())) // Map header size
		iter := val.MapRange()
		for iter.Next() {
			totalSize += getSizeRecursive(iter.Key(), visited)
			totalSize += getSizeRecursive(iter.Value(), visited)
		}
		return totalSize

	default:
		return int(unsafe.Sizeof(val.Interface()))
	}
}
