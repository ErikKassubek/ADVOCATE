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
	"analyzer/bugs"
	"analyzer/trace"
	"analyzer/utils"
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
func RewriteTrace(tr *trace.Trace, bug bugs.Bug, rewrittenBugs map[utils.ResultType][]string) (rewriteNeeded bool, code int, err error) {
	rewriteNeeded = false
	code = utils.ExitCodeNone
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
		code = utils.ExitCodeSendClose
		rewriteNeeded = true
		err = rewriteClosedChannel(tr, bug, utils.ExitCodeSendClose)
	case utils.PRecvOnClosed:
		code = utils.ExitCodeRecvClose
		rewriteNeeded = true
		err = rewriteClosedChannel(tr, bug, utils.ExitCodeRecvClose)
	case utils.PNegWG:
		code = utils.ExitCodeNegativeWG
		rewriteNeeded = true
		err = rewriteGraph(tr, bug, code)
	case utils.PUnlockBeforeLock:
		code = utils.ExitCodeUnlockBeforeLock
		rewriteNeeded = true
		err = rewriteGraph(tr, bug, code)
	// case bugs.MixedDeadlock:
	// 	err = errors.New("Rewriting trace for mixed deadlock is not implemented yet")
	case utils.PCyclicDeadlock:
		rewriteNeeded = true
		err = rewriteCyclicDeadlock(tr, bug)
	case utils.LWithoutBlock:
		err = errors.New("Source of blocking not known. Therefore no rewrite is possible")
	case utils.LUnbufferedWith:
		code = utils.ExitCodeLeakUnbuf
		rewriteNeeded = true
		err = rewriteUnbufChanLeak(tr, bug)
	case utils.LUnbufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace")
	case utils.LBufferedWith:
		code = utils.ExitCodeLeakBuf
		rewriteNeeded = true
		err = rewriteBufChanLeak(tr, bug)
	case utils.LBufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace")
	case utils.LNilChan:
		err = errors.New("Leak on nil channel. Cannot rewrite trace")
	case utils.LSelectWith:
		code = utils.ExitCodeLeakUnbuf
		rewriteNeeded = true
		switch b := bug.TraceElement2[0].(type) {
		case *trace.TraceElementSelect:
			err = rewriteUnbufChanLeak(tr, bug)
		case *trace.TraceElementChannel:
			if b.IsBuffered() {
				err = rewriteBufChanLeak(tr, bug)
			} else {
				err = rewriteUnbufChanLeak(tr, bug)
			}
		default:
			rewriteNeeded = false
			code = utils.ExitCodeNone
			err = errors.New("For the given bug type no trace rewriting is possible")
		}
	case utils.LSelectWithout:
		code = utils.ExitCodeNone
		err = errors.New("No possible partner for stuck select found. Cannot rewrite trace")
	case utils.LMutex:
		rewriteNeeded = true
		code = utils.ExitCodeLeakMutex
		err = rewriteMutexLeak(tr, bug)
	case utils.LWaitGroup:
		rewriteNeeded = true
		code = utils.ExitCodeLeakWG
		err = rewriteWaitGroupLeak(tr, bug)
	case utils.LCond:
		rewriteNeeded = true
		code = utils.ExitCodeLeakCond
		err = rewriteCondLeak(tr, bug)
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
