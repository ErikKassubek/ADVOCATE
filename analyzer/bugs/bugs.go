// Copyright (c) 2024 Erik Kassubek
//
// File: bugs.go
// Brief: Operations for handeling found bugs
//
// Author: Erik Kassubek
// Created: 2023-11-30
//
// License: BSD-3-Clause

package bugs

import (
	"analyzer/analysis"
	"analyzer/utils"
	"errors"
	"sort"
	"strconv"
	"strings"
)

// BugElementSelectCase is a type to store a specific id in a select
//
// Parameter:
//   - ID int: id of the involved channel
//   - ObjType string: object type
//   - Index int: internal index of the select int the case
type BugElementSelectCase struct {
	ID      int
	ObjType string
	Index   int
}

// GetBugElementSelectCase builds a BugElementSelectCase from a string
//
// Parameter:
//   - arg string: the string representing the case
//
// Returns:
//   - BugElementSelectCase: the bug select as a BugElementSelectCase
//   - error
func GetBugElementSelectCase(arg string) (BugElementSelectCase, error) {
	elems := strings.Split(arg, ":")
	id, err := strconv.Atoi(elems[1])
	if err != nil {
		return BugElementSelectCase{}, err
	}
	objType := elems[2]
	index, err := strconv.Atoi(elems[3])
	if err != nil {
		return BugElementSelectCase{}, err
	}
	return BugElementSelectCase{id, objType, index}, nil
}

// Bug is a type to describe and store a found bug
//
// Parameter:
//   - Type ResultType: The type of the bug
//   - TraceElement1 []analysis.TraceElement: first list of trace element involved in the bug
//     normally the elements that actually cause the bug, e.g. for send on close the send
//   - TraceElement2 []analysis.TraceElement: second list of trace element involved in the bug
//     normally the elements indirectly involved or elements to solve the bug (possible partner),
//     e.g. for send on close the close
type Bug struct {
	Type          utils.ResultType
	TraceElement1 []analysis.TraceElement
	// TraceElement1Sel []BugElementSelectCase
	TraceElement2 []analysis.TraceElement
}

// GetBugString Convert the bug to a string. Mostly used internally
//
// Returns:
//   - string: The bug as a string
func (b Bug) GetBugString() string {
	paths := make([]string, 0)

	for _, t := range b.TraceElement1 {
		paths = append(paths, t.GetPos())
	}
	for _, t := range b.TraceElement2 {
		paths = append(paths, t.GetPos())
	}

	sort.Strings(paths)

	res := string(b.Type)
	for _, path := range paths {
		res += path
	}
	return res
}

// ToString convert the bug to a string. Mostly used for output
//
// Returns:
//   - string: The bug as a string
func (b Bug) ToString() string {
	typeStr := ""
	arg1Str := ""
	arg2Str := ""
	switch b.Type {
	case utils.RUnknownPanic:
		typeStr = "Unknown Panic:"
		arg1Str = "Panic: "
	case utils.RTimeout:
		typeStr = "Timeout"
	case utils.ASendOnClosed:
		typeStr = "Actual Send on Closed Channel:"
		arg1Str = "send: "
		arg2Str = "close: "
	case utils.ARecvOnClosed:
		typeStr = "Actual Receive on Closed Channel:"
		arg1Str = "recv: "
		arg2Str = "close: "
	case utils.ACloseOnClosed:
		typeStr = "Actual Close on Closed Channel:"
		arg1Str = "close: "
		arg2Str = "close: "
	case utils.ACloseOnNilChannel:
		typeStr = "Actual close on nil channel:"
		arg1Str = "close: "
		arg2Str = "close: "
	case utils.AConcurrentRecv:
		typeStr = "Concurrent Receive:"
		arg1Str = "recv: "
		arg2Str = "recv: "
	case utils.ASelCaseWithoutPartner:
		typeStr = "Select Case without Partner:"
		arg1Str = "select: "
		arg2Str = "case: "
	case utils.ANegWG:
		typeStr = "Actual negative Wait Group:"
		arg1Str = "done: "
	case utils.AUnlockOfNotLockedMutex:
		typeStr = "Actual unlock of not locked mutex:"
		arg1Str = "unlock:"
	case utils.PSendOnClosed:
		typeStr = "Possible send on closed channel:"
		arg1Str = "send: "
		arg2Str = "close: "
	case utils.PRecvOnClosed:
		typeStr = "Possible receive on closed channel:"
		arg1Str = "recv: "
		arg2Str = "close: "
	case utils.PNegWG:
		typeStr = "Possible negative waitgroup counter:"
		arg1Str = "done: "
		arg2Str = "add: "
	case utils.PUnlockBeforeLock:
		typeStr = "Possible unlock of a not locked mutex:"
		arg1Str = "unlocks: "
		arg2Str = "locks: "
	case utils.PCyclicDeadlock:
		typeStr = "Possible cyclic deadlock:"
		arg1Str = "head: "
		arg2Str = "tail: "
	case utils.LWithoutBlock:
		typeStr = "Leak on routine without any blocking operation"
		arg1Str = "fork: "
	case utils.LUnbufferedWith:
		typeStr = "Leak on unbuffered channel with possible partner:"
		arg1Str = "channel: "
		arg2Str = "partner: "
	case utils.LUnbufferedWithout:
		typeStr = "Leak on unbuffered channel without possible partner:"
		arg1Str = "channel: "
	case utils.LBufferedWith:
		typeStr = "Leak on buffered channel with possible partner:"
		arg1Str = "channel: "
		arg2Str = "partner: "
	case utils.LBufferedWithout:
		typeStr = "Leak on buffered channel without possible partner:"
		arg1Str = "channel: "
	case utils.LNilChan:
		typeStr = "Leak on nil channel:"
		arg1Str = "channel: "
	case utils.LSelectWith:
		typeStr = "Leak on select with possible partner:"
		arg1Str = "select: "
		arg2Str = "partner: "
	case utils.LSelectWithout:
		typeStr = "Leak on select without partner:"
		arg1Str = "select: "
	case utils.LMutex:
		typeStr = "Leak on mutex:"
		arg1Str = "mutex: "
		arg2Str = "last: "
	case utils.LWaitGroup:
		typeStr = "Leak on wait group:"
		arg1Str = "waitgroup: "
	case utils.LCond:
		typeStr = "Leak on conditional variable:"
		arg1Str = "cond: "
	// case utils.SNotExecutedWithPartner:
	// 	typeStr = "Not executed select with potential partner"
	// 	arg1Str = "select: "
	// 	arg2Str = "partner: "

	default:
		utils.LogError("Unknown bug type in toString: " + string(b.Type))
		return ""
	}

	res := typeStr + "\n\t" + arg1Str
	for i, elem := range b.TraceElement1 {
		if i != 0 {
			res += ";"
		}
		res += elem.GetTID()
	}

	if arg2Str != "" {
		res += "\n\t" + arg2Str

		if len(b.TraceElement2) == 0 {
			res += "-"
		}

		for i, elem := range b.TraceElement2 {
			if i != 0 {
				res += ";"
			}
			res += elem.GetTID()
		}
	}

	return res
}

// Println prints the bug
func (b Bug) Println() {
	println(b.ToString())
}

// ProcessBug processes the bug that was selected from the analysis results
//
// Parameter:
//   - bugStr: The bug that was selected
//
// Returns:
//   - bool: true, if the bug was not a possible, but a actually occuring bug
//     Bug: The bug that was selected
//     error: An error if the bug could not be processed
func ProcessBug(bugStr string) (bool, Bug, error) {
	bug := Bug{}

	bugSplit := strings.Split(bugStr, ",")
	if len(bugSplit) != 3 && len(bugSplit) != 2 {
		return false, bug, errors.New("Could not split bug: " + bugStr)
	}

	bugType := bugSplit[0]

	containsArg1 := true
	containsArg2 := true
	actual := false

	switch bugType {
	case "R01":
		bug.Type = utils.RUnknownPanic
		actual = true
	case "R02":
		bug.Type = utils.RTimeout
		actual = true
	case "A01":
		bug.Type = utils.ASendOnClosed
		actual = true
	case "A02":
		bug.Type = utils.ARecvOnClosed
		actual = true
	case "A03":
		bug.Type = utils.ACloseOnClosed
		actual = true
	case "A04":
		bug.Type = utils.ACloseOnNil
		actual = true
	case "A05":
		bug.Type = utils.ANegWG
		actual = true
	case "A06":
		bug.Type = utils.AUnlockOfNotLockedMutex
		actual = true
	case "A07":
		bug.Type = utils.AConcurrentRecv
		actual = true
	case "A08":
		bug.Type = utils.ASelCaseWithoutPartner
		actual = true
	case "P01":
		bug.Type = utils.PSendOnClosed
	case "P02":
		bug.Type = utils.PRecvOnClosed
	case "P03":
		bug.Type = utils.PNegWG
	case "P04":
		bug.Type = utils.PUnlockBeforeLock
	case "P05":
		bug.Type = utils.PCyclicDeadlock
	// case "P06":
	// 	bug.Type = MixedDeadlock
	case "L00":
		containsArg1 = false
		bug.Type = utils.LWithoutBlock
	case "L01":
		bug.Type = utils.LUnbufferedWith
	case "L02":
		bug.Type = utils.LUnbufferedWithout
		containsArg2 = false
	case "L03":
		bug.Type = utils.LBufferedWith
	case "L04":
		bug.Type = utils.LBufferedWithout
		containsArg2 = false
	case "L05":
		bug.Type = utils.LNilChan
		containsArg2 = false
	case "L06":
		bug.Type = utils.LSelectWith
	case "L07":
		bug.Type = utils.LSelectWithout
		containsArg2 = false
	case "L08":
		bug.Type = utils.LMutex
	case "L09":
		bug.Type = utils.LWaitGroup
		containsArg2 = false
	case "L10":
		bug.Type = utils.LCond
		containsArg2 = false
	// case "S00":
	// 	bug.Type = SNotExecutedWithPartner
	// 	containsArg2 = true
	default:
		return actual, bug, errors.New("Unknown bug type in process bug: " + bugStr)
	}

	if !containsArg1 {
		return actual, bug, nil
	}

	bugArg1 := bugSplit[1]
	bugArg2 := ""
	if containsArg2 && len(bugSplit) == 3 {
		bugArg2 = bugSplit[2]
	}

	bug.TraceElement1 = make([]analysis.TraceElement, 0)
	// bug.TraceElement1Sel = make([]BugElementSelectCase, 0)

	for _, bugArg := range strings.Split(bugArg1, ";") {
		if strings.TrimSpace(bugArg) == "" {
			continue
		}

		if strings.HasPrefix(bugArg, "T") {
			elem, err := analysis.GetTraceElementFromBugArg(bugArg)
			if err != nil {
				return actual, bug, err
			}
			bug.TraceElement1 = append(bug.TraceElement1, elem)
		}
		// else if strings.HasPrefix(bugArg, "S") {
		// 	elem, err := GetBugElementSelectCase(bugArg)
		// 	if err != nil {
		// 		println("Could not read: " + bugArg + " from results")
		// 		return actual, bug, err
		// 	}
		// 	// bug.TraceElement1Sel = append(bug.TraceElement1Sel, elem)
		// }
	}

	bug.TraceElement2 = make([]analysis.TraceElement, 0)

	if !containsArg2 {
		return actual, bug, nil
	}

	for _, bugArg := range strings.Split(bugArg2, ";") {
		if strings.TrimSpace(bugArg) == "" {
			continue
		}

		if bugArg[0] == 'T' {
			elem, err := analysis.GetTraceElementFromBugArg(bugArg)
			if err != nil {
				return actual, bug, err
			}

			bug.TraceElement2 = append(bug.TraceElement2, elem)
		}
	}

	return actual, bug, nil
}
