// Copyright (c) 2024 Erik Kassubek
//
// File: results.go
// Brief: Function for debug results and for results found bugs
//
// Author: Erik Kassubek
// Created: 2023-08-30
//
// License: BSD-3-Clause

package results

import (
	"advocate/trace"
	"advocate/utils/control"
	"advocate/utils/flags"
	"advocate/utils/helper"
	"advocate/utils/types"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// resultLevel is an enum for the severity of a result
type resultLevel int

// values for the resultLevel enum
const (
	NONE resultLevel = iota
	CRITICAL
	WARNING
	INFORMATION
)

var resultTypeMap = map[helper.ResultType]string{
	helper.ASendOnClosed:           "Actual Send on Closed Channel",
	helper.ARecvOnClosed:           "Actual Receive on Closed Channel",
	helper.ACloseOnClosed:          "Actual Close on Closed Channel",
	helper.AConcurrentRecv:         "Concurrent Receive",
	helper.ASelCaseWithoutPartner:  "Select Case without Partner",
	helper.ACloseOnNilChannel:      "Actual close on nil channel",
	helper.ANegWG:                  "Actual negative Wait Group",
	helper.AUnlockOfNotLockedMutex: "Actual unlock of not locked mutex",
	helper.AMixedDeadlock:          "Actual mixed deadlock",
	// MIXED DEADLOCK [REMOVE]

	helper.PSendOnClosed:     "Possible send on closed channel",
	helper.PRecvOnClosed:     "Possible receive on closed channel",
	helper.PNegWG:            "Possible negative waitgroup counter",
	helper.PUnlockBeforeLock: "Possible unlock of a not locked mutex",
	helper.PCyclicDeadlock:   "Possible cyclic deadlock",
	helper.PMixedDeadlock:    "Possible mixed deadlock",
	// MIXED DEADLOCK [REMOVE]

	helper.LUnknown:           "Leak on routine or unknown element",
	helper.LUnbufferedWith:    "Leak on unbuffered channel with possible partner",
	helper.LUnbufferedWithout: "Leak on unbuffered channel without possible partner",
	helper.LBufferedWith:      "Leak on buffered channel with possible partner",
	helper.LBufferedWithout:   "Leak on unbuffered channel without possible partner",
	helper.LNilChan:           "Leak on nil channel",
	helper.LSelectWith:        "Leak on select with possible partner",
	helper.LSelectWithout:     "Leak on select without partner or nil case",
	helper.LMutex:             "Leak on mutex",
	helper.LWaitGroup:         "Leak on wait group",
	helper.LCond:              "Leak on conditional variable",
	helper.LContext:           "Leak on a channel or select on context",

	helper.RUnknownPanic: "Unknown Panic",
	helper.RTimeout:      "Timeout",

	// SNotExecutedWithPartner: "Not executed select with potential partner",
}

var (
	outputReadableFile       string
	outputMachineFile        string
	foundBug                 = false
	resultsWarningReadable   []string
	resultsCriticalReadable  []string
	resultsWarningMachine    []string
	resultCriticalMachine    []string
	resultInformationMachine []string
	resultWithoutTime        []string
)

// ResultElem declares an interface for a result elem
type ResultElem interface {
	isInvalid() bool
	stringMachine() string
	stringReadable() string
	stringMachineShort() string
	getFile() string
}

// TraceElementResult is a type to represent an element that is
// part of a found bug
// TODO: replace by pointer to actual element
type TraceElementResult struct {
	RoutineID int
	ObjID     int
	TPre      int
	ObjType   trace.ObjectType
	File      string
	Line      int
}

// getFile returns the file path of the element
//
// Returns:
//   - string: the file path
func (this TraceElementResult) getFile() string {
	return this.File
}

// stringMachineShort returns a short machine readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (this TraceElementResult) stringMachineShort() string {
	return fmt.Sprintf("T:%d:%s:%s:%d", this.ObjID, this.ObjType, this.File, this.Line)
}

// stringMachine returns a machine readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (this TraceElementResult) stringMachine() string {
	return fmt.Sprintf("T:%d:%d:%d:%s:%s:%d", this.RoutineID, this.ObjID, this.TPre, this.ObjType, this.File, this.Line)
}

// stringReadable returns a human readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (this TraceElementResult) stringReadable() string {
	return fmt.Sprintf("%s:%d@%d", this.File, this.Line, this.TPre)
}

// isInvalid checks if the result element is not corrupted/empty
//
// Returns:
//   - bool: true if valid, false otherwise
func (this TraceElementResult) isInvalid() bool {
	return this.ObjType == "" || this.Line == -1
}

// Result logs a found bug
//
// Parameter:
//   - level resultLevel: level of the message (critical, warning, ...)
//   - resType ResultType: type of bug that was found
//   - argType1 string: description of the type of elements in arg1
//   - arg1 []ResultElem]: elements directly involved in the bug (e.g. in send on closed the send)
//   - argType2 string: description of the type of elements in arg2
//   - arg2 []ResultElem]: elements indirectly involved in the bug (e.g. in send on closed the close)
func Result(level resultLevel, resType helper.ResultType, argType1 string, arg1 []ResultElem, argType2 string, arg2 []ResultElem) {
	if filterInvalidResults(resType, arg1) {
		return
	}

	foundBug = true

	resultReadable := resultTypeMap[resType] + ":\n\t" + argType1 + ": "
	resultMachine := string(resType) + ","
	resultMachineShort := string(resType)

	for i, arg := range arg1 {
		if arg.isInvalid() {
			continue
		}
		if i != 0 {
			resultReadable += ";"
			resultMachine += ";"
		}
		resultReadable += arg.stringReadable()
		resultMachine += arg.stringMachine()
		resultMachineShort += arg.stringMachineShort()
	}

	resultReadable += "\n"
	if len(arg2) > 0 {
		resultReadable += "\t" + argType2 + ": "
		resultMachine += ","
		for i, arg := range arg2 {
			if arg.isInvalid() {
				continue
			}
			if i != 0 {
				resultReadable += ";"
				resultMachine += ";"
			}
			resultReadable += arg.stringReadable()
			resultMachine += arg.stringMachine()
			resultMachineShort += arg.stringMachineShort()
		}
	}

	resultReadable += "\n"
	resultMachine += "\n"

	switch level {
	case WARNING:
		if !types.Contains(resultWithoutTime, resultMachineShort) {
			resultsWarningReadable = append(resultsWarningReadable, resultReadable)
			resultsWarningMachine = append(resultsWarningMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	case CRITICAL:
		if !types.Contains(resultWithoutTime, resultMachineShort) {
			resultsCriticalReadable = append(resultsCriticalReadable, resultReadable)
			resultCriticalMachine = append(resultCriticalMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	case INFORMATION:
		if !types.Contains(resultWithoutTime, resultMachineShort) {
			resultInformationMachine = append(resultInformationMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	}
}

// Some results are invalid or intentionally not shown. This function returns,
// if the given parameters constitute such a result
//
// Parameter:
//   - resType ResultType: type of bug that was found
//   - arg1 []ResultElem]: elements directly involved in the bug (e.g. in send on closed the send)
//
// Returns:
//   - bool: true if the result is invalid and should be ignored, false otherwise
func filterInvalidResults(resType helper.ResultType, arg1 []ResultElem) bool {
	if resType != helper.RUnknownPanic && resType != helper.RTimeout && len(arg1) == 0 {
		return true
	}

	if control.CheckCanceled() {
		return true
	}

	if resType == helper.ADeadlock && len(arg1) == 1 && strings.HasSuffix(arg1[0].getFile(), "/src/testing/testing.go") {
		return true
	}

	return false
}

// InitResults sets the output file paths and clears al previous results
//
// Parameter:
//   - outReadable: path to the output file, no output file if empty
//   - outMachine: path to the output file for the reordered trace, no output file if empty
func InitResults(outReadable string, outMachine string) {
	Reset()
	outputReadableFile = outReadable
	outputMachineFile = outMachine
}

// CreateResultFiles writes out the results to the machine and human
// readable result files nad print them to the terminal
//
// Parameter:
//   - noPrint bool: if true, do not print the errors to the terminal
//
// Returns:
//   - int: number of bugs found
//   - error
func CreateResultFiles(noPrint bool) (int, error) {
	counter := 1
	resMachine := ""
	resReadable := "```\n==================== Summary ====================\n\n"

	if !noPrint {
		fmt.Print("==================== Summary ====================\n\n")
	}

	found := false

	if len(resultsCriticalReadable) > 0 {
		found = true
		resReadable += "-------------------- Critical -------------------\n\n"

		if !noPrint {
			fmt.Print("-------------------- Critical -------------------\n\n")
		}

		for _, result := range resultsCriticalReadable {
			resReadable += strconv.Itoa(counter) + " " + result + "\n"

			if !noPrint {
				fmt.Println(strconv.Itoa(counter) + " " + result)
			}

			counter++
		}

		for _, result := range resultCriticalMachine {
			resMachine += result
		}
	}

	if !flags.NoWarning {
		if len(resultsWarningReadable) > 0 {
			found = true
			resReadable += "\n-------------------- Warning --------------------\n\n"
			if !noPrint {
				fmt.Print("\n-------------------- Warning --------------------\n\n")
			}

			for _, result := range resultsWarningReadable {
				resReadable += strconv.Itoa(counter) + " " + result + "\n"

				if !noPrint {
					fmt.Println(strconv.Itoa(counter) + " " + result)
				}

				counter++
			}

			for _, result := range resultsWarningMachine {
				resMachine += result
			}
		}

		for _, result := range resultInformationMachine {
			resMachine += result
		}
	}

	if !found {
		resReadable += "No bugs found" + "\n"

		if !noPrint {
			fmt.Println("No bugs found")
		}
	}

	resReadable += "```"

	// write output readable
	if _, err := os.Stat(outputReadableFile); err == nil {
		if err := os.Remove(outputReadableFile); err != nil {
			return getNumberRes(), err
		}
	}

	file, err := os.OpenFile(outputReadableFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return getNumberRes(), err
	}
	defer file.Close()

	if _, err := file.WriteString(resReadable); err != nil {
		return getNumberRes(), err
	}

	// write output machine
	if _, err := os.Stat(outputMachineFile); err == nil {
		if err := os.Remove(outputMachineFile); err != nil {
			return getNumberRes(), err
		}
	}

	file, err = os.OpenFile(outputMachineFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return getNumberRes(), err
	}
	defer file.Close()

	if _, err := file.WriteString(resMachine); err != nil {
		return getNumberRes(), err
	}

	return getNumberRes(), nil
}

// getNumberRes returns the number of found bugs
func getNumberRes() int {
	if flags.NoWarning {
		return len(resultCriticalMachine)
	}
	return len(resultCriticalMachine) + len(resultsWarningMachine) + len(resultInformationMachine)
}

// Reset the global values storing the found results
func Reset() {
	resultsWarningReadable = make([]string, 0)
	resultsCriticalReadable = make([]string, 0)
	resultsWarningMachine = make([]string, 0)
	resultCriticalMachine = make([]string, 0)
	resultInformationMachine = make([]string, 0)

	resultWithoutTime = make([]string, 0)

	outputMachineFile = ""
	outputReadableFile = ""

	foundBug = false
}

// GetBugWasFound returns if since the last reset, a bug was found
//
// Returns:
//   - foundBug
func GetBugWasFound() bool {
	return foundBug
}
