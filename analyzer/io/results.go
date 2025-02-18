// Copyrigth (c) 2024 Erik Kassubek
//
// File: results.go
// Brief: Read the analysis result file and analyze its content
//
// Author: Erik Kassubek
// Created: 2023-11-30
//
// License: BSD-3-Clause

package io

import (
	"analyzer/bugs"
	"analyzer/utils"
	"bufio"
	"fmt"
	"os"
)

/*
 * Read the fail containing the output of the analysis
 * Extract the needed information to create a trace to replay the selected error
 * Args:
 *   filePath (string): The path to the file containing the analysis results
 *   index (int): The index of the result to create a trace for (0 based)
 * Returns:
 *   bool: true, if the bug was not a possible, but an actually occuring bug
 *   Bug: The bug that was selected
 *   error: An error if the bug could not be processed
 */
func ReadAnalysisResults(filePath string, index int) (bool, bugs.Bug, error) {
	bugStr := ""

	file, err := os.Open(filePath)
	if err != nil {
		utils.LogError("Error opening file: " + filePath)
		return false, bugs.Bug{}, err
	}

	scanner := bufio.NewScanner(file)

	i := 0
	for scanner.Scan() {
		bugStr = scanner.Text()
		if index == i {
			break
		}
		i++

		if err := scanner.Err(); err != nil {
			println("Error reading file line.")
			break
		}
	}

	if bugStr == "" {
		return false, bugs.Bug{}, fmt.Errorf("Empty bug string")
	}

	actual, bug, err := bugs.ProcessBug(bugStr)
	if err != nil {
		utils.LogError("Error processing bug: ", bugStr)
		utils.LogError(err.Error())
		return false, bug, err
	}

	if actual {
		return true, bug, nil
	}

	return false, bug, nil

}
