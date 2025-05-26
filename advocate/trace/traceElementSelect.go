// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementSelect.go
// Brief: Struct and functions for select operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// TraceElementSelect is a trace element for a select statement
// Fields:
//   - traceID: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the select statement
//   - cases []traceElementSelectCase: The cases of the select statement, ordered by casi starting from 0
//   - chosenIndex int: The internal index of chosen case
//   - containsDefault bool: Whether the select statement contains a default case
//   - chosenCase traceElementSelectCase: The chosen case, nil if default case chosen
//   - chosenDefault bool: if the default case was chosen
//   - file string: The file of the select statement in the code
//   - line int: The line of the select statement in the code
//   - posPartner []bool: For each case state, wether a possible partner exists
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - casesWithPosPartner []int: casi of cases with possible partner based on HB
type TraceElementSelect struct {
	traceID             int
	index               int
	routine             int
	tPre                int
	tPost               int
	id                  int
	cases               []TraceElementChannel
	chosenCase          TraceElementChannel
	chosenIndex         int
	containsDefault     bool
	chosenDefault       bool
	file                string
	line                int
	vc                  *clock.VectorClock
	wVc                 *clock.VectorClock
	casesWithPosPartner []int
}

// AddTraceElementSelect adds a new select statement element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the select statement
//   - cases string: The cases of the select statement
//   - chosenIndex string: The internal index of chosen case
//   - pos string: The position of the select statement in the code
func (t *Trace) AddTraceElementSelect(routine int, tPre string,
	tPost string, id string, cases string, chosenIndex string, pos string) error {

	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	chosenIndexInt, err := strconv.Atoi(chosenIndex)
	if err != nil {
		return errors.New("chosenIndex is not an integer")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementSelect{
		index:               t.numberElemsInTrace[routine],
		routine:             routine,
		tPre:                tPreInt,
		tPost:               tPostInt,
		id:                  idInt,
		chosenIndex:         chosenIndexInt,
		file:                file,
		line:                line,
		casesWithPosPartner: make([]int, 0),
		vc:                  nil,
		wVc:                 nil,
	}

	cs := strings.Split(cases, "~")
	casesList := make([]TraceElementChannel, 0)
	containsDefault := false
	chosenDefault := false
	for _, c := range cs {
		if c == "" {
			continue
		}

		if c == "d" {
			containsDefault = true
			break
		}
		if c == "D" {
			containsDefault = true
			chosenDefault = true
			break
		}

		// read channel operation
		caseList := strings.Split(c, ".")
		cTPre, err := strconv.Atoi(caseList[1])
		if err != nil {
			return errors.New("c_tPre is not an integer")
		}
		cTPost, err := strconv.Atoi(caseList[2])
		if err != nil {
			return errors.New("c_tPost is not an integer")
		}

		cID := -1
		if caseList[3] != "*" {
			cID, err = strconv.Atoi(caseList[3])
			if err != nil {
				return errors.New("c_id is not an integer")
			}
		}
		var cOpC = SendOp
		if caseList[4] == "R" {
			cOpC = RecvOp
		} else if caseList[4] == "C" {
			return errors.New("Close in select case list")
		}

		cCl, err := strconv.ParseBool(caseList[5])
		if err != nil {
			return errors.New("c_cr is not a boolean")
		}

		cOID, err := strconv.Atoi(caseList[6])
		if err != nil {
			return errors.New("c_oId is not an integer")
		}
		cOSize, err := strconv.Atoi(caseList[7])
		if err != nil {
			return errors.New("c_oSize is not an integer")
		}

		elemCase := TraceElementChannel{
			routine: routine,
			tPre:    cTPre,
			tPost:   cTPost,
			id:      cID,
			opC:     cOpC,
			cl:      cCl,
			oID:     cOID,
			qSize:   cOSize,
			sel:     &elem,
			file:    file,
			line:    line,
		}

		casesList = append(casesList, elemCase)
		if elemCase.tPost != 0 {
			elem.chosenCase = elemCase
			elemCase.findPartner(t)
		}
	}

	elem.containsDefault = containsDefault
	elem.chosenDefault = chosenDefault
	elem.cases = casesList

	t.AddElement(&elem)

	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (se *TraceElementSelect) GetID() int {
	return se.id
}

// GetCases returns the cases of the select statement
//
// Returns:
//   - []traceElementChannel: The cases of the select statement
func (se *TraceElementSelect) GetCases() []TraceElementChannel {
	return se.cases
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (se *TraceElementSelect) GetRoutine() int {
	return se.routine
}

// GetTPre returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the start of the event
func (se *TraceElementSelect) GetTPre() int {
	return se.tPre
}

// GetTPost returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the end of the event
func (se *TraceElementSelect) GetTPost() int {
	return se.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (se *TraceElementSelect) GetTSort() int {
	if se.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return se.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (se *TraceElementSelect) GetPos() string {
	return fmt.Sprintf("%s:%d", se.file, se.line)
}

// GetReplayID returns the replay id of the operations
//
// Returns:
//   - string: The replay id of the element
func (se *TraceElementSelect) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", se.routine, se.file, se.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - string: The file of the element
func (se *TraceElementSelect) GetFile() string {
	return se.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - string: The line of the element
func (se *TraceElementSelect) GetLine() int {
	return se.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (se *TraceElementSelect) GetTID() string {
	return se.GetPos() + "@" + strconv.Itoa(se.tPre)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (se *TraceElementSelect) SetVc(vc *clock.VectorClock) {
	se.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (se *TraceElementSelect) SetWVc(vc *clock.VectorClock) {
	se.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (se *TraceElementSelect) GetVC() *clock.VectorClock {
	return se.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (se *TraceElementSelect) GetWVc() *clock.VectorClock {
	return se.wVc
}

// GetChosenCase returns the chosen case
//
// Returns:
//   - the chosen case
func (se *TraceElementSelect) GetChosenCase() *TraceElementChannel {
	if se.chosenDefault {
		return nil
	}
	return &se.chosenCase
}

// GetChosenIndex returns the index of the chosen case in se.cases
//
// Returns:
//   - The internal index of the chosen case
func (se *TraceElementSelect) GetChosenIndex() int {
	return se.chosenIndex
}

// GetContainsDefault returns whether the select contains a default case
//
// Returns:
//   - bool: true if select contains default, false otherwise
func (se *TraceElementSelect) GetContainsDefault() bool {
	return se.chosenDefault
}

// GetPartner returns the communication partner of the select. If there is none,
// it returns nil
//
// Returns:
//   - *TraceElementChannel: The communication partner of the select or nil
func (se *TraceElementSelect) GetPartner() *TraceElementChannel {
	if se.chosenCase.tPost != 0 && !se.chosenDefault {
		return se.chosenCase.partner
	}
	return nil
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operations bool: if true, the operation id contains the operations, otherwise just that it is select
//
// Returns:
//   - the object type
func (se *TraceElementSelect) GetObjType(operation bool) string {
	if operation {
		return ObjectTypeSelect + "S"
	}

	return ObjectTypeSelect
}

// GetCasiWithPosPartner returns a list of all internal indices, where the
// corresponding case as a potential partner
//
// Returns:
//   - []int: list of indices
func (se *TraceElementSelect) GetCasiWithPosPartner() []int {
	return se.casesWithPosPartner
}

// IsEqual checks if the given element is equal to the select
//
// Parameter:
//   - elem TraceElement: The element
//
// Returns:
//   - bool: true if they are equal, false otherwise
func (se *TraceElementSelect) IsEqual(elem TraceElement) bool {
	return se.routine == elem.GetRoutine() && se.ToString() == elem.ToString()
}

// GetTraceIndex returns the index of the element in the routine
// Returns
//
//   - int: routine index
//   - int: routine local index of the element
func (se *TraceElementSelect) GetTraceIndex() (int, int) {
	return se.routine, se.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (se *TraceElementSelect) SetT(time int) {
	se.tPre = time
	se.tPost = time

	se.chosenCase.tPost = time

	for _, c := range se.cases {
		c.tPre = time
	}
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (se *TraceElementSelect) SetTPre(tPre int) {
	se.tPre = tPre
	if se.tPost != 0 && se.tPost < tPre {
		se.tPost = tPre
	}

	for _, c := range se.cases {
		c.SetTPre2(tPre)
	}
}

// SetTPre2 sets the tPre of the element. It does not update the chosen case
//
// Parameter:
//   - tPre int: The tPre of the element
func (se *TraceElementSelect) SetTPre2(tPre int) {
	se.tPre = tPre
	if se.tPost != 0 && se.tPost < tPre {
		se.tPost = tPre
	}

	for _, c := range se.cases {
		c.SetTPre2(tPre)
	}
}

// AddCasesWithPosPartner adds an casi to casesWithPosPartner
//
// Parameter:
//   - casi int: the case id to add
func (se *TraceElementSelect) AddCasesWithPosPartner(casi int) {
	se.casesWithPosPartner = append(se.casesWithPosPartner, casi)
}

// GetCasesWithPosPartner returns casesWithPosPartner
//
// Returns:
//   - []int: list of cases with potential partner
func (se *TraceElementSelect) GetCasesWithPosPartner() []int {
	return se.casesWithPosPartner
}

// SetChosenCase sets the chosen case of a select
//
// Parameter:
//   - index of the case that should be set as the chosen case
//
// Returns:
//   - error
func (se *TraceElementSelect) SetChosenCase(index int) error {
	if index >= len(se.cases) {
		return fmt.Errorf("Invalid index %d for size %d", index, len(se.cases))
	}
	se.cases[se.chosenIndex].tPost = 0
	se.chosenIndex = index
	se.cases[index].tPost = se.tPost

	return nil
}

// SetTPost sets the tPost
//
// Parameter:
//   - tSort int: The timer of the element
func (se *TraceElementSelect) SetTPost(tPost int) {
	se.tPost = tPost
	se.chosenCase.SetTPost2(tPost)
}

// SetTPost2 sets the tPost. It does not update the chosen case
//
// Parameter:
//   - tSort int: The timer of the element
func (se *TraceElementSelect) SetTPost2(tPost int) {
	se.tPost = tPost
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (se *TraceElementSelect) SetTSort(tSort int) {
	se.SetTPre(tSort)
	se.tPost = tSort
}

// SetTSort2 set the timer, that is used for the sorting of the trace.
// It does not update the chosen case
//
// Parameter:
//   - tSort int: The timer of the element
func (se *TraceElementSelect) SetTSort2(tSort int) {
	se.SetTPre2(tSort)
	se.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter: tSort int: The timer of the element
func (se *TraceElementSelect) SetTWithoutNotExecuted(tSort int) {
	se.SetTPre(tSort)
	if se.tPost != 0 {
		se.tPost = tSort
	}
	se.chosenCase.SetTWithoutNotExecuted2(tSort)
}

// SetTWithoutNotExecuted2 sets the timer, that is used for the sorting of the trace, only if the original
// value was not 0. Do not update the chosen case
//
// Parameter: tSort int: The timer of the element
func (se *TraceElementSelect) SetTWithoutNotExecuted2(tSort int) {
	se.SetTPre2(tSort)
	if se.tPost != 0 {
		se.tPost = tSort
	}
}

// GetChosenDefault if the default case is the executed case
//
// Returns: bool: true if default case
func (se *TraceElementSelect) GetChosenDefault() bool {
	return se.chosenDefault
}

// SetCase set the case where the channel id and direction is correct as the active one
//
// Parameter:
//   - chanID int: id of the channel in the case, -1 for default
//   - send opChannel: channel operation of case
//
// Returns:
//   - error
func (se *TraceElementSelect) SetCase(chanID int, op OpChannel) error {
	if chanID == -1 {
		if se.containsDefault {
			se.chosenDefault = true
			se.chosenIndex = -1
			for i := range se.cases {
				se.cases[i].SetTPost(0)
			}
			return nil
		}

		return fmt.Errorf("Tried to set select without default to default")
	}

	found := false
	for i, c := range se.cases {
		if c.id == chanID && c.opC == op {
			tPost := se.GetTPost()
			if !se.chosenDefault {
				se.cases[se.chosenIndex].SetTPost(0)
			} else {
				se.chosenDefault = false
			}
			se.cases[i].SetTPost(tPost)
			se.chosenIndex = i
			se.chosenDefault = false
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Select case not found")
	}

	return nil
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (se *TraceElementSelect) ToString() string {
	res := "S" + "," + strconv.Itoa(se.tPre) + "," +
		strconv.Itoa(se.tPost) + "," + strconv.Itoa(se.id) + ","

	notNil := 0
	for _, ca := range se.cases { // cases
		if ca.tPre != 0 { // ignore nil cases
			if notNil != 0 {
				res += "~"
			}
			res += ca.toStringSep(".", false)
			notNil++
		}
	}

	if se.containsDefault {
		if notNil != 0 {
			res += "~"
		}
		if se.chosenDefault {
			res += "D"
		} else {
			res += "d"
		}
	}
	res += "," + strconv.Itoa(se.chosenIndex)
	res += "," + se.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (se *TraceElementSelect) GetTraceID() int {
	return se.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (se *TraceElementSelect) setTraceID(ID int) {
	se.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (se *TraceElementSelect) Copy() TraceElement {
	cases := make([]TraceElementChannel, 0)
	for _, c := range se.cases {
		cases = append(cases, *c.Copy().(*TraceElementChannel))
	}

	chosenCase := *se.chosenCase.Copy().(*TraceElementChannel)

	return &TraceElementSelect{
		traceID:         se.traceID,
		index:           se.index,
		routine:         se.routine,
		tPre:            se.tPre,
		tPost:           se.tPost,
		id:              se.id,
		cases:           cases,
		chosenCase:      chosenCase,
		chosenIndex:     se.chosenIndex,
		containsDefault: se.containsDefault,
		chosenDefault:   se.chosenDefault,
		file:            se.file,
		line:            se.line,
		vc:              se.vc.Copy(),
		wVc:             se.wVc.Copy(),
	}
}
