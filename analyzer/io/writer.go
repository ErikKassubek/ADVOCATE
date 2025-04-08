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
	"analyzer/analysis"
	"analyzer/timer"
	"analyzer/utils"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

/*
 * Write the trace to a file
 * Args:
 *   traceToWrite (*analysis.Trace): Pointer to the trace to write
 *   path (string): The path to the file to write to
 */
func WriteTrace(traceToWrite *analysis.Trace, path string) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	// delete folder if exists
	if _, err := os.Stat(path); err == nil {
		utils.LogInfo(path + " already exists. Delete folder " + path)
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
				utils.LogError("Error in writing trace to file. Could not open file: ", err.Error())
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
				elementString := element.ToString()
				if _, err := file.WriteString(elementString); err != nil {
					utils.LogError("Error in writing trace to file. Could not write string: ", err.Error())
				}
				if index < len(trace)-1 {
					if _, err := file.WriteString("\n"); err != nil {
						utils.LogError("Error in writing trace to file. Could not wrote string: ", err.Error())
					}
				}
			}
			if _, err := file.WriteString("\n"); err != nil {
				utils.LogError("Error in writing trace to file. Could not wrote string: ", err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

/*
 * In path, create a file with the result message and the exit code for the rewrite
 * Args:
 *   path (string): The path to the file folder to write to
 *   resultMessage (string): The result message
 *   exitCode (int): The exit code
 *   resultIndex (int): The index of the result
 * Returns:
 *   error: The error that occurred
 */
func WriteRewriteInfoFile(path string, bugType string, exitCode int, resultIndex int) error {
	timer.Start(timer.Io)
	defer timer.Stop(timer.Io)

	fileName := path + "rewrite_info.log"
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(strconv.Itoa(resultIndex+1) + "#" + bugType + "#" + strconv.Itoa(exitCode)); err != nil {
		return err
	}

	return nil
}
