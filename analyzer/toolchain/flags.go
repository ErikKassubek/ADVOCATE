// Copyright (c) 2025 Erik Kassubek
//
// File: flags.go
// Brief: Store the flags needed in runAnalyzer
//
// Author: Erik Kassubek
// Created: 2025-02-04
//
// License: BSD-3-Clause

package toolchain

var (
	pathToAdvocate   string
	pathToFileOrDir  string
	programName      string
	executableName   string
	testName         string
	timeoutRecording int
	timeoutReplay    int
	numberRerecord   int
	replayAtomic     bool
	measureTime      bool
	notExecuted      bool
	createStats      bool

	noRewriteFlag             bool
	analyisCasesFlag          map[string]bool
	ignoreAtomicsFlag         bool
	fifoFlag                  bool
	ignoreCriticalSectionFlag bool
	rewriteAllFlag            bool
	ignoreRewriteFlag         string
	onlyAPanicAndLeakFlag     bool
)

func SetFlags(noRewrite bool, analyisCases map[string]bool, ignoreAtomics,
	fifo, ignoreCriticalSection, rewriteAll bool, ignoreRewrite string, onlyAPanicAndLeak bool,
	timeoutRec, timeoutRepl int) {

	noRewriteFlag = noRewrite

	analyisCasesFlag = analyisCases

	ignoreAtomicsFlag = ignoreAtomics
	fifoFlag = fifo
	ignoreCriticalSectionFlag = ignoreCriticalSection
	rewriteAllFlag = rewriteAll
	ignoreRewriteFlag = ignoreRewrite
	onlyAPanicAndLeakFlag = onlyAPanicAndLeak

	timeoutRecording = timeoutRec
	timeoutReplay = timeoutRepl

}
