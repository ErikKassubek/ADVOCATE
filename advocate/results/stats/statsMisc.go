// Copyright (c) 2025 Sebastian Pohsner
//
// File: statsMisc.go
// Brief: Collect miscellaneous statistics about the advocate run
//
// Author: Erik Kassubek
// Created: 2025-02-25
//
// License: BSD-3-Clause

package stats

import (
	"advocate/utils/log"
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// collected information
const (
	TestName                      = "TestName"
	NumDeadlocksInfeasible        = "NoDeadlocksInfeasible"
	NumDeadlocksInfeasibleUnique  = "NoDeadlocksInfeasibleUnique"
	NumGuardLock                  = "NoGuardLock"
	NumGuardLockUnique            = "NoGuardLockUnique"
	ReplayDeadlockStuck           = "ReplayDeadlockStuck"
	ReplayDeadlockNumStuckMutexes = "ReplayDeadlockNoStuckMutexes"
	ReplayDeadlockReachedEnd      = "ReplayDeadlockReachedEnd"
	NumIgnoreConcurrent           = "NoIgnoreConcurrent"
	NumIgnoreConcurrentUnique     = "NoIgnoreConcurrentUnique"
	NumUnlockInOtherThread        = "NoUnlockInOtherThread"
	NumUnlockInOtherThreadUnique  = "NoUnlockInOtherThreadUnique"
)

var miscStats = []string{TestName, NumDeadlocksInfeasible, NumDeadlocksInfeasibleUnique, NumGuardLock, NumGuardLockUnique, ReplayDeadlockStuck, ReplayDeadlockNumStuckMutexes, NumIgnoreConcurrent, NumIgnoreConcurrentUnique, ReplayDeadlockReachedEnd, NumUnlockInOtherThread, NumUnlockInOtherThreadUnique}

// Collect miscellaneous statistics about the run
//
// Parameter:
//   - dataPath string: path to the result folder
//
// Returns:
//   - map[string]int: map with the stats
//   - error
func statsMisc(dataPath, testName string) (map[string]int, error) {
	stats := map[string]int{}
	for _, v := range miscStats {
		stats[v] = 0
	}

	filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Base(path) != "output.log" {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		// Check if correct test
		scanner.Scan() // skip the first line
		scanner.Scan()
		line := scanner.Text()
		if !strings.HasSuffix(line, " "+testName) {
			return nil
		}

		for scanner.Scan() {
			line := scanner.Text()

			// ADD MISC STATISTICS HERE
			if strings.Contains(line, "Cycle Entry with no concurrent requests") {
				stats[NumDeadlocksInfeasible]++
				stats[NumDeadlocksInfeasibleUnique] = 1
			}

			if strings.Contains(line, "Locksets are not disjoint (guard)") {
				stats[NumGuardLock]++
				stats[NumGuardLockUnique] = 1
			}

			if strings.Contains(line, "Ignoring an event because it is concurrent with an already stored event") {
				stats[NumIgnoreConcurrent]++
				stats[NumIgnoreConcurrentUnique] = 1
			}
			if strings.Contains(line, "Lock not found in lockset! Has probably been released in another thread, this is an unsupportd case.") {
				stats[NumUnlockInOtherThread]++
				stats[NumUnlockInOtherThreadUnique] = 1
			}

			if strings.Contains(line, "Number of routines waiting on mutexes: ") {
				stats[ReplayDeadlockReachedEnd]++
				a, err := strconv.Atoi(strings.TrimPrefix(line, "Number of routines waiting on mutexes: "))
				if err != nil {
					log.Error("Failed to read number of waiting mutexes:", err.Error())
				} else {
					if a > 0 {
						stats[ReplayDeadlockStuck]++
						stats[ReplayDeadlockNumStuckMutexes] += a
					}
				}
			}

		}
		return nil
	})

	return stats, nil
}
