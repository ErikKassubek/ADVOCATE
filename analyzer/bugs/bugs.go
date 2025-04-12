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

// ResultType is a type for the possible type of a fond bug including its id
type ResultType string

const (
	Empty ResultType = ""

	// actual
	ASendOnClosed           ResultType = "A01"
	ARecvOnClosed           ResultType = "A02"
	ACloseOnClosed          ResultType = "A03"
	ACloseOnNil             ResultType = "A04"
	ANegWG                  ResultType = "A05"
	AUnlockOfNotLockedMutex ResultType = "A06"
	AConcurrentRecv         ResultType = "A07"
	ASelCaseWithoutPartner  ResultType = "A08"

	// possible
	PSendOnClosed     ResultType = "P01"
	PRecvOnClosed     ResultType = "P02"
	PNegWG            ResultType = "P03"
	PUnlockBeforeLock ResultType = "P04"
	PCyclicDeadlock   ResultType = "P05"

	// leaks
	LWithoutBlock      = "L00"
	LUnbufferedWith    = "L01"
	LUnbufferedWithout = "L02"
	LBufferedWith      = "L03"
	LBufferedWithout   = "L04"
	LNilChan           = "L05"
	LSelectWith        = "L06"
	LSelectWithout     = "L07"
	LMutex             = "L08"
	LWaitGroup         = "L09"
	LCond              = "L10"

	// problems during recording
	RUnknownPanic ResultType = "R01"
	RTimeout      ResultType = "R02"

	// SNotExecutedWithPartner = "S00"
)

// Type to store a specific id in a select
//
// Parameter:
//   - ID (int): id of the involved channel
//   - ObjType (string): object type
//   - Index (int): internal index of the select int the case
type BugElementSelectCase struct {
	ID      int
	ObjType string
	Index   int
}

// GetBugElementSelectCase builds a BugElementSelectCase from a string
//
// Parameter:
//   - arg (string): the string representing the case
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
//   - Type (ResultType): The type of the bug
//   - TraceElement1 ([]analysis.TraceElement): first list of trace element involved in the bug
//     normally the elements that actually cause the bug, e.g. for send on close the send
//   - TraceElement2 ([]analysis.TraceElement): second list of trace element involved in the bug
//     normally the elements indirectly involved or elements to solve the bug (possible partner),
//     e.g. for send on close the close
type Bug struct {
	Type          ResultType
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
	case RUnknownPanic:
		typeStr = "Unknown Panic:"
		arg1Str = "Panic: "
	case RTimeout:
		typeStr = "Timeout"
	case ASendOnClosed:
		typeStr = "Actual Send on Closed Channel:"
		arg1Str = "send: "
		arg2Str = "close: "
	case ARecvOnClosed:
		typeStr = "Actual Receive on Closed Channel:"
		arg1Str = "recv: "
		arg2Str = "close: "
	case ACloseOnClosed:
		typeStr = "Actual Close on Closed Channel:"
		arg1Str = "close: "
		arg2Str = "close: "
	case ACloseOnNil:
		typeStr = "Actual close on nil channel:"
		arg1Str = "close: "
		arg2Str = "close: "
	case AConcurrentRecv:
		typeStr = "Concurrent Receive:"
		arg1Str = "recv: "
		arg2Str = "recv: "
	case ASelCaseWithoutPartner:
		typeStr = "Select Case without Partner:"
		arg1Str = "select: "
		arg2Str = "case: "
	case ANegWG:
		typeStr = "Actual negative Wait Group:"
		arg1Str = "done: "
	case AUnlockOfNotLockedMutex:
		typeStr = "Actual unlock of not locked mutex:"
		arg1Str = "unlock:"
	case PSendOnClosed:
		typeStr = "Possible send on closed channel:"
		arg1Str = "send: "
		arg2Str = "close: "
	case PRecvOnClosed:
		typeStr = "Possible receive on closed channel:"
		arg1Str = "recv: "
		arg2Str = "close: "
	case PNegWG:
		typeStr = "Possible negative waitgroup counter:"
		arg1Str = "done: "
		arg2Str = "add: "
	case PUnlockBeforeLock:
		typeStr = "Possible unlock of a not locked mutex:"
		arg1Str = "unlocks: "
		arg2Str = "locks: "
	case PCyclicDeadlock:
		typeStr = "Possible cyclic deadlock:"
		arg1Str = "head: "
		arg2Str = "tail: "
	case LWithoutBlock:
		typeStr = "Leak on routine without any blocking operation"
		arg1Str = "fork: "
	case LUnbufferedWith:
		typeStr = "Leak on unbuffered channel with possible partner:"
		arg1Str = "channel: "
		arg2Str = "partner: "
	case LUnbufferedWithout:
		typeStr = "Leak on unbuffered channel without possible partner:"
		arg1Str = "channel: "
	case LBufferedWith:
		typeStr = "Leak on buffered channel with possible partner:"
		arg1Str = "channel: "
		arg2Str = "partner: "
	case LBufferedWithout:
		typeStr = "Leak on buffered channel without possible partner:"
		arg1Str = "channel: "
	case LNilChan:
		typeStr = "Leak on nil channel:"
		arg1Str = "channel: "
	case LSelectWith:
		typeStr = "Leak on select with possible partner:"
		arg1Str = "select: "
		arg2Str = "partner: "
	case LSelectWithout:
		typeStr = "Leak on select without partner:"
		arg1Str = "select: "
	case LMutex:
		typeStr = "Leak on mutex:"
		arg1Str = "mutex: "
		arg2Str = "last: "
	case LWaitGroup:
		typeStr = "Leak on wait group:"
		arg1Str = "waitgroup: "
	case LCond:
		typeStr = "Leak on conditional variable:"
		arg1Str = "cond: "
	// case SNotExecutedWithPartner:
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

// Print the bug
func (b Bug) Println() {
	println(b.ToString())
}

// Process the bug that was selected from the analysis results
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
		bug.Type = RUnknownPanic
		actual = true
	case "R02":
		bug.Type = RTimeout
		actual = true
	case "A01":
		bug.Type = ASendOnClosed
		actual = true
	case "A02":
		bug.Type = ARecvOnClosed
		actual = true
	case "A03":
		bug.Type = ACloseOnClosed
		actual = true
	case "A04":
		bug.Type = ACloseOnNil
		actual = true
	case "A05":
		bug.Type = ANegWG
		actual = true
	case "A06":
		bug.Type = AUnlockOfNotLockedMutex
		actual = true
	case "A07":
		bug.Type = AConcurrentRecv
		actual = true
	case "A08":
		bug.Type = ASelCaseWithoutPartner
		actual = true
	case "P01":
		bug.Type = PSendOnClosed
	case "P02":
		bug.Type = PRecvOnClosed
	case "P03":
		bug.Type = PNegWG
	case "P04":
		bug.Type = PUnlockBeforeLock
	case "P05":
		bug.Type = PCyclicDeadlock
	// case "P06":
	// 	bug.Type = MixedDeadlock
	case "L00":
		containsArg1 = false
		bug.Type = LWithoutBlock
	case "L01":
		bug.Type = LUnbufferedWith
	case "L02":
		bug.Type = LUnbufferedWithout
		containsArg2 = false
	case "L03":
		bug.Type = LBufferedWith
	case "L04":
		bug.Type = LBufferedWithout
		containsArg2 = false
	case "L05":
		bug.Type = LNilChan
		containsArg2 = false
	case "L06":
		bug.Type = LSelectWith
	case "L07":
		bug.Type = LSelectWithout
		containsArg2 = false
	case "L08":
		bug.Type = LMutex
	case "L09":
		bug.Type = LWaitGroup
		containsArg2 = false
	case "L10":
		bug.Type = LCond
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
