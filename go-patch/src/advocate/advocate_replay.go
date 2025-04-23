// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_replay.go
// Brief: Advocate Replay
//
// Author: Erik Kassubek
// Created: 2024-12-10
//
// License: BSD-3-Clause

package advocate

import (
	"bufio"
	"math"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var timeout = false
var tracePathRewritten = "rewrittenTrace_"

// InitReplay reads the trace from the trace folder.
// The function reads all files in the trace folder and adds the trace to the runtime.
// The trace is added to the runtime by calling the AddReplayTrace function.
//
// Parameter:
//   - index string: The index of the replay case
//   - exitCode bool: Whether the program should exit after the important replay part passed
//   - timeout int: Timeout in seconds, 0: no timeout
//   - atomic bool: if true, replay includes atomic
func InitReplay(index string, exitCode bool, timeout int, atomic bool) {
	// use first as default
	runtime.SetForceExit(exitCode)
	runtime.SetReplayAtomic(atomic) // set to true to include replay atomic

	if index == "0" {
		tracePathRewritten = "advocateTrace"
	} else {
		tracePathRewritten = tracePathRewritten + index
	}

	startReplay(timeout)
}

// startReplay reads the trace from the files. It also start the timeout for
// the replay
//
// Parameter:
//   - timeout int: timeout in seconds
func startReplay(timeout int) {
	println("Reading trace from " + tracePathRewritten)

	// if trace folder does not exist, panic
	if _, err := os.Stat(tracePathRewritten); os.IsNotExist(err) {
		panic("Trace folder " + tracePathRewritten + " does not exist.")
	}

	// traverse all files in the trace folder
	files, err := os.ReadDir(tracePathRewritten)
	if err != nil {
		panic(err)
	}

	chanWithoutPartner := make(map[string]int)

	foundTraceFiles := false
	for _, file := range files {
		// if the file is a directory, ignore it
		if file.IsDir() {
			continue
		}

		if file.Name() == "times.log" {
			continue
		}

		// if the file is a log file, read the trace
		if strings.HasSuffix(file.Name(), ".log") && file.Name() != "rewrite_info.log" {
			routineID, trace := readTraceFile(tracePathRewritten+"/"+file.Name(), &chanWithoutPartner)
			runtime.AddReplayTrace(uint64(routineID), trace)
			foundTraceFiles = true
		}
	}

	if !foundTraceFiles {
		panic("Could not find trace files for replay")
	}

	if timeout > 0 {
		// start time timeout
		go func() {
			time.Sleep(time.Duration(timeout) * time.Second)
			runtime.ExitReplayWithCode(runtime.ExitCodeTimeout)
			if runtime.IsAdvocateFuzzingEnabled() {
				FinishFuzzing()
			}
			panic("Timeout")
		}()
	}

	runtime.EnableReplay()
}

// Import the trace.
// The function creates the replay data structure, that is used to replay the trace.
// We only store the information that is needed to replay the trace.
// This includes operations on
//   - spawn
//   - channels
//   - mutexes
//   - once
//   - waitgroups
//   - select
//
// We only record the relevant information for each operation.
//
// Parameter:
//   - fileName string: The name of the file that contains the trace.
//   - chanWithoutPartner *map[string]int: pointer to map which stores channel
//     communication for which no partner has been read in yet
//
// Returns:
//   - int The routine id
//   - runtime.AdvocateReplayTrace: The trace for this routine
func readTraceFile(fileName string, chanWithoutPartner *map[string]int) (int, runtime.AdvocateReplayTrace) {
	// get the routine id from the file name
	routineID, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(fileName, tracePathRewritten+"/trace_"), ".log"))
	if err != nil {
		panic(err)
	}

	replayData := make(runtime.AdvocateReplayTrace, 0)

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		elem := scanner.Text()
		if elem == "" {
			continue
		}

		var time int
		var op runtime.Operation
		var file string
		var line int
		var pFile string
		var pLine int
		var blocked = false
		var suc = true
		var index int
		fields := strings.Split(elem, ",")
		time, _ = strconv.Atoi(fields[1])
		tPre, _ := strconv.Atoi(fields[1])
		switch fields[0] {
		case "X": // disable replay
			op = runtime.OperationReplayEnd
			line, _ = strconv.Atoi(fields[2]) // misuse the line for the exit code
			runtime.SetExpectedExitCode(line)
		case "G":
			op = runtime.OperationSpawn
			// time, _ = strconv.Atoi(fields[1])
			pos := strings.Split(fields[3], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
			index, _ = strconv.Atoi(fields[2])
		case "C":
			switch fields[4] {
			case "S":
				op = runtime.OperationChannelSend
			case "R":
				op = runtime.OperationChannelRecv
			case "C":
				op = runtime.OperationChannelClose
			default:
				panic("Unknown channel operation " + fields[4] + " in line " + elem + " in file " + fileName + ".")
			}
			time, _ = strconv.Atoi(fields[2])
			if time == 0 {
				blocked = true
			}
			pos := strings.Split(fields[9], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
			if !blocked && (op == runtime.OperationChannelSend || op == runtime.OperationChannelRecv) {
				index := findReplayPartner(fields[3], fields[6], len(replayData), chanWithoutPartner)
				if index != -1 && index < len(replayData) {
					pFile = replayData[index].File
					pLine = replayData[index].Line
					replayData[index].PFile = file
					replayData[index].PLine = line
				}
			}
		case "M":
			rw := false
			if fields[4] == "R" {
				rw = true
			}
			time, _ = strconv.Atoi(fields[2])
			if fields[6] == "f" {
				suc = false
			}
			pos := strings.Split(fields[7], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
			switch fields[5] {
			case "L":
				if rw {
					op = runtime.OperationRWMutexLock
				} else {
					op = runtime.OperationMutexLock
				}
			case "U":
				if rw {
					op = runtime.OperationRWMutexUnlock
				} else {
					op = runtime.OperationMutexUnlock
				}
			case "T":
				if rw {
					op = runtime.OperationRWMutexTryLock
				} else {
					op = runtime.OperationMutexTryLock
				}
			case "R":
				op = runtime.OperationRWMutexRLock
			case "N":
				op = runtime.OperationRWMutexRUnlock
			case "Y":
				op = runtime.OperationRWMutexTryRLock
			default:
				panic("Unknown mutex operation")
			}
			if fields[2] == "0" {
				blocked = true
			}

		case "O":
			op = runtime.OperationOnceDo
			// time, _ = strconv.Atoi(fields[1]) // read tpre to prevent false order
			if time == 0 {
				blocked = true
			}
			if fields[4] == "f" {
				suc = false
			}
			pos := strings.Split(fields[5], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
		case "W":
			switch fields[4] {
			case "W":
				op = runtime.OperationWaitgroupWait
			case "A":
				op = runtime.OperationWaitgroupAddDone
			default:
				panic("Unknown waitgroup operation")
			}
			time, _ = strconv.Atoi(fields[2])
			if time == 0 {
				blocked = true
			}
			pos := strings.Split(fields[7], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
		case "S":
			cases := strings.Split(fields[4], "~")
			if cases[len(cases)-1] == "D" {
				op = runtime.OperationSelectDefault
			} else {
				op = runtime.OperationSelectCase
			}
			time, _ = strconv.Atoi(fields[2])
			if time == 0 {
				blocked = true
			}
			index, _ = strconv.Atoi(fields[5])
			pos := strings.Split(fields[6], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
		case "D":
			switch fields[4] {
			case "W":
				op = runtime.OperationCondWait
			case "S":
				op = runtime.OperationCondSignal
			case "B":
				op = runtime.OperationCondBroadcast
			default:
				panic("Unknown cond operation: " + fields[4])
			}
			pos := strings.Split(fields[5], ":")
			file = pos[0]
			line, _ = strconv.Atoi(pos[1])
			if fields[2] == "0" {
				blocked = true
			}
		case "A":
			if !runtime.GetReplayAtomic() {
				continue
			}
			switch fields[3] {
			case "L":
				op = runtime.OperationAtomicLoad
			case "S":
				op = runtime.OperationAtomicStore
			case "A":
				op = runtime.OperationAtomicAdd
			case "W":
				op = runtime.OperationAtomicSwap
			case "C":
				op = runtime.OperationAtomicCompareAndSwap
			case "N":
				op = runtime.OperationAtomicAnd
			case "O":
				op = runtime.OperationAtomicOr
			}
			pos := strings.Split(fields[4], ":")
			if len(pos) < 2 {
				runtime.SetReplayAtomic(false)
				file = ""
				line = 0
			} else {
				file = pos[0]
				line, _ = strconv.Atoi(pos[1])
			}
		case "N": // new object
			continue
		case "E": // end of routine
			continue

		default:
			panic("Unknown operation " + fields[0] + " in line " + elem + " in file " + fileName + ".")
		}
		if blocked || time == 0 {
			time = math.MaxInt
		}
		if op != runtime.OperationNone && !runtime.AdvocateIgnoreReplay(op, file) {
			replayData = append(replayData, runtime.ReplayElement{
				Op: op, Routine: routineID, Time: time, TimePre: tPre, File: file, Line: line,
				Blocked: blocked, Suc: suc, PFile: pFile, PLine: pLine,
				Index: index})

		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// sort data by tpre
	sortReplayDataByTime(replayData)

	return routineID, replayData
}

// Find the partner of a channel operation.
// The partner is the operation that is executed on the other side of the channel.
//
// Parameter:
//   - cID string: channel id
//   - oID string: operation id
//   - index int: index of the operation in the replay data structure.
//   - chanWithoutPartner *map[string]int: pointer to map which stores channel
//     communication for which no partner has been read in yet
//
// Returns:
//   - int: The function returns the index of the partner operation.
//     If the partner operation is not found, the function returns -1.
func findReplayPartner(cID string, oID string, index int, chanWithoutPartner *map[string]int) int {
	opString := cID + ":" + oID
	if ind, ok := (*chanWithoutPartner)[opString]; ok {
		delete((*chanWithoutPartner), opString)
		return ind
	}

	(*chanWithoutPartner)[opString] = index
	return -1

}

// Sort the replay data structure by time.
//
// Parameter:
//   - replayData runtime.AdvocateReplayTrace: the data tp sport
//
// Returns:
//   - runtime.AdvocateReplayTrace: the sorted data
func sortReplayDataByTime(replayData runtime.AdvocateReplayTrace) runtime.AdvocateReplayTrace {
	sort.Slice(replayData, func(i, j int) bool {
		return replayData[i].Time < replayData[j].Time
	})
	return replayData
}

// FinishReplay waits for the replay to finish.
func FinishReplay() {
	if r := recover(); r != nil {
		println("Replay failed.")
	}

	runtime.WaitForReplayFinish(true)

	time.Sleep(time.Second)

	runtime.ExitReplayWithCode(runtime.ExitCodeDefault)
}

// func InitReplayTracing(index string, exitCode bool, timeout int, atomic bool) {
// 	if index == "-1" {
// 		InitTracing()
// 		return
// 	}

// 	tracePathRecorded = "advocateTraceReplay_" + index

// 	// if the program panics, but is not in the main routine, no trace is written
// 	// to prevent this, the following is done. The corresponding send/recv are in the panic definition
// 	blocked := make(chan struct{})
// 	writingDone := make(chan struct{})
// 	runtime.GetAdvocatePanicChannels(blocked, writingDone)
// 	go func() {
// 		<-blocked
// 		FinishReplayTracing()
// 		writingDone <- struct{}{}
// 	}()

// 	// if the program is terminated by the user, the defer in the header
// 	// is not executed. Therefore capture the signal and write the trace.
// 	interuptSignal := make(chan os.Signal, 1)
// 	signal.Notify(interuptSignal, os.Interrupt)
// 	go func() {
// 		<-interuptSignal
// 		println("\nCancel Run. Write trace. Cancel again to force exit.")
// 		go func() {
// 			<-interuptSignal
// 			os.Exit(1)
// 		}()
// 		if runtime.IsTracingEnabled() {
// 			FinishReplayTracing()
// 		}
// 		os.Exit(1)
// 	}()

// 	// go writeTraceIfFull()
// 	// go removeAtomicsIfFull()
// 	runtime.InitTracing(FinishTracing)

// 	InitReplay(index, exitCode, timeout, atomic)
// }

// func FinishReplayTracing() {
// 	if !runtime.IsReplayEnabled() {
// 		FinishTracing()
// 		return
// 	}

// 	if r := recover(); r != nil {
// 		println("Replay failed.")
// 	}

// 	runtime.WaitForReplayFinish(false)

// 	// runtime.DisableReplay()

// 	FinishTracing()
// }
