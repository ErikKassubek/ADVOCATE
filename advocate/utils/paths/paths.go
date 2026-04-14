// Copyright (c) 2025 Erik Kassubek
//
// File: paths.go
// Brief: Important paths
//
// Author: Erik Kassubek
// Created: 2025-09-25
//
// License: BSD-3-Clause

package paths

import (
	"advocate/utils/flags"
	"advocate/utils/helper"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// names
const (
	NameResult         = "advocateResult"
	NameOutput         = "output.log"
	NameFuzzingData    = "fuzzingData.log"
	NameFuzzingTraces  = "fuzzingTraces"
	NameReplayActive   = "replay_active.log"
	NameTimes          = "times.log"
	NameTraceInfo      = "trace_info.log"
	NameResultMachine  = "results_machine.log"
	NameResultReadable = "results_readable.log"
	NameRewrittenInfo  = "rewrite_info.log"
	NameStats          = "stats"
	NameStatsTime      = "times"
	NameBugs           = "bugs"
	NameTraces         = "traces"
	NameOut            = "output"
)

// advocate
var (
	Advocate = ""
	GoPatch  = ""
	Go       = ""
)

// prog
var (
	Prog    = ""
	ProgDir = ""
)

// work
var (
	FuzzingTraces = ""
	Output        = ""
)

// Result
var (
	Result        = ""
	CurrentResult = ""
	ResultOutput  = ""
	ResultBugs    = ""
	ResultStats   = ""
	ResultTime    = ""
	ResultTraces  = ""
	ResultOut     = ""
)

func BuildPaths(main bool) {
	pathsAdvocate()
	pathsProg()
	pathsResult(main)
}

func pathsAdvocate() {
	execPath, _ := os.Executable()
	Advocate = helper.CleanPathHome(filepath.Dir(filepath.Dir(execPath)))
	GoPatch = filepath.Join(Advocate, "goPatch")
	Go = filepath.Join(Advocate, "goPatch/bin/go")

	if runtime.GOOS == "windows" {
		Go += ".exe"
	}
}

func pathsProg() {
	Prog = helper.CleanPathHome(flags.ProgPath)
	ProgDir = helper.GetDirectory(Prog) // only for main
	Output = filepath.Join(ProgDir, NameOutput)
	FuzzingTraces = filepath.Join(ProgDir, NameFuzzingTraces)
}

func pathsResult(main bool) {
	if main {
		Result = filepath.Join(ProgDir, NameResult)
	} else {
		Result = filepath.Join(Prog, NameResult)
	}
	CurrentResult = Result
	ResultOut = filepath.Join(CurrentResult, NameOut)
	ResultOutput = filepath.Join(ResultOut, NameOutput)
	ResultBugs = filepath.Join(CurrentResult, NameBugs)

	ResultStats = filepath.Join(Result, NameStats)
	ResultTime = filepath.Join(Result, NameStatsTime)
	ResultTraces = filepath.Join(Result, NameTraces)
}

func SetCurrentResult(fileNumber, testNumber int, fileName, testName string) string {
	dirName := fmt.Sprintf("file(%d)-test(%d)-%s-%s", fileNumber, testNumber, fileName, testName)
	CurrentResult = filepath.Join(Result, dirName)
	ResultOut = filepath.Join(CurrentResult, NameOut)
	ResultBugs = filepath.Join(CurrentResult, NameBugs)
	ResultOutput = filepath.Join(ResultOut, NameOutput)
	ResultStats = filepath.Join(CurrentResult, NameStats)
	ResultTime = filepath.Join(CurrentResult, NameStatsTime)
	ResultTraces = filepath.Join(CurrentResult, NameTraces)
	return CurrentResult
}
