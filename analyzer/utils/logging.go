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

func LogInfo(v ...any) {
	log.Println(v...)
}

func LogInfof(format string, v ...any) {
	log.Printf(format, v...)
}

func LogImportant(v ...any) {
	log.Print(Yellow, fmt.Sprint(v...), Reset, "\n")
}

func LogImportantf(format string, v ...any) {
	log.Printf(Yellow+format+Reset, v...)
}

func LogResult(v ...any) {
	log.Print(Green, fmt.Sprint(v...), Reset, "\n")
}

func LogResultf(format string, v ...any) {
	log.Printf(Green+format+Reset, v...)
}

func LogTimeout(v ...any) {
	log.Print(Purple, fmt.Sprint(v...), Reset, "\n")
	numberTimeout++
}

func LogTimeoutf(format string, v ...any) {
	log.Printf(Purple+format+Reset, v...)
	numberTimeout++
}

func LogError(v ...any) {
	log.Print(Red, fmt.Sprint(v...), Reset, "\n")
	numberErr++
}

func LogErrorf(format string, v ...any) {
	log.Printf(Red+format+Reset, v...)
	numberErr++
}

func GetNumberErr() (int, int) {
	return numberErr, numberTimeout
}
