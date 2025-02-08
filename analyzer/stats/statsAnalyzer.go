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
	"analyzer/explanation"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Parse the analyzer and replay output to collect the corresponding information
 * Args:
 *     pathToResults (string): path to the advocateResult folder
 * Returns:
 *     map[string]int: map with information
 *     error
 */
func statsAnalyzer(pathToResults string) (map[string]map[string]int, error) {
	detected := map[string]int{
		"A00": 0, "A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "A06": 0, "A07": 0,
		"A08": 0, "P01": 0, "P02": 0, "P03": 0, "P04": 0, "L00": 0, "L01": 0, "L02": 0,
		"L03": 0, "L04": 0, "L05": 0, "L06": 0, "L07": 0, "L08": 0, "L09": 0, "L10": 0}
	replayWriten := map[string]int{
		"A00": 0, "A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "A06": 0, "A07": 0,
		"A08": 0, "P01": 0, "P02": 0, "P03": 0, "P04": 0, "L00": 0, "L01": 0, "L02": 0,
		"L03": 0, "L04": 0, "L05": 0, "L06": 0, "L07": 0, "L08": 0, "L09": 0, "L10": 0}
	replaySuccessful := map[string]int{
		"A00": 0, "A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "A06": 0, "A07": 0,
		"A08": 0, "P01": 0, "P02": 0, "P03": 0, "P04": 0, "L00": 0, "L01": 0, "L02": 0,
		"L03": 0, "L04": 0, "L05": 0, "L06": 0, "L07": 0, "L08": 0, "L09": 0, "L10": 0}
	rerecorded := map[string]int{
		"A00": 0, "A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "A06": 0, "A07": 0,
		"A08": 0, "P01": 0, "P02": 0, "P03": 0, "P04": 0, "L00": 0, "L01": 0, "L02": 0,
		"L03": 0, "L04": 0, "L05": 0, "L06": 0, "L07": 0, "L08": 0, "L09": 0, "L10": 0}
	unexpactedPanic := map[string]int{
		"A00": 0, "A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "A06": 0, "A07": 0,
		"A08": 0, "P01": 0, "P02": 0, "P03": 0, "P04": 0, "L00": 0, "L01": 0, "L02": 0,
		"L03": 0, "L04": 0, "L05": 0, "L06": 0, "L07": 0, "L08": 0, "L09": 0, "L10": 0}

	res := map[string]map[string]int{
		"detected":         detected,
		"replayWritten":    replayWriten,
		"replaySuccessful": replaySuccessful,
		"rerecorded":       rerecorded,
		"unexpectedPanic":  unexpactedPanic,
	}

	bugs := filepath.Join(pathToResults, "bugs")
	_, err := os.Stat(bugs)
	if os.IsNotExist(err) {
		return res, nil
	}

	err = filepath.Walk(bugs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasPrefix(info.Name(), "bug_") ||
			strings.HasPrefix(info.Name(), "diagnostics_") ||
			strings.HasPrefix(info.Name(), "leak_")) {
			err := processBugFile(path, &res)
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})

	return res, err
}

type processedBug struct {
	paths         string
	detectedBug   string
	replayWritten bool
	replaySuc     bool
}

/*
 * Parse a bug file to get the information
 * Args:
 *     filePath (string): path to the bug file
 *     info (*map[string]map[string]int): map to store the info in
 * Returns:
 *     error
 */
func processBugFile(filePath string, info *map[string]map[string]int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bugType := ""

	bug := processedBug{}

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
				return fmt.Errorf("unknown error type %s", line)
			}
			bug.detectedBug = bugType
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
				(*info)["unexpectedPanic"][bugType]++
			}

			if num >= 20 {
				bug.replaySuc = true
			}
		}
	}

	(*info)["detected"][bugType]++

	if bug.replayWritten {
		(*info)["replayWritten"][bugType]++
	}

	if bug.replaySuc {
		(*info)["replaySuccessful"][bugType]++
	}

	if !strings.Contains(filepath.Base(filePath), "bug_0") {
		(*info)["rerecorded"][bugType]++
	}

	return nil
}
