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
	log.Println("Create fuzzing statistics at " + statsFuzzingPath)

	headers := "TestName,NumberRuns"

	data := ""

	lastTestName := ""
	counter := 0

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
				data += fmt.Sprintf("\n%s,%d", lastTestName, counter)
			}
			lastTestName = testName
			counter = 1
		}
	}

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
	fuzzingFile.WriteString(data)

	return nil
}
