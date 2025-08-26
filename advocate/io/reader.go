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
	"advocate/analysis/data"
	"advocate/trace"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/log"
	"advocate/utils/timer"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CreateTraceFromFiles creates the trace from all files in a folder.
//
// Parameter:
//   - filePath string: The path to the folder
//
// Returns:
//   - int: The number of routines
//   - int: The number of elements
//   - error: An error if the trace could not be created
func CreateTraceFromFiles(folderPath string) (int, int, error) {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	numberRoutines := 0
	// traverse all files in the folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, 0, err
	}

	tr := trace.NewTrace()

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

		numberElems, err := createTraceFromFile(&tr, filePath, routine)
		if err != nil {
			return 0, elemCounter, err
		}
		elemCounter += numberElems
		numberRoutines++

		if elemCounter > flags.MaxNumberElements {
			return numberRoutines, elemCounter, fmt.Errorf("To many elements")
		}

		if control.CheckCanceled() {
			return numberRoutines, elemCounter, fmt.Errorf("Canceled by memory")
		}
	}

	tr.Sort()
	data.SetMainTrace(&tr)

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
		log.Error("Error opening file: " + filePath)
		return err
	}

	scanner := bufio.NewScanner(file)

	exitCode := 0
	exitPos := ""

	timeoutOldest := 0
	timeoutDisabled := 0
	timeoutAck := 0
	activeReleased := 0
	allActiveReleased := 0

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
				data.SetRuntimeDurationSec(rt)
			}
		case "ExitCode":
			ec, err := strconv.Atoi(lineSplit[1])
			if err == nil {
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
		case "ActiveReached":
			activeReplayReached, _ := strconv.Atoi(lineSplit[1])
			if activeReplayReached > 0 {
				activeReleased = 1
			}
		case "AllActiveReleased":
			allActiveReleased, _ = strconv.Atoi(lineSplit[1])
		}
	}

	data.SetExitInfo(exitCode, exitPos)
	data.SetReplayInfo(timeoutOldest, timeoutDisabled, timeoutAck, activeReleased, allActiveReleased)

	return nil
}

// Read and build the trace from a file
//
// Parameter:
//   - tr *trace.Trace: the trace to add the elements to
//   - filePath string: The path to the log file
//   - routine int: The routine id
//
// Returns:
//   - int: number of elements
//   - error: An error if the trace could not be created
func createTraceFromFile(tr *trace.Trace, filePath string, routine int) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Error opening file: " + filePath)
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	tr.AddRoutine(routine)

	counter := 0
	for scanner.Scan() {
		line := scanner.Text()
		err := processElement(tr, line, routine)
		if err != nil {
			log.Error("Error in processing trace element: ", err)
		}
		counter++
	}

	return counter, scanner.Err()
}

// Process one element from the log file.
//
// Parameter:
//   - tr *trace.Trace: the trace to add the elements to
//   - element string: The element to process
//   - routine int: The routine id, equal to the line number
//
// Returns:
//   - error: An error if the element could not be processed
func processElement(tr *trace.Trace, element string, routine int) error {
	if element == "" {
		return errors.New("Element is empty")
	}
	fields := strings.Split(element, ",")
	var err error
	switch fields[0] {
	case "A":
		if flags.IgnoreAtomics {
			return nil
		}
		err = tr.AddTraceElementAtomic(routine, fields[1], fields[2], fields[3], fields[4])
	case "C":
		if len(fields) != 10 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 10", element, len(fields))
		}
		err = tr.AddTraceElementChannel(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7], fields[8], fields[9])
	case "M":
		if len(fields) != 8 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 8", element, len(fields))
		}
		err = tr.AddTraceElementMutex(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7])
	case "G":
		if len(fields) != 4 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 4", element, len(fields))
		}
		err = tr.AddTraceElementFork(routine, fields[1], fields[2], fields[3])
	case "S":
		if len(fields) != 7 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 7", element, len(fields))
		}
		err = tr.AddTraceElementSelect(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6])
	case "W":
		if len(fields) != 8 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 8", element, len(fields))
		}
		err = tr.AddTraceElementWait(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6], fields[7])
	case "O":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = tr.AddTraceElementOnce(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "D":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = tr.AddTraceElementCond(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "N":
		if len(fields) != 6 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 6", element, len(fields))
		}
		err = tr.AddTraceElementNew(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "E":
		if len(fields) != 2 {
			return fmt.Errorf("Invalid element: %s. Len: %d. Expected len: 2", element, len(fields))
		}
		err = tr.AddTraceElementRoutineEnd(routine, fields[1])
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
