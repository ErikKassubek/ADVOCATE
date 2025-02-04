// Copyrigth (c) 2024 Erik Kassubek
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
	"io"
	"os"
	"sort"
	"strconv"
	"sync"
)

/*
 * Copy a file from source to dest
 * Args:
 *   source (string): The path to the source file
 *   dest (string): The path to the destination file
 */
func CopyFolder(source string, dest string) {
	sourceFile, err := os.Open(source)
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		panic(err)
	}

	err = destFile.Sync()
	if err != nil {
		panic(err)
	}
}

/*
 * Write the trace to a file
 * Args:
 *   path (string): The path to the file to write to
 *   numberRoutines (int): The number of routines in the trace
 */
func WriteTrace(path string, numberRoutines int) error {

	// delete folder if exists
	if _, err := os.Stat(path); err == nil {
		println(path + " already exists. Delete folder " + path)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	println("Create new trace at " + path)

	// create new folder
	if err := os.Mkdir(path, 0755); err != nil {
		return err
	}

	// write the files
	wg := sync.WaitGroup{}
	for i := 1; i <= numberRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			fileName := path + "trace_" + strconv.Itoa(i) + ".log"
			// println("Create new file " + fileName + "...")
			file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			// write trace
			// println("Write trace to " + fileName + "...")
			trace := analysis.GetTraceFromId(i)

			// sort trace by tPre
			sort.Slice(trace, func(i, j int) bool {
				return trace[i].GetTPre() < trace[j].GetTPre()
			})

			for index, element := range trace {
				elementString := element.ToString()
				if _, err := file.WriteString(elementString); err != nil {
					panic(err)
				}
				if index < len(trace)-1 {
					if _, err := file.WriteString("\n"); err != nil {
						panic(err)
					}
				}
			}
			if _, err := file.WriteString("\n"); err != nil {
				panic(err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	println("Trace written")
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
