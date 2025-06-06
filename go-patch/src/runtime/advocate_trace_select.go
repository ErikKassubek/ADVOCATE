// ADVOCATE-FILE_START

// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_trace_select.go
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
//   - cases []AdvocateTraceChannel: the operation for each of the non default cases
//     The elements are sorted the same as the internal sorting in the select,
//     first all send, then all recv
//   - selIndex int: The index of the operation in cases that was executed,
//     if default was executed, this is set to -1
//   - hasDef bool: true if the select has a default case, false otherwise
//   - file string: file where the operation occurred
//   - line int: line where the operation occurred
type AdvocateTraceSelect struct {
	tPre     int64
	tPost    int64
	id       uint64
	cases    []AdvocateTraceChannel
	selIndex int
	hasDef   bool
	file     string
	line     int
}

// AdvocateSelectPre adds a select to the trace
//
// Parameter:
//   - cases: cases of the select
//   - nsends: number of send cases
//   - ncases: total number of non default cases
//   - block: true if the select is blocking (has no default), false otherwise
//   - lockOrder: internal order of the locks
//
// Returns:
//   - index of the operation in the trace
func AdvocateSelectPre(cases *[]scase, nsends int, ncases int, block bool, lockorder []uint16) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	if cases == nil {
		return -1
	}

	id := GetAdvocateObjectID()
	caseElements := make([]AdvocateTraceChannel, 0)

	_, file, line, _ := Caller(3)
	if AdvocateIgnore(file) {
		return -1
	}

	i := 0

	maxCasi := 0
	caseElementMap := make(map[int]AdvocateTraceChannel)
	for _, casei := range lockorder {
		casi := int(casei)
		cas := (*cases)[casi]
		c := cas.c

		chanOp := OperationChannelRecv
		if casi < nsends {
			chanOp = OperationChannelSend
		}

		if c == nil { // ignore nil cases
			caseElementMap[casi] = AdvocateTraceChannel{
				tPre:  timer,
				op:    chanOp,
				isNil: true,
			}
		} else {
			i++

			caseElementMap[casi] = AdvocateTraceChannel{
				tPre:  timer,
				op:    chanOp,
				id:    c.id,
				qSize: c.dataqsiz,
			}
		}
		maxCasi = max(maxCasi, casi)
	}

	for i := 0; i < ncases; i++ {
		if _, ok := caseElementMap[i]; ok {
			caseElements = append(caseElements, caseElementMap[i])
		} else {
			chanOp := OperationChannelRecv
			if i < nsends {
				chanOp = OperationChannelSend
			}
			caseElements = append(caseElements, AdvocateTraceChannel{
				tPre:  timer,
				op:    chanOp,
				isNil: true,
			})
		}
	}

	elem := AdvocateTraceSelect{
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

// AdvocateSelectPost adds a post event for select in case of an non-default case
//
// Parameter:
//   - index: index of the operation in the trace
//   - c: channel of the chosen case
//   - selIndex: index of the chosen case in the select
//   - rClosed: true if the channel was closed at another routine
func AdvocateSelectPost(index int, c *hchan, selIndex int, rClosed bool) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceSelect)
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

// AdvocateSelectPreOneNonDef adds a new select element to the trace if the
// select has exactly one non-default case and a default case
//
// Parameter:
//   - c: channel of the non-default case
//   - send: true if the non-default case is a send, false otherwise
//
// Returns:
//   - index of the operation in the trace
func AdvocateSelectPreOneNonDef(c *hchan, send bool) int {
	if advocateTracingDisabled {
		return -1
	}

	timer := GetNextTimeStep()

	id := GetAdvocateObjectID()

	opChan := OperationChannelRecv
	if send {
		opChan = OperationChannelSend
	}

	var caseElem AdvocateTraceChannel

	if c != nil {
		if c.id == 0 {
			c.id = AdvocateChanMake(int(c.dataqsiz))
		}
		caseElem = AdvocateTraceChannel{
			tPre:  timer,
			id:    c.id,
			op:    opChan,
			qSize: c.dataqsiz,
		}
	} else {
		caseElem = AdvocateTraceChannel{
			tPre: timer,
			op:   opChan,
		}
	}

	_, file, line, _ := Caller(2)
	if AdvocateIgnore(file) {
		return -1
	}

	cases := make([]AdvocateTraceChannel, 1)
	cases[0] = caseElem

	elem := AdvocateTraceSelect{
		tPre:   timer,
		id:     id,
		cases:  cases,
		hasDef: true,
		file:   file,
		line:   line,
	}

	return insertIntoTrace(elem)
}

// AdvocateSelectPostOneNonDef adds the selected case for a select with one
// non-default and one default case
//
// Parameter:
//   - index: index of the operation in the trace
//   - res: true for channel, false for default
//   - c *hchan: the channel in the select cases
func AdvocateSelectPostOneNonDef(index int, res bool, c *hchan) {
	if advocateTracingDisabled {
		return
	}

	timer := GetNextTimeStep()

	if index == -1 {
		return
	}

	elem := currentGoRoutineInfo().getElement(index).(AdvocateTraceSelect)

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
//     The [case] is build using AdvocateTraceChannel.toStringForSelect()
func (elem AdvocateTraceSelect) toString() string {
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
func (elem AdvocateTraceSelect) getOperation() Operation {
	if elem.selIndex == -1 {
		return OperationSelectDefault
	}
	return OperationSelectCase
}
