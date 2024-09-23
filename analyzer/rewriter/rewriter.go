// Copyrigth (c) 2024 Erik Kassubek
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
	"errors"
)

const (
	exitCodeNone           = -1
	exitCodeStuckFinish    = 10
	exitCodeStuckWaitElem  = 11
	exitCodeStuckNoElem    = 12
	exitCodeElemEmptyTrace = 13
	exitCodeLeakUnbuf      = 20
	exitCodeLeakBuf        = 21
	exitCodeLeakMutex      = 22
	exitCodeLeakCond       = 23
	exitCodeLeakWG         = 24
	exitSendClose          = 30
	exitRecvClose          = 31
	exitNegativeWG         = 32
	exitCodeCyclic         = 41
)

/*
 * Create a new trace from the given bug
 * Args:
 *   bug (Bug): The bug to create a trace for
 * Returns:
 *   bool: true if rewrite was needed, false otherwise (e.g. actual bug, warning)
 *
 *   error: An error if the trace could not be created
 */
func RewriteTrace(bug bugs.Bug) (rewriteNeeded bool, code int, err error) {
	rewriteNeeded = false
	code = exitCodeNone
	switch bug.Type {
	case bugs.ASendOnClosed:
		err = errors.New("Actual send on closed in trace. Therefore no rewrite is needed.")
	case bugs.ARecvOnClosed:
		err = errors.New("Actual receive on closed in trace. Therefore no rewrite is needed.")
	case bugs.ACloseOnClosed:
		err = errors.New("Only actual close on close can be detected. Therefor no rewrite is needed.")
	case bugs.AConcurrentRecv:
		err = errors.New("Rewriting trace for concurrent receive is not possible")
	case bugs.ASelCaseWithoutPartner:
		err = errors.New("Rewriting trace for select without partner is not possible")
	case bugs.PSendOnClosed:
		code = exitSendClose
		rewriteNeeded = true
		err = rewriteClosedChannel(bug, exitSendClose)
	case bugs.PRecvOnClosed:
		code = exitRecvClose
		rewriteNeeded = true
		err = rewriteClosedChannel(bug, exitRecvClose)
	case bugs.PNegWG:
		code = exitNegativeWG
		rewriteNeeded = true
		err = rewriteWaitGroup(bug)
	// case bugs.MixedDeadlock:
	// 	err = errors.New("Rewriting trace for mixed deadlock is not implemented yet")
	// case bugs.CyclicDeadlock:
	// 	rewriteNeeded = true
	// err = rewriteCyclicDeadlock(bug)

	case bugs.LUnbufferedWith:
		code = exitCodeLeakUnbuf
		rewriteNeeded = true
		err = rewriteUnbufChanLeak(bug)
	case bugs.LUnbufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace.")
	case bugs.LBufferedWith:
		code = exitCodeLeakBuf
		rewriteNeeded = true
		err = rewriteBufChanLeak(bug)
	case bugs.LBufferedWithout:
		err = errors.New("No possible partner for stuck channel found. Cannot rewrite trace.")
	case bugs.LNilChan:
		err = errors.New("Leak on nil channel. Cannot rewrite trace.")
	case bugs.LSelectWith:
		code = exitCodeLeakUnbuf
		rewriteNeeded = true
		switch b := (*bug.TraceElement2[0]).(type) {
		case *analysis.TraceElementSelect:
			err = rewriteUnbufChanLeak(bug)
		case *analysis.TraceElementChannel:
			if b.IsBuffered() {
				err = rewriteBufChanLeak(bug)
			} else {
				err = rewriteUnbufChanLeak(bug)
			}
		default:
			rewriteNeeded = false
			code = exitCodeNone
			err = errors.New("For the given bug type no trace rewriting is possible")
		}
	case bugs.LSelectWithout:
		code = exitCodeNone
		err = errors.New("No possible partner for stuck select found. Cannot rewrite trace.")
	case bugs.LMutex:
		rewriteNeeded = true
		code = exitCodeLeakMutex
		err = rewriteMutexLeak(bug)
	case bugs.LWaitGroup:
		rewriteNeeded = true
		code = exitCodeLeakWG
		err = rewriteWaitGroupLeak(bug)
	case bugs.LCond:
		rewriteNeeded = true
		code = exitCodeLeakCond
		err = rewriteCondLeak(bug)
	default:
		err = errors.New("For the given bug type no trace rewriting is implemented")
	}
	if rewriteNeeded && err != nil {
		println("Error rewriting trace")
	}
	return rewriteNeeded, code, err
}
