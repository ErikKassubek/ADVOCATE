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

// Create a new trace from the given bug
//
// Parameter:
//   - trace *analysis.Trace: Pointer to the trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - rewrittenBugs *map[bugs.ResultType][]string: map of already rewritten bugs
//   - retwriteOnce bool: skip double bugs
//
// Returns:
//   - bool: true if rewrite was needed, false otherwise (e.g. actual bug, warning)
//   - skip: true if the rewrite can be skipped, because it was rewritten before
//   - code: expected exit code
//   - error: An error if the trace could not be created
func RewriteTrace(trace *analysis.Trace, bug bugs.Bug, rewrittenBugs map[bugs.ResultType][]string, rewriteOnce bool) (rewriteNeeded bool, skip bool, code int, err error) {
	if rewriteOnce {
		bugString := bug.GetBugString()
		if _, ok := rewrittenBugs[bug.Type]; !ok {
			rewrittenBugs[bug.Type] = make([]string, 0)
		} else {
			if utils.Contains((rewrittenBugs)[bug.Type], bugString) {
				return false, true, 0, nil
			}
		}
		rewrittenBugs[bug.Type] = append(rewrittenBugs[bug.Type], bugString)
	}

	rewriteNeeded = false
	code = analysis.ExitCodeNone
	switch bug.Type {
	case bugs.ASendOnClosed:
		err = errors.New("Actual send on closed. Therefore no rewrite is needed.")
	case bugs.ARecvOnClosed:
		err = errors.New("Actual receive on closed in trace. Therefore no rewrite is needed.")
	case bugs.ACloseOnClosed:
		err = errors.New("Actual close on close detected. Therefor no rewrite is needed.")
	case bugs.ACloseOnNil:
		err = errors.New("Actual close on nil detected. Therefor no rewrite is needed.")
	case bugs.ANegWG:
		err = errors.New("Actual negative Wait Group. Therefore no rewrite is needed.")
	case bugs.AUnlockOfNotLockedMutex:
		err = errors.New("Actual unlock of not locked mutex. Therefore no rewrite is needed.")
	case bugs.AConcurrentRecv:
		err = errors.New("Rewriting trace for concurrent receive is not possible")
	case bugs.ASelCaseWithoutPartner:
		err = errors.New("Rewriting trace for select without partner is not possible")
	case bugs.PSendOnClosed:
		code = analysis.ExitCodeSendClose
		rewriteNeeded = true
		err = rewriteClosedChannel(trace, bug, analysis.ExitCodeSendClose)
	case bugs.PRecvOnClosed:
		code = analysis.ExitCodeRecvClose
		rewriteNeeded = true
		err = rewriteClosedChannel(trace, bug, analysis.ExitCodeRecvClose)
	case bugs.PNegWG:
		code = analysis.ExitCodeNegativeWG
		rewriteNeeded = true
		err = rewriteGraph(trace, bug, code)
	case bugs.PUnlockBeforeLock:
		code = analysis.ExitCodeUnlockBeforeLock
		rewriteNeeded = true
		err = rewriteGraph(trace, bug, code)
	// case bugs.MixedDeadlock:
	// 	err = errors.New("Rewriting trace for mixed deadlock is not implemented yet")
	case bugs.PCyclicDeadlock:
		rewriteNeeded = true
		err = rewriteCyclicDeadlock(trace, bug)
	case bugs.LWithoutBlock:
		err = errors.New("Source of blocking not known. Therefore no rewrite is possible.")
	case bugs.LUnbufferedWith:
		code = analysis.ExitCodeLeakUnbuf
		rewriteNeeded = true
		err = rewriteUnbufChanLeak(trace, bug)
	case bugs.LUnbufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace.")
	case bugs.LBufferedWith:
		code = analysis.ExitCodeLeakBuf
		rewriteNeeded = true
		err = rewriteBufChanLeak(trace, bug)
	case bugs.LBufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace.")
	case bugs.LNilChan:
		err = errors.New("Leak on nil channel. Cannot rewrite trace.")
	case bugs.LSelectWith:
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
	case bugs.LSelectWithout:
		code = analysis.ExitCodeNone
		err = errors.New("No possible partner for stuck select found. Cannot rewrite trace.")
	case bugs.LMutex:
		rewriteNeeded = true
		code = analysis.ExitCodeLeakMutex
		err = rewriteMutexLeak(trace, bug)
	case bugs.LWaitGroup:
		rewriteNeeded = true
		code = analysis.ExitCodeLeakWG
		err = rewriteWaitGroupLeak(trace, bug)
	case bugs.LCond:
		rewriteNeeded = true
		code = analysis.ExitCodeLeakCond
		err = rewriteCondLeak(trace, bug)
		// case bugs.SNotExecutedWithPartner:
		// 	rewriteNeeded = false
		// 	err = errors.New("Rewrite for select not exec with partner not available")
	case bugs.RUnknownPanic:
		err = errors.New("Unknown panic. No rewrite possible.")
	case bugs.RTimeout:
		err = errors.New("Timeout. No rewrite possible.")
	default:
		err = errors.New("For the given bug type no trace rewriting is implemented")
	}
	return rewriteNeeded, false, code, err
}
