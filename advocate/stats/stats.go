// Copyright (c) 2024 Erik Kassubek
//
// File: stats.go
// Brief: Create statistics about programs and traces
//
// Author: Erik Kassubek
// Created: 2023-07-13
//
// License: BSD-3-Clause

package stats

import (
	"advocate/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// testData stores information about a test
//
// Parameter:
//   - name string: name of the test
//   - numberRuns int: for fuzzing, how often the test was run
//   - results map[string]map[string]int: information about the found bugs in this test
type testData struct {
	name       string
	numberRuns int
	results    map[string]map[string]int
}

// toString returns the string representation of the statistics of a test
//
// Returns:
//   - string: the string representation
func (td *testData) toString() string {
	res := fmt.Sprintf("%s,%d", td.name, td.numberRuns)

	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "R01", "R02"} {
			res += fmt.Sprintf(",%d", td.results[mode][code])
		}
	}

	return res
}

// CreateStats adds the information of an analyzed test to the stats info
//
// Parameter:
//   - pathFolder string: path to where the stats file should be created
//   - progName string: name of the analyzed program
//   - testName string: name of the analyzed test
//   - traceID int: id of the trace
//   - fuzzing int: number of fuzzing run
//
// Returns:
//   - error
func CreateStats(pathFolder, progName string, testName string, traceID, fuzzing int) error {
	// statsProg, err := statsProgram(pathToProgram)
	// if err != nil {
	// 	return err
	// }

	utils.LogInfo("Create statistics")

	statsTrace, err := statsTraces(pathFolder, traceID)
	if err != nil {
		return err
	}

	statsMisc, err := statsMisc(pathFolder, testName)
	if err != nil {
		return err
	}

	statsAnalyzerTotal, statsAnalyzerUnique, err := statsAnalyzer(pathFolder, fuzzing)
	if err != nil {
		return err
	}

	err = writeStatsToFile(filepath.Dir(pathFolder), progName, testName, statsTrace, statsMisc, statsAnalyzerTotal, statsAnalyzerUnique)
	if err != nil {
		return err
	}

	return nil

}

// Write the collected statistics to files
//
// Parameter:
//   - path string: path to where the stats file should be created
//   - progName string: name of the program
//   - testName string: name of the test
//   - statsProg map[string]int: statistics about the program
//   - statsTraces map[string]int: statistics about the trace
//   - statsMisc map[string]int: miscellaneous statistics
//   - statsAnalyzerTotal map[string]map[string]int: statistics about the total analysis and replay
//   - statsAnalyzerUnique map[string]map[string]int: statistics about the unique analysis and replay
//
// Returns:
//   - error
func writeStatsToFile(path string, progName string, testName string, statsTraces map[string]int, statsMisc map[string]int,
	statsAnalyzerTotal, statsAnalyzerUnique map[string]map[string]int) error {

	fileMiscPath := filepath.Join(path, "statsMisc_"+progName+".csv")
	fileTracingPath := filepath.Join(path, "statsTrace_"+progName+".csv")
	fileAnalysisPath := filepath.Join(path, "statsAnalysis_"+progName+".csv")
	fileAllPath := filepath.Join(path, "statsAll_"+progName+".csv")

	headerTracing := "TestName,NoEvents,NoGoroutines,NoAtomicEvents," +
		"NoChannelEvents,NoSelectEvents,NoMutexEvents,NoWaitgroupEvents," +
		"NoCondVariablesEvents,NoOnceOperations"
	dataTracing := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d", testName,
		statsTraces["numberElements"], statsTraces["numberRoutines"],
		statsTraces["numberAtomicOperations"], statsTraces["numberChannelOperations"],
		statsTraces["numberSelects"], statsTraces["numberMutexOperations"],
		statsTraces["numberWaitGroupOperations"], statsTraces["numberCondVarOperations"],
		statsTraces["numberOnceOperations"])

	writeStatsFile(fileTracingPath, headerTracing, dataTracing)

	actualCodes := []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08"}
	numberOfActualBugsTotal := 0
	numberOfActualBugsUnique := 0
	for _, code := range actualCodes {
		numberOfActualBugsTotal += statsAnalyzerTotal["detected"][code]
		numberOfActualBugsUnique += statsAnalyzerUnique["detected"][code]
	}

	leakCodes := []string{"L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10"}

	numberOfLeaksTotal := 0
	numberOfLeaksUnique := 0
	for _, code := range leakCodes {
		numberOfLeaksTotal += statsAnalyzerTotal["detected"][code]
		numberOfLeaksUnique += statsAnalyzerUnique["detected"][code]
	}

	numberOfLeaksWithRewriteTotal := 0
	numberOfLeaksWithRewriteUnique := 0
	for _, code := range leakCodes {
		numberOfLeaksWithRewriteTotal += statsAnalyzerTotal["replayWritten"][code]
		numberOfLeaksWithRewriteUnique += statsAnalyzerUnique["replayWritten"][code]
	}

	numberOfLeaksResolvedViaReplayTotal := 0
	numberOfLeaksResolvedViaReplayUnique := 0
	for _, code := range leakCodes {
		numberOfLeaksResolvedViaReplayTotal += statsAnalyzerTotal["replaySuccessful"][code]
		numberOfLeaksResolvedViaReplayUnique += statsAnalyzerUnique["replaySuccessful"][code]
	}

	posPanicCodes := []string{"P01", "P03", "P04", "P05"}

	numberOfPanicsTotal := 0
	numberOfPanicsUnique := 0
	for _, code := range posPanicCodes {
		numberOfPanicsTotal += statsAnalyzerTotal["detected"][code]
		numberOfPanicsUnique += statsAnalyzerUnique["detected"][code]
	}

	numberOfPanicsVerifiedViaReplayTotal := 0
	numberOfPanicsVerifiedViaReplayUnique := 0
	for _, code := range posPanicCodes {
		numberOfPanicsVerifiedViaReplayTotal += statsAnalyzerTotal["replaySuccessful"][code]
		numberOfPanicsVerifiedViaReplayUnique += statsAnalyzerUnique["replaySuccessful"][code]
	}

	numberUnexpectedPanicsInReplayTotal := 0
	numberUnexpectedPanicsInReplayUnique := 0
	for _, code := range posPanicCodes {
		numberUnexpectedPanicsInReplayTotal += statsAnalyzerTotal["unexpectedPanic"][code]
		numberUnexpectedPanicsInReplayUnique += statsAnalyzerUnique["unexpectedPanic"][code]
	}

	probInRecCodes := []string{"R01", "R02"}
	numberProbInRecord := 0
	for _, code := range probInRecCodes {
		numberProbInRecord += statsAnalyzerTotal["detected"][code]
	}

	headerAnalysis := "TestName,NumberActualBugTotal,NoLeaksTotal,NoLeaksWithRewriteTotal,NoLeaksResolvedViaReplayTotal,NoPanicsTotal,NoPanicsVerifiedViaReplayTotal,NoUnexpectedPanicsInReplayTotal,NoProbInRecordingTotal,NumberActualBugUnique,NoLeaksUnique,NoLeaksWithRewriteUnique,NoLeaksResolvedViaReplayUnique,NoPanicsUnique,NoPanicsVerifiedViaReplayUnique,NoUnexpectedPanicsInReplayUnique"
	dataAnalysis := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d", testName, numberOfActualBugsTotal, numberOfLeaksTotal,
		numberOfLeaksWithRewriteTotal, numberOfLeaksResolvedViaReplayTotal, numberOfPanicsTotal, numberOfPanicsVerifiedViaReplayTotal, numberUnexpectedPanicsInReplayTotal, numberProbInRecord, numberOfActualBugsUnique, numberOfLeaksUnique,
		numberOfLeaksWithRewriteUnique, numberOfLeaksResolvedViaReplayUnique, numberOfPanicsUnique, numberOfPanicsVerifiedViaReplayUnique, numberUnexpectedPanicsInReplayUnique)

	writeStatsFile(fileAnalysisPath, headerAnalysis, dataAnalysis)

	headerDetails := "TestName," +
		"NoEvents,NoGoroutines,NoNotEmptyGoroutines,NoSpawnEvents,NoRoutineEndEvents," +
		"NoAtomics,NoAtomicEvents,NoChannels,NoBufferedChannels,NoUnbufferedChannels," +
		"NoChannelEvents,NoBufferedChannelEvents,NoUnbufferedChannelEvents,NoSelectEvents," +
		"NoSelectCases,NoSelectNonDefaultEvents,NoSelectDefaultEvents,NoMutex,NoMutexEvents," +
		"NoWaitgroup,NoWaitgroupEvent,NoCondVariables,NoCondVariablesEvents,NoOnce,NoOnceOperations,"
	dataDetails := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,",
		testName, statsTraces["numberElements"],
		statsTraces["numberRoutines"], statsTraces["numberNonEmptyRoutines"],
		statsTraces["numberOfSpawns"], statsTraces["numberRoutineEnds"],
		statsTraces["numberAtomics"], statsTraces["numberAtomicOperations"],
		statsTraces["numberChannels"], statsTraces["numberBufferedChannels"],
		statsTraces["numberUnbufferedChannels"], statsTraces["numberChannelOperations"],
		statsTraces["numberBufferedOps"], statsTraces["numberUnbufferedOps"],
		statsTraces["numberSelects"], statsTraces["numberSelectCases"],
		statsTraces["numberSelectChanOps"], statsTraces["numberSelectDefaultOps"],
		statsTraces["numberMutexes"], statsTraces["numberMutexOperations"],
		statsTraces["numberWaitGroups"], statsTraces["numberWaitGroupOperations"],
		statsTraces["numberCondVars"], statsTraces["numberCondVarOperations"],
		statsTraces["numberOnce"], statsTraces["numberOnceOperations"])

	headers := make([]string, 0)
	data := make([]string, 0)
	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "unexpectedPanic"} {
		for _, count := range []string{"Total", "Unique"} {
			for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "P01", "P02", "P03", "P04", "P05", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10", "R01", "R02"} {
				headers = append(headers, "No"+count+strings.ToUpper(string(mode[0]))+mode[1:]+code)
				if count == "Total" {
					data = append(data, strconv.Itoa(statsAnalyzerTotal[mode][code]))
				} else {
					data = append(data, strconv.Itoa(statsAnalyzerUnique[mode][code]))
				}
			}
		}
	}
	headerDetails += strings.Join(headers, ",")
	dataDetails += strings.Join(data, ",")

	writeStatsFile(fileAllPath, headerDetails, dataDetails)

	miscData := make([]string, len(miscStats))
	for i, header := range miscStats {
		if header == TestName {
			miscData[i] = testName
			continue
		}
		if val, exists := statsMisc[header]; exists {
			miscData[i] = strconv.Itoa(val)
		} else {
			miscData[i] = "0"
		}
	}

	writeStatsFile(fileMiscPath, strings.Join(miscStats, ","), strings.Join(miscData, ","))

	return nil
}

// writeStatsFile writes the collected stats to a csv file
//
// Parameter:
//   - path string: path to where the stat file should be created
//   - header string: first line of the stat file containing column names
//   - data string: the stats data to write into the files
func writeStatsFile(path, header, data string) {
	newFile := false
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		newFile = true
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
		return
	}
	defer file.Close()

	if newFile {
		file.WriteString(header)
		file.WriteString("\n")
	}
	file.WriteString(data)
	file.WriteString("\n")
}
