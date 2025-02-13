// Copyright (c) 2024 Erik Kassubek
//
// File: statsTrace.go
// Brief: Collect statistics about the trace
//
// Author: Erik Kassubek
// Created: 2024-09-20
//
// License: BSD-3-Clause

package stats

import (
	"analyzer/utils"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Collect stats about the traces
 * Args:
 *     dataPath (string): path to the result folder
 * Returns:
 *     map[string]int: map with the stats
 *     error
 */
func statsTraces(dataPath string) (map[string]int, error) {
	res := map[string]int{
		"numberElements": 0,

		"numberRoutines":         0,
		"numberNonEmptyRoutines": 0,

		"numberOfSpawns":    0,
		"numberRoutineEnds": 0,

		"numberAtomics":          0,
		"numberAtomicOperations": 0,

		"numberChannels":           0,
		"numberBufferedChannels":   0,
		"numberUnbufferedChannels": 0,
		"numberChannelOperations":  0,
		"numberBufferedOps":        0,
		"numberUnbufferedOps":      0,

		"numberSelects":          0,
		"numberSelectCases":      0,
		"numberSelectChanOps":    0, // number of executed channel operations in select
		"numberSelectDefaultOps": 0, // number of executed default operations in select

		"numberMutexes":         0,
		"numberMutexOperations": 0,

		"numberWaitGroups":          0,
		"numberWaitGroupOperations": 0,

		"numberCondVars":          0,
		"numberCondVarOperations": 0,

		"numberOnce":           0,
		"numberOnceOperations": 0,
	}

	tracePath := filepath.Join(dataPath, "advocateTrace")

	// do not count the same twice
	known := map[string][]string{
		"atomic":    []string{},
		"channel":   []string{},
		"mutex":     []string{},
		"waitGroup": []string{},
		"condVar":   []string{},
		"once":      []string{},
	}

	err := filepath.Walk(tracePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() != "trace_info.log" {
			err = parseTraceFile(path, &res, &known)
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})

	return res, err
}

func parseTraceFile(tracePath string, stats *map[string]int, known *map[string][]string) error {
	// open the file
	file, err := os.Open(tracePath)
	if err != nil {
		return err
	}

	// routine, err := getRoutineFromFileName(filepath.Base(tracePath))
	// if err != nil {
	// 	return fmt.Errorf("%s not an trace file", tracePath)
	// }
	(*stats)["numberRoutines"]++

	scanner := bufio.NewScanner(file)

	// read the file
	foundNonEmpty := false
	for scanner.Scan() {
		elem := scanner.Text()

		if elem != "" && !foundNonEmpty {
			(*stats)["numberNonEmptyRoutines"]++
			foundNonEmpty = true
		}
		(*stats)["numberElements"]++
		fields := strings.Split(elem, ",")
		switch fields[0] {
		case "G":
			(*stats)["numberOfSpawns"]++
		case "A":
			(*stats)["numberAtomicOperations"]++
			if !utils.ContainsString((*known)["atomic"], fields[2]) {
				(*stats)["numberAtomics"]++
				(*known)["atomic"] = append((*known)["atomic"], fields[2])
			}
		case "C":
			(*stats)["numberChannelOperations"]++
			if fields[7] == "0" {
				(*stats)["numberUnbufferedOps"]++
			} else {
				(*stats)["numberBufferedOps"]++
			}
			if !utils.ContainsString((*known)["channel"], fields[3]) {
				(*stats)["numberChannels"]++
				if fields[7] == "0" {
					(*stats)["numberUnbufferedChannels"]++
				} else {
					(*stats)["numberBufferedChannels"]++
				}
				(*known)["channel"] = append((*known)["channel"], fields[3])
			}
		case "S":
			(*stats)["numberSelects"]++
			cases := strings.Split(fields[4], "~")
			(*stats)["numberSelectCases"] += len(cases)
			if cases[len(cases)-1] == "D" {
				(*stats)["numberSelectDefaultOps"]++
			} else {
				(*stats)["numberSelectChanOps"] += len(cases)
			}
		case "M":
			(*stats)["numberMutexOperations"]++
			if !utils.ContainsString((*known)["mutex"], fields[3]) {
				(*stats)["numberMutexes"]++
				(*known)["mutex"] = append((*known)["mutex"], fields[3])
			}
		case "W":
			(*stats)["numberWaitGroupOperations"]++
			if !utils.ContainsString((*known)["waitGroup"], fields[3]) {
				(*stats)["numberWaitGroups"]++
				(*known)["waitGroup"] = append((*known)["waitGroup"], fields[3])
			}
		case "O":
			(*stats)["numberOnceOperations"]++
			if !utils.ContainsString((*known)["once"], fields[3]) {
				(*stats)["numberOnce"]++
				(*known)["once"] = append((*known)["once"], fields[3])
			}
		case "D":
			(*stats)["numberCondVarOperations"]++
			if !utils.ContainsString((*known)["condVar"], fields[3]) {
				(*stats)["numberCondVars"]++
				(*known)["condVar"] = append((*known)["condVar"], fields[3])
			}
		case "E":
			(*stats)["numberRoutineEnds"]++
		case "N":
			// do notring
		default:
			err = errors.New("Unknown trace element: " + fields[0])
		}
	}
	return err
}

func getRoutineFromFileName(fileName string) (int, error) {
	// the file name is "trace_routineID.log"
	// remove the .log at the end
	fileName1 := strings.TrimSuffix(fileName, ".log")
	if fileName1 == fileName {
		return 0, errors.New("File name does not end with .log")
	}

	fileName2 := strings.TrimPrefix(fileName1, "trace_")
	if fileName2 == fileName1 {
		return 0, errors.New("File name does not start with trace_")
	}

	routine, err := strconv.Atoi(fileName2)
	if err != nil {
		return 0, err
	}

	return routine, nil
}
