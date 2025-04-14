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

// LogInfo logs an information to the terminal
// Printed in base color
//
// Parameter:
//   - v ...any: the content of the log
func LogInfo(v ...any) {
	log.Println(v...)
}

// LogInfof logs an information to the terminal
// Printed in base color
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogInfof(format string, v ...any) {
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
//   - v ...any: the content of the log
func LogResult(v ...any) {
	log.Print(Green, fmt.Sprint(v...), Reset, "\n")
}

// LogResultf logs a result to the terminal
// Printed in green
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func LogResultf(format string, v ...any) {
	log.Printf(Green+format+Reset, v...)
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

// GetNumberErr returns the number of errors and timeouts
//
// Returns:
//   - int: number of errors
//   - int: number of timeouts
func GetNumberErr() (int, int) {
	return numberErr, numberTimeout
}
