// Copyright (c) 2024 Erik Kassubek
//
// File: reader.go
// Brief: Read trace files and create the internal trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package io

import (
	"analyzer/analysis"
	"analyzer/timer"
	"analyzer/utils"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Create the trace from all files in a folder.
//
// Parameter:
//   - filePath string: The path to the folder
//   - ignoreAtomics bool: If atomic operations should be ignored
//
// Returns:
//   - int: The number of routines
//   - int: The number of elements
//   - error: An error if the trace could not be created
func CreateTraceFromFiles(folderPath string, ignoreAtomics bool) (int, int, error) {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	numberRoutines := 0
	// traverse all files in the folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, 0, err
	}

	analysis.SetMainTraceToNewTrace()

	elemCounter := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name() == "times.log" {
			continue
		}

		filePath := filepath.Join(folderPath, file.Name())

		if file.Name() == "trace_info.log" {
			getTraceInfoFromFile(filePath)
		}

		routine, err := getRoutineFromFileName(file.Name())
		if err != nil {
			continue
		}
		numberRoutines = max(numberRoutines, routine)

		numberElems, err := createTraceFromFile(filePath, routine, ignoreAtomics)
		if err != nil {
			return 0, elemCounter, err
		}
		elemCounter += numberElems
	}

	analysis.SortTrace()

	return numberRoutines, elemCounter, nil
}

// getTraceInfoFromFile reads in the information from a the trace_info.log file
//
// Parameter:
//   - filePath string: the path to the trace_info.log file
//
// Returns:
//   - error
func getTraceInfoFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		utils.LogError("Error opening file: " + filePath)
		return err
	}

	scanner := bufio.NewScanner(file)

	exitCode := 0
	exitPos := ""

	timeoutOldest := 0
	timeoutDisabled := 0
	timeoutAck := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineSplit := strings.Split(line, "!")
		if len(lineSplit) != 2 {
			continue
		}

		switch lineSplit[0] {
		case "Runtime":
			rt, err := strconv.Atoi(lineSplit[1])
			if err == nil {
				analysis.SetRuntimeDurationSec(rt)
			}
		case "ExitCode":
			ec, err := strconv.Atoi(lineSplit[1])
			if err != nil {
				exitCode = ec
			}
		case "ExitPosition":
			exitPos = lineSplit[1]
		case "ReplayTimeout":
			timeoutOldest, _ = strconv.Atoi(lineSplit[1])
		case "ReplayDisabled":
			timeoutDisabled, _ = strconv.Atoi(lineSplit[1])
		case "ReplayAck":
			timeoutAck, _ = strconv.Atoi(lineSplit[1])
		}
	}

	analysis.SetExitInfo(exitCode, exitPos)
	analysis.SetReplayTimeoutInfo(timeoutOldest, timeoutDisabled, timeoutAck)

	return nil
}

// Read and build the trace from a file
//
// Parameter:
//   - filePath string: The path to the log file
//   - routine int: The routine id
//   - ignoreAtomics bool: If atomic operations should be ignored
//
// Returns:
//   - - int: number of elements
//   - error: An error if the trace could not be created
func createTraceFromFile(filePath string, routine int, ignoreAtomics bool) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		utils.LogError("Error opening file: " + filePath)
		return 0, err
	}

	scanner := bufio.NewScanner(file)

	counter := 0
	for scanner.Scan() {
		line := scanner.Text()
		err := processElement(line, routine, ignoreAtomics)
		if err != nil {
			utils.LogError("Error in processing trace element: ", err)
		}
		counter++
	}

	file.Close()

	return counter, scanner.Err()
}

// Process one element from the log file.
//
// Parameter:
//   - element string: The element to process
//   - routine int: The routine id, equal to the line number
//   - ignoreAtomics bool: If atomic operations should be ignored
//
// Returns:
//   - error: An error if the element could not be processed
func processElement(element string, routine int, ignoreAtomics bool) error {
	if element == "" {
		return errors.New("Element is empty")
	}
	fields := strings.Split(element, ",")
	var err error
	switch fields[0] {
	case "A":
		if ignoreAtomics {
			return nil
		}
		err = analysis.AddTraceElementAtomic(routine, fields[1], fields[2], fields[3], fields[4])
	case "C":
		if len(fields) != 10 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 10", element, len(fields))
		}
		err = analysis.AddTraceElementChannel(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7], fields[8], fields[9])
	case "M":
		if len(fields) != 8 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 8", element, len(fields))
		}
		err = analysis.AddTraceElementMutex(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7])
	case "G":
		if len(fields) != 4 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 4", element, len(fields))
		}
		err = analysis.AddTraceElementFork(routine, fields[1], fields[2], fields[3])
	case "S":
		if len(fields) != 7 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = analysis.AddTraceElementSelect(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6])
	case "W":
		if len(fields) != 8 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 8", element, len(fields))
		}
		err = analysis.AddTraceElementWait(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6], fields[7])
	case "O":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = analysis.AddTraceElementOnce(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "D":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = analysis.AddTraceElementCond(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "N":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = analysis.AddTraceElementNew(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "E":
		if len(fields) != 2 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 2", element, len(fields))
		}
		err = analysis.AddTraceElementRoutineEnd(routine, fields[1])
	default:
		return errors.New("Unknown element type in: " + element)
	}

	if err != nil {
		return err
	}

	return nil
}

// getRoutineFromFileName extracts the file ID from a trace file. Trace files
// always have the name trace_[ID]
//
// Parameter:
//   - fileName string: name of the trace file
//
// Returns:
//   - int: if fileName is valid the trace id, otherwise 0
//   - error
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
