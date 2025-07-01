// Copyright (c) 2024 Erik Kassubek
//
// File: writer.go
// Brief: Write the internal trace into files
//
// Author: Erik Kassubek
// Created: 2023-12-01
//
// License: BSD-3-Clause

package io

import (
	"advocate/results/bugs"
	"advocate/trace"
	"advocate/utils/helper"
	"advocate/utils/log"
	"advocate/utils/timer"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

// WriteTrace writes a trace to a file
//
// Parameter:
//   - traceToWrite *analysis.Trace: Pointer to the trace to write
//   - path string: The path to the file to write to
//   - replay bool: If true, write only the elements relevant for replay
func WriteTrace(traceToWrite *trace.Trace, path string, replay bool) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// delete folder if exists
	if _, err := os.Stat(path); err == nil {
		log.Info(path + " already exists. Delete folder " + path)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	// create new folder
	if err := os.Mkdir(path, 0755); err != nil {
		return err
	}

	// write the files

	numberRoutines := traceToWrite.GetNoRoutines()

	wg := sync.WaitGroup{}
	for i := 1; i <= numberRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			fileName := filepath.Join(path, "trace_"+strconv.Itoa(i)+".log")
			file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Error("Error in writing trace to file. Could not open file: ", err.Error())
			}
			defer file.Close()

			// write trace
			// println("Write trace to " + fileName + "...")
			trace := traceToWrite.GetRoutineTrace(i)

			// sort trace by tPre
			sort.Slice(trace, func(i, j int) bool {
				return trace[i].GetTPre() < trace[j].GetTPre()
			})

			for index, element := range trace {
				if !replay || !isReplay(element) {
					continue
				}
				elementString := element.ToString()
				if _, err := file.WriteString(elementString); err != nil {
					log.Error("Error in writing trace to file. Could not write string: ", err.Error())
				}
				if index < len(trace)-1 {
					if _, err := file.WriteString("\n"); err != nil {
						log.Error("Error in writing trace to file. Could not wrote string: ", err.Error())
					}
				}
			}
			if _, err := file.WriteString("\n"); err != nil {
				log.Error("Error in writing trace to file. Could not wrote string: ", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

// Check if the element is relevant for replay
//
// Parameter:
//   - element trace.TraceElement: element to check
//
// Returns:
//   - true if relevant for replay, false if ignored in replay
func isReplay(element trace.TraceElement) bool {
	t := element.GetObjType(false)
	return !(t == trace.ObjectTypeNew || t == trace.ObjectTypeRoutineEnd)
}

// WriteRewriteInfoFile create a file with the result message and the exit code for the rewrite
//
// Parameter:
//   - path string: The path to the file folder to write to
//   - bug bugs.Bug: The rewritten bug
//   - exitCode int: The exit code
//   - resultIndex int: The index of the result
//
// Returns:
//   - error: The error that occurred
func WriteRewriteInfoFile(path string, bug bugs.Bug, exitCode int, resultIndex int) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	fileName := filepath.Join(path, helper.RewrittenInfo)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(strconv.Itoa(resultIndex+1) + "#" + bug.GetBugString() + "#" + strconv.Itoa(exitCode)); err != nil {
		return err
	}

	return nil
}
