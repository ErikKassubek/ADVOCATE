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
	analysisCasesFlag         map[string]bool
	ignoreAtomicsFlag         bool
	fifoFlag                  bool
	ignoreCriticalSectionFlag bool
	rewriteAllFlag            bool
	onlyAPanicAndLeakFlag     bool
	replayAllFlag             bool
)

// SetFlags makes the relevant command line arguments given to the analyzer
// locally available for the toolchain
//
// Parameter:
//   - noRewrite bool: do not rewrite found bugs
//   - analysisCases map[string]bool: set which analysis scenarios should be run
//   - ignoreAtomics bool: if true atomics are ignored for replay
//   - fifo bool: assume that channels work as fifo queue
//   - ignoreCriticalSection bool: ignore order of lock/unlock
//   - rewriteAll bool: rewrite bugs that have been confirmed before
//   - onlyAPanicAndLeak bool: do not run a HB analysis, but only detect actually occurring bugs
//   - timeoutRec int: timeout of recording in seconds
//   - timeoutRepl int: timeout of replay in seconds
func SetFlags(noRewrite bool, analysisCases map[string]bool, ignoreAtomics,
	fifo, ignoreCriticalSection, rewriteAll bool, onlyAPanicAndLeak bool,
	timeoutRec, timeoutRepl int, replayAll bool) {

	noRewriteFlag = noRewrite

	analysisCasesFlag = analysisCases

	ignoreAtomicsFlag = ignoreAtomics
	fifoFlag = fifo
	ignoreCriticalSectionFlag = ignoreCriticalSection
	rewriteAllFlag = rewriteAll
	onlyAPanicAndLeakFlag = onlyAPanicAndLeak

	timeoutRecording = timeoutRec
	timeoutReplay = timeoutRepl

	replayAllFlag = replayAll
}
