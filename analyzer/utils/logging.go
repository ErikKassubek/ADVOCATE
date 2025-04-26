// Copyright (c) 2025 Erik Kassubek
//
// File: logging.go
// Brief: Logging function
//
// Author: Erik Kassubek
// Created: 2025-02-18
//
// License: BSD-3-Clause

package utils

import (
	"fmt"
	"log"
)

// Color codes for the logging output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Purple = "\033[35m"
)

var numberErr = 0
var numberTimeout = 0
var numberResults = 0
var numberResultsConf = 0

var noInfoFlag bool

var preventPanicFlag bool

// LogInit initializes the logging
//
// Parameter:
//   - noInfo bool: if set, no info is shown during execution
//     errors, results and important are still shown
//   - preventPanic bool: is set to true, panics will only stop the current
//     analysis or test and not the whole analyzer
func LogInit(noInfo, preventPanic bool) {
	noInfoFlag = noInfo
	preventPanicFlag = preventPanic
}

// LogInfo logs an information to the terminal
// Printed in base color
//
// Parameter:
//   - v ...any: the content of the log
func LogInfo(v ...any) {
	if noInfoFlag {
		return
	}

	log.Println(v...)
}

// LogInfof logs an information to the terminal
// Printed in base color
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogInfof(format string, v ...any) {
	if noInfoFlag {
		return
	}

	log.Printf(format, v...)
}

// LogImportant logs an important information to the terminal
// Printed in yellow
//
// Parameter:
//   - v ...any: the content of the log
func LogImportant(v ...any) {
	log.Print(Yellow, fmt.Sprint(v...), Reset, "\n")
}

// LogImportantf logs an important information to the terminal
// Printed in yellow
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogImportantf(format string, v ...any) {
	log.Printf(Yellow+format+Reset, v...)
}

// LogResult logs a result to the terminal
// Printed in green
//
// Parameter:
//   - confirmed bool: true of bug is actual or replay was suc, false otherwise
//   - v ...any: the content of the log
func LogResult(confirmed bool, v ...any) {
	log.Print(Green, fmt.Sprint(v...), Reset, "\n")
	numberResults++
	if confirmed {
		numberResultsConf++
	}
}

// LogResultf logs a result to the terminal
// Printed in green
//
// Parameter:
//   - confirmed bool: true of bug is actual or replay was suc, false otherwise
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogResultf(confirmed bool, format string, v ...any) {
	log.Printf(Green+format+Reset, v...)
	numberResults++
	if confirmed {
		numberResultsConf++
	}
}

// LogTimeout logs a timeout to the terminal
// Printed in purple
// Counts number of timeouts
//
// Parameter:
//   - v ...any: the content of the log
func LogTimeout(v ...any) {
	log.Print(Purple, fmt.Sprint(v...), Reset, "\n")
	numberTimeout++
}

// LogTimeoutf logs a timeout to the terminal
// Printed in purple
// Counts number of timeouts
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogTimeoutf(format string, v ...any) {
	log.Printf(Purple+format+Reset, v...)
	numberTimeout++
}

// LogError logs an error to the terminal
// Printed in red
// Counts number of error
//
// Parameter:
//   - v ...any: the content of the log
func LogError(v ...any) {
	log.Print(Red, fmt.Sprint(v...), Reset, "\n")
	numberErr++
}

// LogErrorf logs an error to the terminal
// Printed in red
// Counts number of error
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogErrorf(format string, v ...any) {
	log.Printf(Red+format+Reset, v...)
	numberErr++
}

// GetLoggingNumbers returns the number of results, errors and timeouts
//
// Returns:
//   - int: number of results
//   - int: number of confirmed results
//   - int: number of errors
//   - int: number of timeouts
func GetLoggingNumbers() (int, int, int, int) {
	return numberResults, numberResultsConf, numberErr, numberTimeout
}

// IsPanicPrevent returns if panic should be suppressed
//
// Returns:
//   - bool: true if panics should recover
func IsPanicPrevent() bool {
	return preventPanicFlag
}
