// Copyright (c) 2025 Erik Kassubek
//
// File: io.go
// Brief: Function to write the time file
//
// Author: Erik Kassubek
// Created: 2025-02-25
//
// License: BSD-3-Clause

package timer

import (
	"analyzer/utils"
	"fmt"
	"os"
	"path/filepath"
)

var (
	resultFolder = ""
)

// UpdateTimeFileDetail writes the time information to a file
//
// Parameter:
//   - progName string: name of the program
//   - testName string: name of the test
//   - numberReplay int: number of replay
func UpdateTimeFileDetail(progName string, testName string, numberReplay int) {
	if !measureTime {
		return
	}

	timeFilePath := filepath.Join(resultFolder, "times_detail_"+progName+".csv")

	newFile := false
	_, err := os.Stat(timeFilePath)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(timeFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.LogError("Error opening or creating file:", err)
		return
	}
	defer file.Close()

	if newFile {
		csvTitels := "TestName,Run,Recording,Io,Analysis,AnaHb,AnaExitCode,AnaLeak,AnaClose,AnaConcurrent,AnaResource,AnaSelWithoutPartner,AnaUnlock,AnaWait,AnaMixed,FuzzingAna,FuzzingMut,Rewrite,Replay,NumberReplay\n"
		if _, err := file.WriteString(csvTitels); err != nil {
			utils.LogError("Could not write time: ", err)
		}
	}

	timeInfo := fmt.Sprintf(
		"%s,%s,%d\n", testName, ToString(), numberReplay)

	// Write to the file
	if _, err := file.WriteString(timeInfo); err != nil {
		utils.LogError("Could not write time: ", err)
	}
}

// UpdateTimeFileOverview write the current timer values to a file
// if time measurement is enabled,
//
// Parameter:
//   - progName string: name of the prog
//   - testName string: name of the test
func UpdateTimeFileOverview(progName string, testName string) {
	if !measureTime {
		return
	}

	timeFilePath := filepath.Join(resultFolder, "times_total_"+progName+".csv")

	newFile := false
	_, err := os.Stat(timeFilePath)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(timeFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.LogError("Error opening or creating file:", err)
		return
	}
	defer file.Close()

	if newFile {
		csvTitels := "TestName,Time\n"
		if _, err := file.WriteString(csvTitels); err != nil {
			utils.LogError("Could not write time: ", err)
		}
	}

	timeInfo := ""
	if testName == "*Total*" {
		timeInfo = fmt.Sprintf(
			"%s,%.5f\n", "Total", GetTime(Total).Seconds())
	} else {
		timeInfo = fmt.Sprintf(
			"%s,%.5f\n", testName, GetTime(TotalTest).Seconds())
	}

	// Write to the file
	if _, err := file.WriteString(timeInfo); err != nil {
		utils.LogError("Could not write time: ", err)
	}
}
