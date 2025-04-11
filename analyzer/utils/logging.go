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

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Purple = "\033[35m"
)

var numberErr = 0
var numberTimeout = 0

/*
 * Log function for information
 * Printed in base color
 * Args:
 * 	v (...any): the content of the log
 */
func LogInfo(v ...any) {
	log.Println(v...)
}

/*
 * Formatted log function for information
 * Printed in base color
 * Args:
 * 	format (string): the format (e.g. "%s")
 * 	v (...any): the content of the log
 */
func LogInfof(format string, v ...any) {
	log.Printf(format, v...)
}

/*
 * Log function for important information
 * Printed in yellow
 * Args:
 * 	v (...any): the content of the log
 */
func LogImportant(v ...any) {
	log.Print(Yellow, fmt.Sprint(v...), Reset, "\n")
}

/*
 * Formatted log function for important information
 * Printed in yellow
 * Args:
 * 	format (string): the format (e.g. "%s")
 * 	v (...any): the content of the log
 */
func LogImportantf(format string, v ...any) {
	log.Printf(Yellow+format+Reset, v...)
}

/*
 * Log function for results
 * Printed in green
 * Args:
 * 	v (...any): the content of the log
 */
func LogResult(v ...any) {
	log.Print(Green, fmt.Sprint(v...), Reset, "\n")
}

/*
 * Formatted log function for results
 * Printed in green
 * Args:
 * 	format (string): the format (e.g. "%s")
 * 	v (...any): the content of the log
 */
func LogResultf(format string, v ...any) {
	log.Printf(Green+format+Reset, v...)
}

/*
 * Log function for timeout
 * Printed in purple
 * Counts number of timeouts
 * Args:
 * 	v (...any): the content of the log
 */
func LogTimeout(v ...any) {
	log.Print(Purple, fmt.Sprint(v...), Reset, "\n")
	numberTimeout++
}

/*
 * Formatted log function for timeout
 * Printed in purple
 * Counts number of timeouts
 * Args:
 * 	format (string): the format (e.g. "%s")
 * 	v (...any): the content of the log
 */
func LogTimeoutf(format string, v ...any) {
	log.Printf(Purple+format+Reset, v...)
	numberTimeout++
}

/*
 * Log function for errors
 * Printed in red
 * Counts number of error
 * Args:
 * 	v (...any): the content of the log
 */
func LogError(v ...any) {
	log.Print(Red, fmt.Sprint(v...), Reset, "\n")
	numberErr++
}

/*
 * Formatted log function for errors
 * Printed in red
 * Counts number of error
 * Args:
 * 	format (string): the format (e.g. "%s")
 * 	v (...any): the content of the log
 */
func LogErrorf(format string, v ...any) {
	log.Printf(Red+format+Reset, v...)
	numberErr++
}

/*
 * GetNumberErr returns the number of errors and timeouts
 * Returns:
 * 	int: number of errors
 * 	int: number of timeouts
 */
func GetNumberErr() (int, int) {
	return numberErr, numberTimeout
}
