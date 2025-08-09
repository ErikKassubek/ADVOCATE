//
// File: flags.go
// Brief: Store the flags needed in runAnalyzer
//
// Created: 2025-02-04
//
// License: BSD-3-Clause

package toolchain

var (
	pathToGoCR       string
	pathToFileOrDir  string
	programName      string
	executableName   string
	testName         string
	timeoutRecording int
	timeoutReplay    int
	numberRerecord   int
	replayAtomic     bool
	measureTime      bool
	createStats      bool

	ignoreAtomicsFlag         bool
	fifoFlag                  bool
	ignoreCriticalSectionFlag bool
	noWarningFlag             bool
	tracePathFlag             string

	outputFlag bool
)

// SetFlags makes the relevant command line arguments given to the analyzer
// locally available for the toolchain
//
// Parameter:
//   - noRewrite bool: do not rewrite found bugs
//   - ignoreAtomics bool: if true atomics are ignored for replay
//   - fifo bool: assume that channels work as fifo queue
//   - ignoreCriticalSection bool: ignore order of lock/unlock
//   - rewriteAll bool: rewrite bugs that have been confirmed before
//   - timeoutRec int: timeout of recording in seconds
//   - timeoutRepl int: timeout of replay in seconds
//   - tracePath string: path to the trace for replay mode
func SetFlags(ignoreAtomics,
	fifo, ignoreCriticalSection bool,
	timeoutRec, timeoutRepl int, noWarning bool,
	tracePath string, output bool) {

	ignoreAtomicsFlag = ignoreAtomics
	fifoFlag = fifo
	ignoreCriticalSectionFlag = ignoreCriticalSection

	timeoutRecording = timeoutRec
	timeoutReplay = timeoutRepl

	noWarningFlag = noWarning

	tracePathFlag = tracePath

	outputFlag = output
}
