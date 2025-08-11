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
	"advocate/utils/log"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TODO: for each test get the number of unique bugs

// CreateStatsFuzzing creates statistics about fuzzing runs
//
// Parameter:
//   - pathFolder string: path to where the stats files should be created
//   - progName string: name of the analyzed program
//
// Returns:
//   - error
func CreateStatsFuzzing(pathFolder, progName string) error {
	// collect the info from the analyzer
	resultPath := filepath.Join(pathFolder, "advocateResult")
	statsAnalyzerPath := filepath.Join(resultPath, "statsAnalysis_"+progName+".csv")
	statsFuzzingPath := filepath.Join(resultPath, "statsFuzzing_"+progName+".csv")
	statsFuzzPath := filepath.Join(resultPath, "statsFuzz_"+progName+".csv")

	log.Info("Create fuzzing statistics")

	headers := "TestName,NrRuns,NrMuts,NrMutsInvalid,NrMutsDouble"

	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "L11", "R01", "R02"} {
			headers += fmt.Sprintf(",Nr%s%s", strings.ToUpper(string(mode[0]))+mode[1:], code)
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
			continue
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

	// read the fuzz data
	fuzzFile, err := os.Open(statsFuzzPath)
	if err != nil {
		return err
	}
	defer analysisFile.Close()

	scannerFuzz := bufio.NewScanner(fuzzFile)

	// skip the first line
	scannerFuzz.Scan()

	for scannerFuzz.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")

		if len(fields) < 4 {
			continue
		}

		testName := fields[0]

		td := data[testName]
		if td.fuzzData == nil {
			td.fuzzData = make(map[string]int)
		}

		nrMut, _ := strconv.Atoi(fields[1])
		nrMutInvalid, _ := strconv.Atoi(fields[2])
		nrMutDouble, _ := strconv.Atoi(fields[3])
		nrActiveReleased, _ := strconv.Atoi(fields[4])

		td.fuzzData["nrMut"] += nrMut
		td.fuzzData["nrMutInvalid"] += nrMutInvalid
		td.fuzzData["nrMutDouble"] += nrMutDouble
		td.fuzzData["ActiveReleased"] += nrActiveReleased

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
