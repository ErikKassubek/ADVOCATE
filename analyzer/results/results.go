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
	"analyzer/utils"
	"fmt"
	"os"
	"strconv"
)

// resultLevel is an enum for the severity of a result
type resultLevel int

const (
	NONE resultLevel = iota
	CRITICAL
	WARNING
	INFORMATION
)

// ResultType is an ID for a type of result
type ResultType string

const (
	Empty ResultType = ""

	// actual
	ASendOnClosed           ResultType = "A01"
	ARecvOnClosed           ResultType = "A02"
	ACloseOnClosed          ResultType = "A03"
	ACloseOnNilChannel      ResultType = "A04"
	ANegWG                  ResultType = "A05"
	AUnlockOfNotLockedMutex ResultType = "A06"
	AConcurrentRecv         ResultType = "A07"
	ASelCaseWithoutPartner  ResultType = "A08"

	// possible
	PSendOnClosed     ResultType = "P01"
	PRecvOnClosed     ResultType = "P02"
	PNegWG            ResultType = "P03"
	PUnlockBeforeLock ResultType = "P04"
	PCyclicDeadlock   ResultType = "P05"

	// leaks
	LWithoutBlock      = "L00"
	LUnbufferedWith    = "L01"
	LUnbufferedWithout = "L02"
	LBufferedWith      = "L03"
	LBufferedWithout   = "L04"
	LNilChan           = "L05"
	LSelectWith        = "L06"
	LSelectWithout     = "L07"
	LMutex             = "L08"
	LWaitGroup         = "L09"
	LCond              = "L10"

	// recording
	RUnknownPanic ResultType = "R01"
	RTimeout      ResultType = "R02"

	// not executed select
	// SNotExecutedWithPartner = "S00"
)

var resultTypeMap = map[ResultType]string{
	ASendOnClosed:           "Actual Send on Closed Channel:",
	ARecvOnClosed:           "Actual Receive on Closed Channel:",
	ACloseOnClosed:          "Actual Close on Closed Channel:",
	AConcurrentRecv:         "Concurrent Receive:",
	ASelCaseWithoutPartner:  "Select Case without Partner",
	ACloseOnNilChannel:      "Actual close on nil channel",
	ANegWG:                  "Actual negative Wait Group",
	AUnlockOfNotLockedMutex: "Actual unlock of not locked mutex",

	PSendOnClosed:     "Possible send on closed channel:",
	PRecvOnClosed:     "Possible receive on closed channel:",
	PNegWG:            "Possible negative waitgroup counter:",
	PUnlockBeforeLock: "Possible unlock of a not locked mutex:",
	PCyclicDeadlock:   "Possible cyclic deadlock:",

	LWithoutBlock:      "Leak on routine without any blocking operation",
	LUnbufferedWith:    "Leak on unbuffered channel with possible partner:",
	LUnbufferedWithout: "Leak on unbuffered channel without possible partner:",
	LBufferedWith:      "Leak on buffered channel with possible partner:",
	LBufferedWithout:   "Leak on unbuffered channel without possible partner:",
	LNilChan:           "Leak on nil channel:",
	LSelectWith:        "Leak on select with possible partner:",
	LSelectWithout:     "Leak on select without partner or nil case",
	LMutex:             "Leak on mutex:",
	LWaitGroup:         "Leak on wait group:",
	LCond:              "Leak on conditional variable:",

	RUnknownPanic: "Unknown Panic",
	RTimeout:      "Timeout",

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
}

// TraceElementResult is a type to represent an element that is
// part of a found bug
// TODO: replace by pointer to actual element
type TraceElementResult struct {
	RoutineID int
	ObjID     int
	TPre      int
	ObjType   string
	File      string
	Line      int
}

// stringMachineShort returns a short machine readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (t TraceElementResult) stringMachineShort() string {
	return fmt.Sprintf("T:%d:%s:%s:%d", t.ObjID, t.ObjType, t.File, t.Line)
}

// stringMachine returns a machine readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (t TraceElementResult) stringMachine() string {
	return fmt.Sprintf("T:%d:%d:%d:%s:%s:%d", t.RoutineID, t.ObjID, t.TPre, t.ObjType, t.File, t.Line)
}

// stringReadable returns a human readable string representation
// of a result element
//
// Returns:
//   - string: the string representation
func (t TraceElementResult) stringReadable() string {
	return fmt.Sprintf("%s:%d@%d", t.File, t.Line, t.TPre)
}

// isInvalid checks if the result element is not corrupted/empty
//
// Returns:
//   - bool: true if valid, false otherwise
func (t TraceElementResult) isInvalid() bool {
	return t.ObjType == "" || t.Line == -1
}

// SelectCaseResult is a type to represent an select case that is
// part of a found bug
type SelectCaseResult struct {
	SelID   int
	ObjID   int
	ObjType string
	Routine int
	Index   int
}

// stringMachineShort returns a short machine readable string representation
// of a result select case
//
// Returns:
//   - string: the string representation
func (s SelectCaseResult) stringMachineShort() string {
	return fmt.Sprintf("S:%d:%s:%d", s.ObjID, s.ObjType, s.Index)
}

// stringMachineShort returns a machine readable string representation
// of a result select case
//
// Returns:
//   - string: the string representation
func (s SelectCaseResult) stringMachine() string {
	return fmt.Sprintf("S:%d:%s:%d", s.ObjID, s.ObjType, s.Index)
}

// stringReadable returns a human readable string representation
// of a result select case
//
// Returns:
//   - string: the string representation
func (s SelectCaseResult) stringReadable() string {
	return fmt.Sprintf("%d:%s", s.ObjID, s.ObjType)
}

// isInvalid checks if the result select case is not corrupted/empty
//
// Returns:
//   - bool: true if valid, false otherwise
func (s SelectCaseResult) isInvalid() bool {
	return s.ObjType == ""
}

// Result logs a found bug
//
// Parameter:
//   - level (resultLevel): level of the message (critical, warning, ...)
//   - resType (ResultType): type of bug that was found
//   - argType1 (string): description of the type of elements in arg1
//   - arg1 ([]ResultElem]): elements directly involved in the bug (e.g. in send on closed the send)
//   - argType2 (string): description of the type of elements in arg2
//   - arg2 ([]ResultElem]): elements indirectly involved in the bug (e.g. in send on closed the close)
func Result(level resultLevel, resType ResultType, argType1 string, arg1 []ResultElem, argType2 string, arg2 []ResultElem) {
	if len(arg1) == 0 {
		return
	}

	foundBug = true

	resultReadable := resultTypeMap[resType] + "\n\t" + argType1 + ": "
	resultMachine := string(resType) + ","
	resultMachineShort := string(resType)

	for i, arg := range arg1 {
		if arg.isInvalid() {
			return
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
				return
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

	if level == WARNING {
		if !utils.Contains(resultWithoutTime, resultMachineShort) {
			resultsWarningReadable = append(resultsWarningReadable, resultReadable)
			resultsWarningMachine = append(resultsWarningMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	} else if level == CRITICAL {
		if !utils.Contains(resultWithoutTime, resultMachineShort) {
			resultsCriticalReadable = append(resultsCriticalReadable, resultReadable)
			resultCriticalMachine = append(resultCriticalMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	} else if level == INFORMATION {
		if !utils.Contains(resultWithoutTime, resultMachineShort) {
			resultInformationMachine = append(resultInformationMachine, resultMachine)
			resultWithoutTime = append(resultWithoutTime, resultMachineShort)
		}
	}
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
//   - noWarning (bool): if true, only critical errors will be shown
//   - noPrint (bool): if true, do not print the errors to the terminal
//
// Returns:
//   - int: number of bugs found
//   - error
func CreateResultFiles(noWarning bool, noPrint bool) (int, error) {
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

	if !noWarning {
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
			return getNumberRes(noWarning), err
		}
	}

	file, err := os.OpenFile(outputReadableFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return getNumberRes(noWarning), err
	}
	defer file.Close()

	if _, err := file.WriteString(resReadable); err != nil {
		return getNumberRes(noWarning), err
	}

	// write output machine
	if _, err := os.Stat(outputMachineFile); err == nil {
		if err := os.Remove(outputMachineFile); err != nil {
			return getNumberRes(noWarning), err
		}
	}

	file, err = os.OpenFile(outputMachineFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return getNumberRes(noWarning), err
	}
	defer file.Close()

	if _, err := file.WriteString(resMachine); err != nil {
		return getNumberRes(noWarning), err
	}

	return getNumberRes(noWarning), nil
}

// getNumberRes returns the number of found bugs
//
// Parameters:
//   - noWarning (bool): only get the number of
func getNumberRes(noWarning bool) int {
	if noWarning {
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
