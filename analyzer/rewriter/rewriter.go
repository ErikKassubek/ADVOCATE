// Copyright (c) 2024 Erik Kassubek
//
// File: rewriter.go
// Brief: Main functions to rewrite the trace
//
// Author: Erik Kassubek
// Created: 2023-11-30
//
// License: BSD-3-Clause

// Package rewriter provides functions for rewriting traces.
package rewriter

import (
	"analyzer/analysis"
	"analyzer/bugs"
	"analyzer/utils"
	"errors"
)

// RewriteTrace creates a new trace from the given bug
//
// Parameter:
//   - trace *analysis.Trace: Pointer to the trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - rewrittenBugs *map[bugs.ResultType][]string: map of already rewritten bugs
//
// Returns:
//   - bool: true if rewrite was needed, false otherwise (e.g. actual bug, warning)
//   - code: expected exit code
//   - error: An error if the trace could not be created
func RewriteTrace(trace *analysis.Trace, bug bugs.Bug, rewrittenBugs map[utils.ResultType][]string) (rewriteNeeded bool, code int, err error) {
	// if rewriteOnce {
	// 	bugString := bug.GetBugString()
	// 	if _, ok := rewrittenBugs[bug.Type]; !ok {
	// 		rewrittenBugs[bug.Type] = make([]string, 0)
	// 	} else {
	// 		if utils.Contains((rewrittenBugs)[bug.Type], bugString) {
	// 			return false, true, 0, nil
	// 		}
	// 	}
	// 	rewrittenBugs[bug.Type] = append(rewrittenBugs[bug.Type], bugString)
	// }

	rewriteNeeded = false
	code = analysis.ExitCodeNone
	switch bug.Type {
	case utils.ASendOnClosed:
		err = errors.New("Actual send on closed. Therefore no rewrite is needed")
	case utils.ARecvOnClosed:
		err = errors.New("Actual receive on closed in trace. Therefore no rewrite is needed")
	case utils.ACloseOnClosed:
		err = errors.New("Actual close on close detected. Therefor no rewrite is needed")
	case utils.ACloseOnNilChannel:
		err = errors.New("Actual close on nil detected. Therefor no rewrite is needed")
	case utils.ANegWG:
		err = errors.New("Actual negative Wait Group. Therefore no rewrite is needed")
	case utils.AUnlockOfNotLockedMutex:
		err = errors.New("Actual unlock of not locked mutex. Therefore no rewrite is needed")
	case utils.AConcurrentRecv:
		err = errors.New("Rewriting trace for concurrent receive is not possible")
	case utils.ASelCaseWithoutPartner:
		err = errors.New("Rewriting trace for select without partner is not possible")
	case utils.PSendOnClosed:
		code = analysis.ExitCodeSendClose
		rewriteNeeded = true
		err = rewriteClosedChannel(trace, bug, analysis.ExitCodeSendClose)
	case utils.PRecvOnClosed:
		code = analysis.ExitCodeRecvClose
		rewriteNeeded = true
		err = rewriteClosedChannel(trace, bug, analysis.ExitCodeRecvClose)
	case utils.PNegWG:
		code = analysis.ExitCodeNegativeWG
		rewriteNeeded = true
		err = rewriteGraph(trace, bug, code)
	case utils.PUnlockBeforeLock:
		code = analysis.ExitCodeUnlockBeforeLock
		rewriteNeeded = true
		err = rewriteGraph(trace, bug, code)
	// case bugs.MixedDeadlock:
	// 	err = errors.New("Rewriting trace for mixed deadlock is not implemented yet")
	case utils.PCyclicDeadlock:
		rewriteNeeded = true
		err = rewriteCyclicDeadlock(trace, bug)
	case utils.LWithoutBlock:
		err = errors.New("Source of blocking not known. Therefore no rewrite is possible")
	case utils.LUnbufferedWith:
		code = analysis.ExitCodeLeakUnbuf
		rewriteNeeded = true
		err = rewriteUnbufChanLeak(trace, bug)
	case utils.LUnbufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace")
	case utils.LBufferedWith:
		code = analysis.ExitCodeLeakBuf
		rewriteNeeded = true
		err = rewriteBufChanLeak(trace, bug)
	case utils.LBufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace")
	case utils.LNilChan:
		err = errors.New("Leak on nil channel. Cannot rewrite trace")
	case utils.LSelectWith:
		code = analysis.ExitCodeLeakUnbuf
		rewriteNeeded = true
		switch b := bug.TraceElement2[0].(type) {
		case *analysis.TraceElementSelect:
			err = rewriteUnbufChanLeak(trace, bug)
		case *analysis.TraceElementChannel:
			if b.IsBuffered() {
				err = rewriteBufChanLeak(trace, bug)
			} else {
				err = rewriteUnbufChanLeak(trace, bug)
			}
		default:
			rewriteNeeded = false
			code = analysis.ExitCodeNone
			err = errors.New("For the given bug type no trace rewriting is possible")
		}
	case utils.LSelectWithout:
		code = analysis.ExitCodeNone
		err = errors.New("No possible partner for stuck select found. Cannot rewrite trace")
	case utils.LMutex:
		rewriteNeeded = true
		code = analysis.ExitCodeLeakMutex
		err = rewriteMutexLeak(trace, bug)
	case utils.LWaitGroup:
		rewriteNeeded = true
		code = analysis.ExitCodeLeakWG
		err = rewriteWaitGroupLeak(trace, bug)
	case utils.LCond:
		rewriteNeeded = true
		code = analysis.ExitCodeLeakCond
		err = rewriteCondLeak(trace, bug)
		// case bugs.SNotExecutedWithPartner:
		// 	rewriteNeeded = false
		// 	err = errors.New("Rewrite for select not exec with partner not available")
	case utils.RUnknownPanic:
		err = errors.New("Unknown panic. No rewrite possible")
	case utils.RTimeout:
		err = errors.New("Timeout. No rewrite possible")
	default:
		err = errors.New("For the given bug type no trace rewriting is implemented")
	}
	return rewriteNeeded, code, err
}
