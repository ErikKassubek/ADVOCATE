// Copyright (c) 2026 Erik Kassubek
//
// File: logging.go
// Brief: Logging function
//
// Author: Erik Kassubek
// Created: 2025-02-18
//
// License: BSD-3-Clause

package log

import (
	"advocate/utils/flags"
	"fmt"
	"log"
)

// Color codes for the logging output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
)

var numberErr = 0
var numberTimeout = 0
var numberResults = 0
var numberTestWithRes = 0
var numberResultsConf = 0

var seenTests = make(map[string]struct{})

var preventPanicFlag bool

var guiChanSet bool
var guiChan chan GuiInfo

type InfoLevel int

const (
	InfoLv InfoLevel = iota
	ImportantLv
	DebugLv
	ResultLv
	ProgressLv
	TimeoutLv
	ErrorLv
	GuiLv
	OutputLv
)

type GuiInfo struct {
	Msg string
	Lv  InfoLevel
}

func GetGuiChan() chan GuiInfo {
	guiChanSet = true
	guiChan = make(chan GuiInfo)
	return guiChan
}

// Info logs an information to the terminal
// Printed in base color
//
// Parameter:
//   - v ...any: the content of the log
func Info(v ...any) {
	if !flags.Verbose {
		return
	}

	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), InfoLv}
	} else {
		log.Println(v...)
	}
}

// Infof logs an information to the terminal
// Printed in base color
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Infof(format string, v ...any) {
	if !flags.Verbose {
		return
	}

	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprintf(format, v...), InfoLv}
	} else {
		log.Printf(format, v...)
	}
}

// Important logs an important information to the terminal
// Printed in yellow
//
// Parameter:
//   - v ...any: the content of the log
func Important(v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), ImportantLv}
	} else {
		log.Print(Yellow, fmt.Sprint(v...), Reset, "\n")
	}
}

// Importantf logs an important information to the terminal
// Printed in yellow
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Importantf(format string, v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprintf(format, v...), ImportantLv}
	} else {
		log.Printf(Yellow+format+Reset, v...)
	}
}

// Debug logs an debug information to the terminal
// Printed in yellow
//
// Parameter:
//   - v ...any: the content of the log
func Debug(v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), DebugLv}
	} else {
		log.Print(Yellow, fmt.Sprint(v...), Reset, "\n")
	}
}

// Debugf logs an debug information to the terminal
// Printed in yellow
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Debugf(format string, v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), DebugLv}
	} else {
		log.Printf(Yellow+format+Reset, v...)
	}
}

// Result logs a result to the terminal
// Printed in green
//
// Parameter:
//   - count bool: wether to count for the number of results
//   - confirmed bool: true of bug is actual or replay was suc, false otherwise\
//   - name string: unique id for the program or test
//   - v ...any: the content of the log
func Result(count, confirmed bool, name string, v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), ResultLv}
	} else {
		log.Print(Green, fmt.Sprint(v...), Reset, "\n")
	}
	if count {
		numberResults++
		if _, ok := seenTests[name]; name != "" && !ok {
			seenTests[name] = struct{}{}
			numberTestWithRes++
		}
		if confirmed {
			numberResultsConf++
		}
	}
}

// Resultf logs a result to the terminal
// Printed in green
//
// Parameter:
//   - count bool: wether to count for the number of results
//   - confirmed bool: true of bug is actual or replay was suc, false otherwise\
//   - name string: unique id for the program or test
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Resultf(count, confirmed bool, name string, format string, v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprintf(format, v...), ResultLv}
	} else {
		log.Printf(Green+format+Reset, v...)
	}
	if count {
		numberResults++
		if _, ok := seenTests[name]; name != "" && !ok {
			seenTests[name] = struct{}{}
			numberTestWithRes++
		}
		if confirmed {
			numberResultsConf++
		}
	}
}

// Progress logs a the progress to the terminal
// Printed in green
//
// Parameter:
//   - v ...any: the content of the log
func Progress(v ...any) {
	if flags.NoProgress {
		return
	}
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), ProgressLv}
	} else {
		log.Print(Blue, fmt.Sprint(v...), Reset, "\n")
	}
}

// Progressf logs a the progress to the terminal
// Printed in green
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Progressf(format string, v ...any) {
	if flags.NoProgress {
		return
	}
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprintf(format, v...), ProgressLv}
	} else {
		log.Printf(Blue+format+Reset, v...)
	}
}

// Timeout logs a timeout to the terminal
// Printed in purple
// Counts number of timeouts
//
// Parameter:
//   - v ...any: the content of the log
func Timeout(v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), TimeoutLv}
	} else {
		log.Print(Purple, fmt.Sprint(v...), Reset, "\n")
	}
	numberTimeout++
}

// Timeoutf logs a timeout to the terminal
// Printed in purple
// Counts number of timeouts
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Timeoutf(format string, v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprintf(format, v...), TimeoutLv}
	} else {
		log.Printf(Purple+format+Reset, v...)
	}
	numberTimeout++
}

// Error logs an error to the terminal
// Printed in red
// Counts number of error
//
// Parameter:
//   - v ...any: the content of the log
func Error(v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprint(v...), ErrorLv}
	} else {
		log.Print(Red, fmt.Sprint(v...), Reset, "\n")
	}
	numberErr++
}

// Errorf logs an error to the terminal
// Printed in red
// Counts number of error
//
// Parameter:
//   - format string: the format (e.g. "%s")
//   - v ...any: the content of the log
func Errorf(format string, v ...any) {
	if guiChanSet {
		guiChan <- GuiInfo{fmt.Sprintf(format, v...), ErrorLv}
	} else {
		log.Printf(Red+format+Reset, v...)
	}
	numberErr++
}

// GetLoggingNumbers returns the number of results, errors and timeouts
//
// Returns:
//   - int: number of results
//   - int: number of confirmed results
//   - int: number of tests with found bug
//   - int: number of errors
//   - int: number of timeouts
func GetLoggingNumbers() (int, int, int, int, int) {
	return numberResults, numberResultsConf, numberTestWithRes, numberErr, numberTimeout
}

// IsPanicPrevent returns if panic should be suppressed
//
// Returns:
//   - bool: true if panics should recover
func IsPanicPrevent() bool {
	return !flags.AlwaysPanic
}
