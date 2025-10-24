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
	"advocate/results/bugs"
	"advocate/trace"
	"advocate/utils/helper"
	"errors"
)

// RewriteTrace creates a new trace from the given bug
//
// Parameter:
//   - tr *trace.Trace: Pointer to the trace to rewrite
//   - bug Bug: The bug to create a trace for
//   - rewrittenBugs *map[bugs.ResultType][]string: map of already rewritten bugs
//
// Returns:
//   - bool: true if rewrite was needed, false otherwise (e.g. actual bug, warning)
//   - code: expected exit code
//   - error: An error if the trace could not be created
func RewriteTrace(tr *trace.Trace, bug bugs.Bug, rewrittenBugs map[helper.ResultType][]string) (rewriteNeeded bool, code int, err error) {
	rewriteNeeded = false
	code = helper.ExitCodeNone
	switch bug.Type {

	// ACTUAL BUGS
	case helper.ASendOnClosed:
		err = errors.New("Actual send on closed. Therefore no rewrite is needed")
	case helper.ARecvOnClosed:
		err = errors.New("Actual receive on closed in trace. Therefore no rewrite is needed")
	case helper.ACloseOnClosed:
		err = errors.New("Actual close on close detected. Therefor no rewrite is needed")
	case helper.ACloseOnNilChannel:
		err = errors.New("Actual close on nil detected. Therefor no rewrite is needed")
	case helper.ANegWG:
		err = errors.New("Actual negative Wait Group. Therefore no rewrite is needed")
	case helper.AUnlockOfNotLockedMutex:
		err = errors.New("Actual unlock of not locked mutex. Therefore no rewrite is needed")
	case helper.AConcurrentRecv:
		err = errors.New("Rewriting trace for concurrent receive is not possible")
	case helper.ASelCaseWithoutPartner:
		err = errors.New("Rewriting trace for select without partner is not possible")
	// MIXED DEADLOCK [REMOVE]
	case helper.AMixedDeadlock:
		err = errors.New("Actual mixed deadlock found in trace. No rewrite needed")

	// POSSIBLE BUGS
	case helper.PSendOnClosed:
		code = helper.ExitCodeSendClose
		rewriteNeeded = true
		err = rewriteClosedChannel(tr, bug, helper.ExitCodeSendClose)
	case helper.PRecvOnClosed:
		code = helper.ExitCodeRecvClose
		rewriteNeeded = true
		err = rewriteClosedChannel(tr, bug, helper.ExitCodeRecvClose)
	case helper.PNegWG:
		code = helper.ExitCodeNegativeWG
		rewriteNeeded = true
		err = rewriteGraph(tr, bug, code)
	case helper.PUnlockBeforeLock:
		code = helper.ExitCodeUnlockBeforeLock
		rewriteNeeded = true
		err = rewriteGraph(tr, bug, code)
	case helper.PCyclicDeadlock:
		rewriteNeeded = true
		err = rewriteCyclicDeadlock(tr, bug)
	// MIXED DEADLOCK [REMOVE]
	case helper.PMixedDeadlock:
		rewriteNeeded = true
		err = errors.New("Rewriting trace for mixed deadlock is not implemented yet")
	case helper.LUnknown:
		err = errors.New("Source of blocking not known. Therefore no rewrite is possible")
	case helper.LUnbufferedWith:
		code = helper.ExitCodeLeakUnbuf
		rewriteNeeded = true
		err = rewriteUnbufChanLeak(tr, bug)
	case helper.LUnbufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace")
	case helper.LBufferedWith:
		code = helper.ExitCodeLeakBuf
		rewriteNeeded = true
		err = rewriteBufChanLeak(tr, bug)
	case helper.LBufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace")
	case helper.LNilChan:
		err = errors.New("Leak on nil channel. Cannot rewrite trace")
	case helper.LSelectWith:
		code = helper.ExitCodeLeakUnbuf
		rewriteNeeded = true
		switch b := bug.TraceElement2[0].(type) {
		case *trace.ElementSelect:
			err = rewriteUnbufChanLeak(tr, bug)
		case *trace.ElementChannel:
			if b.IsBuffered() {
				err = rewriteBufChanLeak(tr, bug)
			} else {
				err = rewriteUnbufChanLeak(tr, bug)
			}
		default:
			rewriteNeeded = false
			code = helper.ExitCodeNone
			err = errors.New("For the given bug type no trace rewriting is possible")
		}
	case helper.LSelectWithout:
		code = helper.ExitCodeNone
		err = errors.New("No possible partner for stuck select found. Cannot rewrite trace")
	case helper.LMutex:
		rewriteNeeded = true
		code = helper.ExitCodeLeakMutex
		err = rewriteMutexLeak(tr, bug)
	case helper.LWaitGroup:
		rewriteNeeded = true
		code = helper.ExitCodeLeakWG
		err = rewriteWaitGroupLeak(tr, bug)
	case helper.LCond:
		rewriteNeeded = true
		code = helper.ExitCodeLeakCond
		err = rewriteCondLeak(tr, bug)
	case helper.LContext:
		err = errors.New("For the given bug type no trace rewriting is possible")
		// case bugs.SNotExecutedWithPartner:
		// 	rewriteNeeded = false
		// 	err = errors.New("Rewrite for select not exec with partner not available")
	case helper.RUnknownPanic:
		err = errors.New("Unknown panic. No rewrite possible")
	case helper.RTimeout:
		err = errors.New("Timeout. No rewrite possible")
	default:
		err = errors.New("For the given bug type no trace rewriting is implemented")
	}
	return rewriteNeeded, code, err
}
