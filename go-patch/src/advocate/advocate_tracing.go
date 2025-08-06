// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Advocate tracing
//
// Author: Erik Kassubek
// Created: 2024-11-5
//
// License: BSD-3-Clause

package advocate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var traceFileCounter = 0
var tracePathRecorded = "advocateTrace"

var hasFinished = false

var timerStarted = false
var startTime time.Time
var duration time.Duration

// InitTracing initializes the tracing.
// The function creates the trace folder and starts the background memory test.
func InitTracing(timeout int) {
	startTime = time.Now()
	timerStarted = true

	if timeout > 0 {
		// start time timeout
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			panic("Timeout")
		}()
	}

	// go writeTraceIfFull()
	// go removeAtomicsIfFull()
	runtime.InitTracing(FinishTracing)
}

// Write the trace of the program to a file.
// The trace is written in the file named file_name.
// The trace is written in the format of advocate.
func FinishTracing() {
	if hasFinished {
		// needed to prevent program stop while still writing
		// otherwise, trace may be empty
		return
	}
	hasFinished = true

	if !finishFuzzingStarted {
		time.Sleep(time.Second)
	}

	// remove the trace folder if it exists
	err := os.RemoveAll(tracePathRecorded)
	if err != nil {
		if !os.IsNotExist(err) {
			println("Cannot remove: ", err.Error())
			return
		}
	}

	// create the trace folder
	err = os.Mkdir(tracePathRecorded, 0755)
	if err != nil {
		if !os.IsExist(err) {
			println("Cannot write: ", err.Error())
			return
		}
	}

	runtime.AdvocatRoutineExit()

	runtime.DisableTracing()

	if timerStarted {
		duration = time.Since(startTime)
	}

	writeToTraceFiles(tracePathRecorded)
}

// Write the trace to a set of files. The traces are written into a folder
// with name trace. For each routine, a file is created. The file is named
// trace_routineId.log. The trace of the routine is written into the file.
//
// Parameter:
//
//	tracePath string:t path to where the trace should be written
func writeToTraceFiles(tracePath string) {
	numRout := runtime.GetNumberOfRoutines()
	writeToTraceFileInfo(tracePath, numRout)

	for i := 1; i <= numRout; i++ {
		// write the trace to the file
		writeToTraceFile(i, tracePath)
	}
}

// Write the trace of a routine to a file.
// The trace is written in the file named trace_routineId.log.
// The trace is written in the format of advocate.
//
// Parameter:
//   - routine: The id of the routine
//   - tracePath string: path to where the trace should be written
func writeToTraceFile(routine int, tracePath string) {
	// create the file if it does not exist and open it
	fileName := filepath.Join(tracePath, "trace_"+strconv.Itoa(routine)+".log")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// get the runtime to send the trace
	advocateChan := make(chan string)
	go func() {
		runtime.TraceToStringByIDChannel(routine, advocateChan)
		close(advocateChan)
	}()

	// receive the trace and write it to the file
	for trace := range advocateChan {
		if _, err := file.WriteString(trace); err != nil {
			panic(err)
		}
	}
}

/*
 * Write a trace info file
 */
func writeToTraceFileInfo(tracePath string, numberRoutines int) {
	fileName := filepath.Join(tracePath, "trace_info.log")

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reachedActive := 0
	if runtime.NumberActiveReleased > 0 {
		reachedActive = 1
	}
	allActiveReleased := 0
	if runtime.NumberActiveReleased == runtime.NumberActive {
		allActiveReleased = 1
	}

	exitCode, exitPos := runtime.GetExitCode()
	replayOldest, replayDisabled, replayAck := runtime.GetReplayStatus()
	allActiveReleased = 0

	file.WriteString(fmt.Sprintf("ExitCode!%d\n", exitCode))
	file.WriteString(fmt.Sprintf("ExitPosition!%s\n", exitPos))
	file.WriteString(fmt.Sprintf("ReplayTimeout!%d\n", replayOldest))
	file.WriteString(fmt.Sprintf("ReplayDisabled!%d\n", replayDisabled))
	file.WriteString(fmt.Sprintf("ReplayAck!%d\n", replayAck))
	file.WriteString(fmt.Sprintf("NumberRoutines!%d\n", numberRoutines))
	file.WriteString(fmt.Sprintf("ActiveReached!%d\n", reachedActive))
	file.WriteString(fmt.Sprintf("AllActiveReleased!%d\n", allActiveReleased))
	if timerStarted {
		file.WriteString(fmt.Sprintf("Runtime!%d", int(duration.Seconds())))
	} else {
		file.WriteString("Runtime:0")
	}

}

// Delete empty files in the trace folder.
// The function deletes all files in the trace folder that are empty.
// func deleteEmptyFiles() {
// 	files, err := os.ReadDir(tracePathRecorded)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for _, file := range files {
// 		if file.IsDir() {
// 			continue
// 		}

// 		stat, err := os.Stat(tracePathRecorded + "/" + file.Name())
// 		if err != nil {
// 			continue
// 		}
// 		if stat.Size() == 0 {
// 			err := os.Remove(tracePathRecorded + "/" + file.Name())
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 	}

// }
