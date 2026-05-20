// GOCDR-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: gocdr_trace_select.go
// Brief: Functionality for selects
//
// Author: Erik Kassubek
// Created: 2024-02-16
//
// License: BSD-3-Clause

package runtime

// Struct to store a spawn
//
// Fields
//   - tPre int64: time when the operation started
//   - tPost int64: time when the operation finished
//   - id uint64: id of the select
//   - cases []GocdrTraceChannel: the operation for each of the non default cases
//     The elements are sorted the same as the internal sorting in the select,
//     first all send, then all recv
//   - selIndex int: The index of the operation in cases that was executed,
//     if default was executed, this is set to -1
//   - hasDef bool: true if the select has a default case, false otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type GocdrTraceSelect struct {
	tPre     int64
	tPost    int64
	id       uint64
	cases    []GocdrTraceChannel
	selIndex int
	hasDef   bool
	file     string
	line     int
}

// GocdrSelectPre adds a select to the trace
//
// Parameter:
//   - cases: cases of the select
//   - nsends: number of send cases
//   - ncases: total number of non default cases
//   - block: true if the select is blocking (has no default), false otherwise
//
// Returns:
//   - index of the operation in the trace
func GocdrSelectPre(cases *[]scase, nsends int, ncases int, block bool) int {
	if gocdrTracingDisabled || cases == nil {
		return -1
	}

	timer := GetNextTimeStep()

	_, file, line, _ := Caller(CallerSkipSelect)
	if GocdrIgnore(file) {
		return -1
	}

	id := GetGocdrObjectID()
	caseElements := make([]GocdrTraceChannel, ncases)

	for casi := 0; casi < ncases; casi++ {
		cas := (*cases)[casi]
		c := cas.c

		chanOp := OperationChannelRecv
		if casi < nsends {
			chanOp = OperationChannelSend
		}

		if c == nil { // ignore nil cases
			caseElements[casi] = GocdrTraceChannel{
				tPre:  timer,
				op:    chanOp,
				isNil: true,
			}
		} else {
			caseElements[casi] = GocdrTraceChannel{
				tPre:  timer,
				op:    chanOp,
				id:    c.id,
				qSize: c.dataqsiz,
			}
		}
	}

	elem := GocdrTraceSelect{
		tPre:  timer,
		id:    id,
		cases: caseElements,
		file:  file,
		line:  line,
	}

	if !block {
		elem.hasDef = true
	}

	return insertIntoTrace(elem)
}

// GocdrSelectPost adds a post event for select in case of an non-default case
//
// Parameter:
//   - index: index of the operation in the trace
//   - c: channel of the chosen case
//   - selIndex: index of the chosen case in the select
//   - rClosed: true if the channel was closed at another routine
func GocdrSelectPost(index int, c *hchan, selIndex int, rClosed bool) {
	if gocdrTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(GocdrTraceSelect)
	elem.tPost = timer
	elem.selIndex = selIndex

	if selIndex != -1 { // not default case
		// set tpost and cl of chosen case
		chosenCase := elem.cases[selIndex]
		chosenCase.tPost = timer
		if rClosed {
			chosenCase.cl = true
		}

		// set oId
		if chosenCase.op == OperationChannelSend {
			chosenCase.oId = c.numberSend
			c.numberSend++
		} else {
			chosenCase.oId = c.numberRecv
			c.numberRecv++
		}
		chosenCase.qCount = uint(c.numberSend - c.numberRecv)

		elem.cases[selIndex] = chosenCase
	}

	currentGoRoutineInfo().updateElement(index, elem)
}

// GocdrSelectPreOneNonDef adds a new select element to the trace if the
// select has exactly one non-default case and a default case
//
// Parameter:
//   - c: channel of the non-default case
//   - send: true if the non-default case is a send, false otherwise
//
// Returns:
//   - index of the operation in the trace
func GocdrSelectPreOneNonDef(c *hchan, send bool) int {
	if gocdrTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	id := GetGocdrObjectID()

	opChan := OperationChannelRecv
	if send {
		opChan = OperationChannelSend
	}

	var caseElem GocdrTraceChannel

	if c != nil {
		if c.id == 0 {
			c.id = GocdrChanMake(int(c.dataqsiz))
		}
		caseElem = GocdrTraceChannel{
			tPre:  timer,
			id:    c.id,
			op:    opChan,
			qSize: c.dataqsiz,
		}
	} else {
		caseElem = GocdrTraceChannel{
			tPre: timer,
			op:   opChan,
		}
	}

	_, file, line, _ := Caller(CallerSkipSelectOneDef)
	if GocdrIgnore(file) {
		return -1
	}

	cases := make([]GocdrTraceChannel, 1)
	cases[0] = caseElem

	elem := GocdrTraceSelect{
		tPre:   timer,
		id:     id,
		cases:  cases,
		hasDef: true,
		file:   file,
		line:   line,
	}

	return insertIntoTrace(elem)
}

// GocdrSelectPostOneNonDef adds the selected case for a select with one
// non-default and one default case
//
// Parameter:
//   - index: index of the operation in the trace
//   - res: true for channel, false for default
//   - c *hchan: the channel in the select cases
func GocdrSelectPostOneNonDef(index int, res bool, c *hchan) {
	if gocdrTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(GocdrTraceSelect)

	elem.tPost = timer

	if res { // channel case
		ca := elem.cases[0]
		ca.tPost = timer
		if ca.op == OperationChannelSend {
			c.numberSend++
		} else {
			c.numberRecv++
		}
		ca.qCount = uint(c.numberSend - c.numberRecv)
		elem.cases[0] = ca
		elem.selIndex = 0
	} else { // default case
		elem.selIndex = -1
	}

	currentGoRoutineInfo().updateElement(index, elem)
}

// Get a string representation of the trace element
//
// Returns:
//   - string: the string representation of the form
//     [S],[tPre],[tPost],[id],[cases],[selIndex],[file],[line]
//     where cases consists of the form [case]~[case]~..., followed by a d
//     if the select has a default that was not executed, or D if it was executed.
//     The [case] is build using GocdrTraceChannel.toStringForSelect()
func (elem GocdrTraceSelect) toString() string {
	p1 := buildTraceElemString("S", elem.tPre, elem.tPost, elem.id)
	p2 := buildTraceElemString(elem.selIndex, posToString(elem.file, elem.line))
	cases := ""
	for i, c := range elem.cases {
		if i != 0 {
			cases += "~"
		}
		cases += c.toStringForSelect()
	}
	if elem.hasDef {
		if cases != "" {
			cases += "~"
		}
		if elem.selIndex == -1 {
			cases += "D"
		} else {
			cases += "d"
		}
	}

	return buildTraceElemString(p1, cases, p2)
}

// getOperation is a getter for the operation
//
// Returns:
//   - Operation: the operation
func (elem GocdrTraceSelect) getOperation() Operation {
	if elem.selIndex == -1 {
		return OperationSelectDefault
	}
	return OperationSelectCase
}
