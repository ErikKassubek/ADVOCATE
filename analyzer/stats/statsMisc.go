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
	"analyzer/utils"
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	TestName                      = "TestName"
	NumDeadlocksInfeasible        = "NoDeadlocksInfeasible"
	NumDeadlocksInfeasibleUnique  = "NoDeadlocksInfeasibleUnique"
	NumGuardLock                  = "NoGuardLock"
	NumGuardLockUnique            = "NoGuardLockUnique"
	ReplayDeadlockStuck           = "ReplayDeadlockStuck"
	ReplayDeadlockNumStuckMutexes = "ReplayDeadlockNoStuckMutexes"
	ReplayDeadlockReachedEnd      = "ReplayDeadlockReachedEnd"
)

var MiscStats = []string{TestName, NumDeadlocksInfeasible, NumDeadlocksInfeasibleUnique, NumGuardLock, NumGuardLockUnique, ReplayDeadlockStuck, ReplayDeadlockNumStuckMutexes, ReplayDeadlockReachedEnd}

/*
 * Collect miscellaneous statistics about the run
 * Args:
 *     dataPath (string): path to the result folder
 * Returns:
 *     map[string]int: map with the stats
 *     error
 */
func statsMisc(dataPath, testName string) (map[string]int, error) {
	stats := map[string]int{}
	for _, v := range MiscStats {
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
		scanner.Scan() // TODO UGLY!
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

			if strings.Contains(line, "Number of routines waiting on mutexes: ") {
				stats[ReplayDeadlockReachedEnd]++
				a, err := strconv.Atoi(strings.TrimPrefix(line, "Number of routines waiting on mutexes: "))
				if err != nil {
					utils.LogError("Failed to read number of waiting mutexes:", err.Error())
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
