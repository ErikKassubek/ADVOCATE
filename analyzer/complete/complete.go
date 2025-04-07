// Copyright (c) 2024 Erik Kassubek
//
// File: complete.go
// Brief: Functions to check if all program elements have been executed at least once
//
// Author: Erik Kassubek
// Created: 2024-06-26
//
// License: BSD-3-Clause

package complete

import (
	"analyzer/explanation"
	"analyzer/utils"
	"fmt"
	"os"
	"strings"
)

/*
 * Check if all program elements are in trace
 * Args:
 * 	resultFolderPath: path to the folder containing the trace files
 * 	progPath: path to the program file
 * Returns:
 * 	error: error if any
 */
func Check(resultFolderPath string, progPath string) error {
	progElems, err := getProgramElements(progPath)
	if err != nil {
		println("Error in getProgramElements")
		return err
	}

	traceElems, err := getTraceElements(resultFolderPath)
	if err != nil {
		println("Error in getTraceElements")
		return err
	}

	notInTrace := areAllProgElemInTrace(progElems, traceElems)
	notSelectedSelectCase := getNotSelectedSelectCases()

	err = printNotExecutedToFiles(notInTrace, notSelectedSelectCase, resultFolderPath)

	return err
}

func areAllProgElemInTrace(progElems map[string][]int, traceElems map[string][]int) map[string][]int {
	res := map[string][]int{}

	for file, lines := range progElems {
		// file not recorded in trace
		if _, ok := traceElems[file]; !ok {
			if _, ok := res[file]; !ok {
				res[file] = make([]int, 0)
			}

			res[file] = append(res[file], -1) // -1 signaling, that no element in file was in trace
			res[file] = append(res[file], lines...)
		}

		for _, line := range lines {
			if !utils.ContainsInt(traceElems[file], line) {
				if _, ok := res[file]; !ok {
					res[file] = make([]int, 0)
				}

				res[file] = append(res[file], line)
			}
		}
	}

	return res
}

/*
 * GetNotSelectedSelectCases prints the elements and select cases that were not executed
 * into a file.
 * Args:
 * 	elements: the elements that were not executed
 * 	selects: the select cases that were not selected
 */
func printNotExecutedToFiles(elements map[string][]int, selects map[string]map[int][]int,
	path string) error {

	path = fmt.Sprintf("%s/AdvocateNotExecuted", path)

	// create a folder to store results
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	pathOperations := fmt.Sprintf("%s/AdvocateNotExecutedOperations.txt", path)
	pathSelects := fmt.Sprintf("%s/AdvocateNotExecutedSelectCases.txt", path)

	notExecutedOperationsFile, err := os.Create(pathOperations)
	if err != nil {
		return err
	}
	defer notExecutedOperationsFile.Close()

	notExecutedSelectFile, err := os.Create(pathSelects)
	if err != nil {
		return err
	}
	defer notExecutedSelectFile.Close()

	pathOperationsFolder := fmt.Sprintf("%s/Operations", path)
	err = os.MkdirAll(pathOperationsFolder, os.ModePerm)
	if err != nil {
		return err
	}

	// write elements that were not executed
	if len(elements) > 0 {
		for file, lines := range elements {
			fileName := strings.ReplaceAll(file, "/", "_")
			fileName = strings.TrimPrefix(fileName, "_")
			pathFile := fmt.Sprintf("%s/%s.md", pathOperationsFolder, fileName)
			fileFile, err := os.Create(pathFile)
			fileFile.WriteString(fmt.Sprintf("# %s\n", file))
			fileFile.WriteString("## Not executed operations\n")
			if err != nil {
				return err
			}

			notExecutedOperationsFile.WriteString(fmt.Sprintf("%s:[", file))
			for i, line := range lines {
				if line == -1 {
					notExecutedOperationsFile.WriteString("No element in file was executed")
					fileFile.Write([]byte("No element in file was executed"))
					break
				} else {
					notExecutedOperationsFile.WriteString(fmt.Sprintf("%d", line))
					if i != len(lines)-1 {
						notExecutedOperationsFile.WriteString(",")
					}
					fileFile.WriteString(fmt.Sprintf("### Line: %d\n", line))
					code, err := explanation.GetProgramCode(file, line, true)
					if err != nil {
						fileFile.WriteString("Error reading file: " + err.Error() + "\n")
						continue
					}
					fileFile.WriteString(code)
				}
			}
			notExecutedOperationsFile.WriteString("]\n")
			fileFile.Close()
		}
	} else {
		notExecutedOperationsFile.WriteString("All program elements were executed\n")
	}

	// write select cases that were not selected
	if len(selects) > 0 {
		for file, lines := range selects {
			fileName := strings.ReplaceAll(file, "/", "_")
			fileName = strings.TrimPrefix(fileName, "_")
			pathFile := fmt.Sprintf("%s/%s.md", pathOperationsFolder, fileName)
			// if file does not exist, create it otherwise append to it
			fileFile, err := os.OpenFile(pathFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			fileFile.WriteString(fmt.Sprintf("## Not selected select cases\n"))

			for line, cases := range lines {
				notExecutedSelectFile.WriteString(fmt.Sprintf("%s:%d:[", file, line))
				fileFile.WriteString(fmt.Sprintf("### Line: %d\n", line))
				for i, c := range cases {
					if c == -1 {
						notExecutedSelectFile.WriteString("D")
						fileFile.WriteString("Default case was never selected\n")
					} else {
						notExecutedSelectFile.WriteString(fmt.Sprintf("%d", c))
						fileFile.WriteString(fmt.Sprintf("Case: %d was never selected\n", c))
					}

					if i != len(cases)-1 {
						notExecutedSelectFile.WriteString(",")
					}

					code, err := explanation.GetProgramCode(file, line, true)
					if err != nil {
						fileFile.WriteString("Error reading file: " + err.Error() + "\n")
						continue
					}
					fileFile.WriteString(code)
				}
				notExecutedSelectFile.WriteString("]\n")
			}
			fileFile.Close()
		}
	} else {
		notExecutedSelectFile.WriteString("All select cases were executed\n")
	}

	return nil
}
