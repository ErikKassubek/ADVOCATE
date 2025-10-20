// Copyright (c) 2024 Erik Kassubek
//
// File: statsAnalyzer.go
// Brief: Collect stats about the analysis and the replay
//
// Author: Erik Kassubek
// Created: 2024-09-20
//
// License: BSD-3-Clause

package stats

import (
	"advocate/results/explanation"
	"advocate/utils/log"
	"advocate/utils/paths"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getNewDataMap provides a new map to store the analyzer stats.
// It has the form bugTypeID -> counter
//
// Returns:
//   - map[string]int: The new map
func getNewDataMap() map[string]int {
	keys := []string{
		"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "A09",
		"P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02",
		"L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "L11",
		"R01", "R02"}

	m := make(map[string]int)
	for _, key := range keys {
		m[key] = 0
	}

	return m
}

// getNewDataMapMap provides a map used for collecting statistics of the analysis.
// The fields are detected, replayWritten, replaySuccessful, unexpectedPanic.
// Each field contains a data map as created by getNewDataMap()
//
// Returns:
//   - map[string]map[string]int: The map
func getNewDataMapMap() map[string]map[string]int {
	return map[string]map[string]int{
		"detected":         getNewDataMap(),
		"replayWritten":    getNewDataMap(),
		"replaySuccessful": getNewDataMap(),
		"unexpectedPanic":  getNewDataMap(),
	}
}

// Parse the analyzer and replay output to collect the corresponding information
//
// Parameter:
//   - fuzzing int: number of fuzzing run, -1 for not fuzzing
//
// Returns:
//   - map[string]int: map with total information
//   - map[string]int: map with unique information
//   - error
func statsAnalyzer(fuzzing int) (map[string]map[string]int, map[string]map[string]int, error) {
	// reset foundBugs
	foundBugs := make(map[string]processedBug)

	resUnique := getNewDataMapMap()

	resTotal := getNewDataMapMap()

	bugs := filepath.Join(paths.CurrentResult, "bugs")
	_, err := os.Stat(bugs)
	if os.IsNotExist(err) {
		return resUnique, nil, nil
	}

	err = filepath.Walk(bugs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if fuzzing == -1 {
			if strings.HasPrefix(info.Name(), "bug_") ||
				strings.HasPrefix(info.Name(), "diagnostics_") ||
				strings.HasPrefix(info.Name(), "leak_") {
				err := processBugFile(path, foundBugs, resTotal, resUnique)
				if err != nil {
					log.Error(err)
				}
			}
		} else {
			if strings.HasPrefix(info.Name(), "bug_"+strconv.Itoa(fuzzing)+"_") ||
				strings.HasPrefix(info.Name(), "diagnostics_"+strconv.Itoa(fuzzing)+"_") ||
				strings.HasPrefix(info.Name(), "leak_"+strconv.Itoa(fuzzing)+"_") {
				err := processBugFile(path, foundBugs, resTotal, resUnique)
				if err != nil {
					log.Error(err)
				}
			}
		}

		for _, bug := range foundBugs {
			resUnique["detected"][bug.bugType]++

			if bug.replayWritten {
				resUnique["replayWritten"][bug.bugType]++
			}

			if bug.replaySuc {
				resUnique["replaySuccessful"][bug.bugType]++
			}
		}

		return nil
	})

	return resTotal, resUnique, err
}

// Store a bug that has been processed in the statistics.
// Used to count number of unique bugs
// Properties:
//
//   - paths []string: list of paths to each element involved in the bug
//   - bugType string: ID of the bug type
//   - replayWritten bool: true if a replay trace was created for the bug
//   - replaySuc bool: true if the replay of the bug was successful
type processedBug struct {
	paths         []string
	bugType       string
	replayWritten bool
	replaySuc     bool
}

// Get a string representation of a bug
//
// Returns:
//   - string: string representation of the bug
func (this *processedBug) getKey() string {
	res := this.bugType
	for _, path := range this.paths {
		res += path
	}
	return res
}

// Parse a bug file to get the information
//
// Parameter:
//   - filePath string: path to the bug file
//   - resTotal map[string]map[string]int: total results
//   - resUnique map[string]map[string]int: unique results
//
// Returns:
//   - error
func processBugFile(filePath string, foundBugs map[string]processedBug,
	resTotal map[string]map[string]int, resUnique map[string]map[string]int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bugType := ""

	bug := processedBug{}
	bug.paths = make([]string, 0)

	// read the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// get detected bug
		if strings.HasPrefix(line, "# ") {
			textSplit := strings.Split(line, ": ")
			if len(textSplit) != 2 {
				continue
			}

			line = textSplit[1]

			bugType = explanation.GetCodeFromDescription(line)
			if bugType == "" {
				return fmt.Errorf("Unknown error type %s", line)
			}
			bug.bugType = bugType
		} else if strings.HasPrefix(line, "-> ") { // get paths
			bug.paths = append(bug.paths, strings.TrimPrefix(line, "-> "))
		} else if strings.Contains(line, "The analyzer found a way to resolve the leak") {
			bug.replayWritten = true
		} else if strings.Contains(line, "The analyzer has tries to rewrite the trace in such a way") {
			bug.replayWritten = true
		} else if strings.HasPrefix(line, "It exited with the following code: ") {
			code := strings.TrimPrefix(line, "It exited with the following code: ")

			num, err := strconv.Atoi(code)
			if err != nil {
				num = -1
			}

			if num == 3 {
				(resUnique)["unexpectedPanic"][bugType]++
				if resTotal != nil {
					(resTotal)["unexpectedPanic"][bugType]++
				}
			}

			if num >= 20 {
				bug.replaySuc = true
			}
		}
	}

	if bug.bugType == "" {
		return fmt.Errorf("Invalid bug file")
	}

	if resTotal != nil {
		(resTotal)["detected"][bugType]++

		if bug.replayWritten {
			(resTotal)["replayWritten"][bugType]++
		}

		if bug.replaySuc {
			(resTotal)["replaySuccessful"][bugType]++
		}
	}

	key := bug.getKey()
	if b, ok := foundBugs[key]; ok {
		if bug.replaySuc {
			b.replaySuc = true
		}
		if bug.replayWritten {
			b.replayWritten = true
		}
		foundBugs[key] = b
	} else {
		foundBugs[key] = bug
	}

	return nil
}
