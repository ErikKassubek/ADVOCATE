// Copyright (c) 2024 Erik Kassubek
//
// File: statsProgram.go
// Brief: Collect statistics about the program
//
// Author: Erik Kassubek
// Created: 2024-09-20
//
// License: BSD-3-Clause

package stats

import (
	"advocate/utils/log"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateStatsTotal creates an overview over all statistics for a program
// (not split by test/run)
//
// Parameter:
//   - pathFolder string: path the where the stat files should be created
//   - progName string: name of the analyzed program
func CreateStatsTotal(pathFolder, progName string) error {
	resultPath := filepath.Join(pathFolder, "advocateResult")
	statsAnalyzerPath := filepath.Join(resultPath, "statsAnalysis_"+progName+".csv")
	statsTotalPath := filepath.Join(resultPath, "statsProgram_"+progName+".csv")

	log.Info("Create program statistics")

	progData, err := statsProgram(pathFolder)
	if err != nil {
		log.Error("Could not collect program info: ", err.Error())
	}

	// get the number of tests and traces from statsAnalysis
	analysisFile, err := os.Open(statsAnalyzerPath)
	if err != nil {
		return err
	}
	defer analysisFile.Close()

	scanner := bufio.NewScanner(analysisFile)

	// skip the first line
	scanner.Scan()

	noRuns := 0
	noTests := 0
	lastName := ""
	for scanner.Scan() {
		testName := strings.Split(scanner.Text(), ",")[0]
		if testName == "" {
			continue
		}

		noRuns++
		if lastName != testName {
			noTests++
			lastName = testName
		}
	}

	// get the found bugs
	foundBugs := make(map[string]processedBug)
	data := getNewDataMapMap()

	err = filepath.Walk(resultPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if !strings.Contains(path, "/bugs/") {
			return nil
		}

		if !(strings.HasPrefix(info.Name(), "leak_") ||
			strings.HasPrefix(info.Name(), "bug_") ||
			strings.HasPrefix(info.Name(), "diagnostics_")) {
			return nil
		}

		return processBugFile(path, foundBugs, nil, data)
	})

	for _, bug := range foundBugs {
		data["detected"][bug.bugType]++

		if bug.replayWritten {
			data["replayWritten"][bug.bugType]++
		}

		if bug.replaySuc {
			data["replaySuccessful"][bug.bugType]++
		}
	}

	if err != nil {
		return err
	}

	headers := "NrFiles,NrLines,NrNonEmptyLines,NrTests,NrRuns"

	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "R01", "R02"} {
			headers += fmt.Sprintf(",Nr%s%s", strings.ToUpper(string(mode[0]))+mode[1:], code)
		}
	}
	headers += "\n"

	res := ""

	if progData != nil {
		res += fmt.Sprintf("%d,%d,%d,", progData["numberFiles"], progData["numberLines"], progData["numberNonEmptyLines"])
	} else {
		res += "0,0,0,"
	}

	res += fmt.Sprintf("%d,%d", noTests, noRuns)

	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "R01", "R02"} {
			res += fmt.Sprintf(",%d", data[mode][code])
		}
	}

	// write to file
	fuzzingFile, err := os.OpenFile(statsTotalPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fuzzingFile.Close()

	fuzzingFile.WriteString(headers)
	fuzzingFile.WriteString(res)

	return err
}

// Parse a program to measure the number of files, and number of lines
//
// Parameter:
//   - programPath string: path to the folder containing the program
//
// Returns:
//   - map[string]int: map with numberFiles, numberLines, numberNonEmptyLines
//   - error
func statsProgram(programPath string) (map[string]int, error) {
	res := make(map[string]int)
	res["numberFiles"] = 0
	res["numberLines"] = 0
	res["numberNonEmptyLines"] = 0

	err := filepath.Walk(programPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".go" {
			resFile, err := parseProgramFile(path)
			if err != nil {
				return err
			}

			res["numberFiles"]++
			res["numberLines"] += resFile["numberLines"]
			res["numberNonEmptyLines"] += resFile["numberNonEmptyLines"]
		}

		return nil
	})
	return res, err
}

// Parse one program file to measure the number of lines
//
// Parameter:
//   - programPath string: path to the file
//
// Returns:
//   - map[string]int: map with numberLines, numberNonEmptyLines
//   - error
func parseProgramFile(filePath string) (map[string]int, error) {
	res := make(map[string]int)
	res["numberLines"] = 0
	res["numberNonEmptyLines"] = 0

	// open the file
	file, err := os.Open(filePath)
	if err != nil {
		return res, err
	}
	defer file.Close()

	// read the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())

		res["numberLines"]++
		if text != "" && text != "\n" && !strings.HasPrefix(text, "//") {
			res["numberNonEmptyLines"]++
		}
	}

	if err := scanner.Err(); err != nil {
		return res, err
	}

	return res, nil
}
