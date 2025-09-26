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
	NameReplayActive   = "replay_active.log"
	NameTimes          = "times.log"
	NameTraceInfo      = "trace_info.log"
	NameResultMachine  = "results_machine.log"
	NameResultReadable = "results_readable.log"
	NameRewrittenInfo  = "rewrite_info.log"
	NameStats          = "stats"
	NameStatsTime      = "times"
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

// Result
var (
	Result        = ""
	CurrentResult = ""
	ResultOutput  = ""
	Output        = ""
	ResultStats   = ""
	ResultTime    = ""
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
}

func pathsResult(main bool) {
	if main {
		Result = filepath.Join(ProgDir, NameResult)
	} else {
		Result = filepath.Join(Prog, NameResult)
	}
	CurrentResult = Result
	ResultOutput = filepath.Join(Result, NameOutput)
	Output = filepath.Join(ProgDir, NameOutput)
	ResultStats = filepath.Join(Result, NameStats)
	ResultTime = filepath.Join(Result, NameStatsTime)
}

func SetCurrentResult(fileNumber, testNumber int, fileName, testName string) string {
	dirName := fmt.Sprintf("file(%d)-test(%d)-%s-%s", fileNumber, testNumber, fileName, testName)
	CurrentResult = filepath.Join(Result, dirName)
	ResultOutput = filepath.Join(CurrentResult, NameOutput)
	ResultStats = filepath.Join(CurrentResult, NameStats)
	ResultTime = filepath.Join(CurrentResult, NameStatsTime)
	return CurrentResult
}
