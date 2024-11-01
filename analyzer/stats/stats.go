// Copyrigth (c) 2024 Erik Kassubek
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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Create files with the required stats
 * Args:
 *     pathToProgram (string): path to the program
 *     pathToTrace (string): path to the traces
 *     progName (string): name of the analyzed program
 *     testName (string): name of the test
 */
func CreateStats(pathFolder, progName string, testName string) error {
	// statsProg, err := statsProgram(pathToProgram)
	// if err != nil {
	// 	return err
	// }

	statsTrace, err := statsTraces(pathFolder)
	if err != nil {
		return err
	}

	statsAnalyzer, err := statsAnalyzer(pathFolder)
	if err != nil {
		return err
	}

	err = writeStatsToFile(filepath.Dir(pathFolder), progName, testName, statsTrace, statsAnalyzer)
	if err != nil {
		return err
	}

	return nil

}

/*
* Write the collected statistics to files
* Args:
*     path (string): path to where the stats file should be created
*     progName (string): name of the program
*     testName (string): name of the test
*     statsProg (map[string]int): statistics about the program
*     statsTraces (map[string]int): statistics about the trace
*     statsAnalyzer (map[string]map[string]int): statistics about the analysis and replay
* Returns:
*     error
 */
func writeStatsToFile(path string, progName string, testName string, statsTraces map[string]int,
	statsAnalyzer map[string]map[string]int) error {

	fileTracingPath := filepath.Join(path, "statsTrace_"+progName+".csv")
	fileAnalysisPath := filepath.Join(path, "statsAnalysis_"+progName+".csv")
	fileAllPath := filepath.Join(path, "statsAll_"+progName+".csv")

	headerTracing := "TestName,NumberOfEvents,NumberOfGoroutines,NumberOfAtomicEvents," +
		"NumberOfChannelEvents,NumberOfSelectEvents,NumberOfMutexEvents,NumberOfWaitgroupEvents," +
		"NumberOfCondVariablesEvents,NumberOfOnceOperations"
	dataTracing := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d,%d", testName,
		statsTraces["numberElements"], statsTraces["numberRoutines"],
		statsTraces["numberAtomicOperations"], statsTraces["numberChannelOperations"],
		statsTraces["numberSelects"], statsTraces["numberMutexOperations"],
		statsTraces["numberWaitGroupOperations"], statsTraces["numberCondVarOperations"],
		statsTraces["numberOnceOperations"])

	writeStatsFile(fileTracingPath, headerTracing, dataTracing)

	leakCodes := []string{"L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10"}

	numberOfLeaks := 0
	for _, code := range leakCodes {
		numberOfLeaks += statsAnalyzer["detected"][code]
	}

	numberOfLeaksWithRewrite := 0
	for _, code := range leakCodes {
		numberOfLeaksWithRewrite += statsAnalyzer["replayWritten"][code]
	}

	numberOfLeaksResolvedViaReplay := 0
	for _, code := range leakCodes {
		numberOfLeaksResolvedViaReplay += statsAnalyzer["replaySuccessful"][code]
	}

	panicCodes := []string{"P01", "P03", "P04"}

	numberOfPanics := 0
	for _, code := range panicCodes {
		numberOfPanics += statsAnalyzer["detected"][code]
	}

	numberOfPanicsVerifiedViaReplay := 0
	for _, code := range panicCodes {
		numberOfPanicsVerifiedViaReplay += statsAnalyzer["replaySuccessful"][code]
	}

	numberOfLeaksDetectedWithRerecording := 0
	for _, code := range leakCodes {
		numberOfLeaksDetectedWithRerecording += statsAnalyzer["rerecorded"][code]
	}

	numberOfNumberOfPanicsDetectedWithRerecordingPanics := 0
	for _, code := range panicCodes {
		numberOfPanics += statsAnalyzer["rerecorded"][code]
	}

	NumberOfUnexpectedPanicsInReplay := 0
	for _, code := range panicCodes {
		NumberOfUnexpectedPanicsInReplay += statsAnalyzer["unexpectedPanic"][code]
	}

	headerAnalysis := "TestName,NumberOfLeaks,NumberOfLeaksWithRewrite,NumberOfLeaksResolvedViaReplay,NumberOfPanics,NumberOfPanicsVerifiedViaReplay,NumberOfLeaksDetectedWithRerecording,NumberOfPanicsDetectedWithRerecording,NumberOfUnexpectedPanicsInReplay"
	dataAnalysis := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%d,%d", testName, numberOfLeaks,
		numberOfLeaksWithRewrite, numberOfLeaksResolvedViaReplay, numberOfPanics, numberOfPanicsVerifiedViaReplay, numberOfLeaksDetectedWithRerecording, numberOfNumberOfPanicsDetectedWithRerecordingPanics, NumberOfUnexpectedPanicsInReplay)

	writeStatsFile(fileAnalysisPath, headerAnalysis, dataAnalysis)

	headerDetails := "TestName," +
		"NumberOfEvents,NumberOfGoroutines,NumberOfNotEmptyGoroutines,NumberOfSpawnEvents,NumberOfRoutineEndEvents," +
		"NumberOfAtomics,NumberOfAtomicEvents,NumberOfChannels,NumberOfBufferedChannels,NumberOfUnbufferedChannels," +
		"NumberOfChannelEvents,NumberOfBufferedChannelEvents,NumberOfUnbufferedChannelEvents,NumberOfSelectEvents," +
		"NumberOfSelectCases,NumberOfSelectNonDefaultEvents,NumberOfSelectDefaultEvents,NumberOfMutex,NumberOfMutexEvents," +
		"NumberOfWaitgroup,NumberOfWaitgroupEvent,NumberOfCondVariables,NumberOfCondVariablesEvents,NumberOfOnce,NumberOfOnceOperations,"
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
	for _, mode := range []string{"detected", "replayWritten", "replaySuccessful", "rerecorded", "unexpectedPanic"} {
		for _, code := range []string{"A01", "A02", "A03", "A04", "A05", "P01", "P02", "P03", "P04", "L00", "L01", "L02", "L03", "L04", "L05", "L06", "L07", "L08", "L09", "L10"} {
			headers = append(headers, "NumberOf"+strings.ToUpper(string(mode[0]))+mode[1:]+code)
			data = append(data, strconv.Itoa(statsAnalyzer[mode][code]))
		}
	}
	headerDetails += strings.Join(headers, ",")
	dataDetails += strings.Join(data, ",")

	writeStatsFile(fileAllPath, headerDetails, dataDetails)

	return nil
}

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
