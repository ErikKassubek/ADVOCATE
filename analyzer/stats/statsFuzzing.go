// Copyright (c) 2025 Erik Kassubek
//
// File: statsFuzzing.go
// Brief: Create stats about fuzzing
//
// Author: Erik Kassubek
// Created: 2025-02-17
//
// License: BSD-3-Clause

package stats

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TODO: for each test get the number of unique bugs
func CreateStatsFuzzing(pathFolder, progName string) error {
	// collect the info from the analyzer
	resultPath := filepath.Join(pathFolder, "advocateResult")
	statsAnalyzerPath := filepath.Join(resultPath, "statsAnalysis_"+progName+".csv")
	statsFuzzingPath := filepath.Join(resultPath, "statsFuzzing_"+progName+".csv")

	log.Println("Create fuzzing statistics")

	headers := "TestName,NoRuns"

	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, code := range []string{"A00", "A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10"} {
			headers += fmt.Sprintf(",No%s%s", strings.ToUpper(string(mode[0]))+mode[1:], code)
		}
	}
	headers += "\n"

	data := make(map[string]testData)

	lastTestName := ""
	counter := 0

	// get the test names and number of fuzzing runs

	analysisFile, err := os.Open(statsAnalyzerPath)
	if err != nil {
		return err
	}
	defer analysisFile.Close()

	scanner := bufio.NewScanner(analysisFile)

	// skip the first line
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		elems := strings.Split(line, ",")
		if len(elems) == 0 {
			continue
		}

		testName := elems[0]
		if lastTestName == testName {
			counter++
		} else {
			if lastTestName != "" {
				td := testData{name: lastTestName, numberRuns: counter}
				data[lastTestName] = td
			}
			lastTestName = testName
			counter = 1
		}
	}

	if lastTestName != "" {
		td := testData{name: lastTestName, numberRuns: counter}
		data[lastTestName] = td
	}

	// get the number of unique bugs for this test
	testDir, err := os.ReadDir(resultPath)
	if err != nil {
		return err
	}

	for _, test := range testDir {
		if !test.IsDir() {
			continue
		}

		testNameSplit := strings.Split(test.Name(), "-")
		testName := testNameSplit[len(testNameSplit)-1]

		bugDirPath := filepath.Join(resultPath, test.Name(), "bugs")
		bugDir, err := os.ReadDir(bugDirPath)
		if err != nil {
			return err
		}

		foundBugs := make(map[string]processedBug)
		res := getNewDataMapMap()

		for _, bug := range bugDir {
			processBugFile(filepath.Join(bugDirPath, bug.Name()), foundBugs, nil, res)
		}

		for _, bug := range foundBugs {
			res["detected"][bug.bugType]++

			if bug.replayWritten {
				res["replayWritten"][bug.bugType]++
			}

			if bug.replaySuc {
				res["replaySuccessful"][bug.bugType]++
			}
		}

		td := data[testName]
		td.results = res
		data[testName] = td
	}

	// write the data to the file
	_, err = os.Stat(statsFuzzingPath)
	fileExisted := (err == nil || !os.IsNotExist(err))

	fuzzingFile, err := os.OpenFile(statsFuzzingPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fuzzingFile.Close()

	if !fileExisted {
		fuzzingFile.WriteString(headers)
	}

	for _, d := range data {
		fuzzingFile.WriteString(d.toString() + "\n")
	}

	return nil
}
