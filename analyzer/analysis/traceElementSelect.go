// Copyrigth (c) 2024 Erik Kassubek
//
// File: traceElementSelect.go
// Brief: Struct and functions for select operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"analyzer/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

/*
 * TraceElementSelect is a trace element for a select statement
 * MARK: Struct
 * Fields:
 *   routine (int): The routine id
 *   tpre (int): The timestamp at the start of the event
 *   tpost (int): The timestamp at the end of the event
 *   id (int): The id of the select statement
 *   cases ([]traceElementSelectCase): The cases of the select statement, ordered by casi starting from 0
 *   chosenIndex (int): The internal index of chosen case
 *   containsDefault (bool): Whether the select statement contains a default case
 *   chosenCase (traceElementSelectCase): The chosen case, nil if default case chosen
 *   chosenDefault (bool): if the default case was chosen
 *   pos (string): The position of the select statement in the code
 *   posPartner ([]bool): For each case state, wether a possible partner exists
 *   casesWithPosPartner ([]int): Casis of cases with possible partner based on HB
 */
type TraceElementSelect struct {
	routine             int
	tPre                int
	tPost               int
	id                  int
	cases               []TraceElementChannel
	chosenCase          TraceElementChannel
	chosenIndex         int
	containsDefault     bool
	chosenDefault       bool
	pos                 string
	vc                  clock.VectorClock
	casesWithPosPartner []int
}

/*
 * Add a new select statement trace element
 * MARK: New
 * Args:
 *   routine (int): The routine id
 *   tPre (string): The timestamp at the start of the event
 *   tPost (string): The timestamp at the end of the event
 *   id (string): The id of the select statement
 *   cases (string): The cases of the select statement
 *   chosenIndex (string): The internal index of chosen case
 *   pos (string): The position of the select statement in the code
 */
func AddTraceElementSelect(routine int, tPre string,
	tPost string, id string, cases string, chosenIndex string, pos string) error {

	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	chosenIndexInt, err := strconv.Atoi(chosenIndex)
	if err != nil {
		return errors.New("chosenIndex is not an integer")
	}

	elem := TraceElementSelect{
		routine:             routine,
		tPre:                tPreInt,
		tPost:               tPostInt,
		id:                  idInt,
		chosenIndex:         chosenIndexInt,
		pos:                 pos,
		casesWithPosPartner: make([]int, 0),
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
			return errors.New("c_tpre is not an integer")
		}
		cTPost, err := strconv.Atoi(caseList[2])
		if err != nil {
			return errors.New("c_tpost is not an integer")
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
			pos:     pos,
		}

		casesList = append(casesList, elemCase)
		if elemCase.tPost != 0 {
			elem.chosenCase = elemCase
		}
	}

	elem.containsDefault = containsDefault
	elem.chosenDefault = chosenDefault
	elem.cases = casesList

	// check if partner was already processed, otherwise add to channelWithoutPartner
	if tPostInt != 0 {
		id := elem.chosenCase.id
		oID := elem.chosenCase.oID
		if _, ok := channelWithoutPartner[id][oID]; ok {
			elem.chosenCase.partner = channelWithoutPartner[id][oID]
			channelWithoutPartner[elem.chosenCase.id][oID].partner = &elem.chosenCase
			delete(channelWithoutPartner[id], oID)
		} else {
			if _, ok := channelWithoutPartner[id]; !ok {
				channelWithoutPartner[id] = make(map[int]*TraceElementChannel)
			}

			channelWithoutPartner[id][oID] = &elem.chosenCase
		}
	}

	return AddElementToTrace(&elem)
}

// MARK: Getter

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (se *TraceElementSelect) GetID() int {
	return se.id
}

/*
 * Get the cases of the select statement
 * Returns:
 *   []traceElementChannel: The cases of the select statement
 */
func (se *TraceElementSelect) GetCases() []TraceElementChannel {
	return se.cases
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (se *TraceElementSelect) GetRoutine() int {
	return se.routine
}

/*
 * Get the timestamp at the start of the event
 * Returns:
 *   int: The timestamp at the start of the event
 */
func (se *TraceElementSelect) GetTPre() int {
	return se.tPre
}

/*
 * Get the timestamp at the end of the event
 * Returns:
 *   int: The timestamp at the end of the event
 */
func (se *TraceElementSelect) GetTPost() int {
	return se.tPost
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   int: The timer of the element
 */
func (se *TraceElementSelect) GetTSort() int {
	if se.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return se.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (se *TraceElementSelect) GetPos() string {
	return se.pos
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (se *TraceElementSelect) GetTID() string {
	return se.pos + "@" + strconv.Itoa(se.tPre)
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (se *TraceElementSelect) GetVC() clock.VectorClock {
	return se.vc
}

/*
 * Get the chosen case
 * Return: the chosen case
 */
func (se *TraceElementSelect) GetChosenCase() *TraceElementChannel {
	if se.chosenDefault {
		return nil
	}
	return &se.chosenCase
}

/*
 * Get the index of the chosen case
 */
func (se *TraceElementSelect) GetChosenIndex() int {
	return se.chosenIndex
}

/*
 * Get whether the select contains a default case
 */
func (se *TraceElementSelect) GetContainsDefault() bool {
	return se.chosenDefault
}

/*
 * Get the communication partner of the select
 * Returns:
 *   *TraceElementChannel: The communication partner of the select or nil
 */
func (se *TraceElementSelect) GetPartner() *TraceElementChannel {
	if se.chosenCase.tPost != 0 && !se.chosenDefault {
		return se.chosenCase.partner
	}
	return nil
}

/*
 * Get the string representation of the object type
 */
func (se *TraceElementSelect) GetObjType() string {
	return "SS"
}

func (se *TraceElementSelect) GetCasiWithPosPartner() []int {
	return se.casesWithPosPartner
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (se *TraceElementSelect) SetT(time int) {
	se.tPre = time
	se.tPost = time

	se.chosenCase.tPost = time

	for _, c := range se.cases {
		c.tPre = time
	}
}

/*
 * Set the tpre of the element.
 * Args:
 *   tPre (int): The tpre of the element
 */
func (se *TraceElementSelect) SetTPre(tPre int) {
	se.tPre = tPre
	if se.tPost != 0 && se.tPost < tPre {
		se.tPost = tPre
	}

	for _, c := range se.cases {
		c.SetTPre2(tPre)
	}
}

/*
 * Set the tpre of the element. Do not update the chosen case
 * Args:
 *   tPre (int): The tpre of the element
 */
func (se *TraceElementSelect) SetTPre2(tPre int) {
	se.tPre = tPre
	if se.tPost != 0 && se.tPost < tPre {
		se.tPost = tPre
	}

	for _, c := range se.cases {
		c.SetTPre2(tPre)
	}
}

func (se *TraceElementSelect) SetChosenCase(index int) error {
	if index >= len(se.cases) {
		return fmt.Errorf("Invalid index %d for size %d", index, len(se.cases))
	}
	se.cases[se.chosenIndex].tPost = 0
	se.chosenIndex = index
	se.cases[index].tPost = se.tPost

	return nil
}

/*
 * Set tPost
 * Args:
 *   tSort (int): The timer of the element
 */
func (se *TraceElementSelect) SetTPost(tPost int) {
	se.tPost = tPost
	se.chosenCase.SetTPost2(tPost)
}

/*
 * Set tPost. Do not update the chosen case
 * Args:
 *   tSort (int): The timer of the element
 */
func (se *TraceElementSelect) SetTPost2(tPost int) {
	se.tPost = tPost
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (se *TraceElementSelect) SetTSort(tSort int) {
	se.SetTPre(tSort)
	se.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace. Do not update the chosen case
 * Args:
 *   tSort (int): The timer of the element
 */
func (se *TraceElementSelect) SetTSort2(tSort int) {
	se.SetTPre2(tSort)
	se.tPost = tSort
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 * tSort (int): The timer of the element
 */
func (se *TraceElementSelect) SetTWithoutNotExecuted(tSort int) {
	se.SetTPre(tSort)
	if se.tPost != 0 {
		se.tPost = tSort
	}
	se.chosenCase.SetTWithoutNotExecuted2(tSort)
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0. Do not update the chosen case
 * Args:
 * tSort (int): The timer of the element
 */
func (se *TraceElementSelect) SetTWithoutNotExecuted2(tSort int) {
	se.SetTPre2(tSort)
	if se.tPost != 0 {
		se.tPost = tSort
	}
}

func (se *TraceElementSelect) GetChosenDefault() bool {
	return se.chosenDefault
}

/*
 * Set the case where the channel id and direction is correct as the active one
 * Args:
 *     chanID int: id of the channel in the case, -1 for default
 *     send opChannel: channel operation of case
 * Returns:
 *     error
 */
func (se *TraceElementSelect) SetCase(chanID int, op OpChannel) error {
	if chanID == -1 {
		if se.containsDefault {
			se.chosenDefault = true
			se.chosenIndex = -1
			for i := range se.cases {
				se.cases[i].SetTPost(0)
			}
			return nil
		} else {
			return fmt.Errorf("Tried to set select without default to default")
		}
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

/*
 * Get the simple string representation of the element
 * MARK: ToString
 * Returns:
 *   string: The simple string representation of the element
 */
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
	res += "," + se.pos
	return res
}

/*
* Update and calculate the vector clock of the select element.
* MARK: VectorClock
 */
func (se *TraceElementSelect) updateVectorClock() {
	noChannel := se.chosenDefault || se.tPost == 0

	se.vc = currentVCHb[se.routine].Copy()

	if noChannel {
		currentVCHb[se.routine] = currentVCHb[se.routine].Inc(se.routine)
	} else {
		// update the vector clock
		se.chosenCase.vc = se.vc.Copy()
		se.chosenCase.updateVectorClock()
	}

	if analysisCases["selectWithoutPartner"] || modeIsFuzzing {
		CheckForSelectCaseWithoutPartnerSelect(se, currentVCHb[se.routine])
	}

	for _, c := range se.cases {
		c.vc = se.vc.Copy()
		if c.opC == SendOp {
			SetChannelAsLastSend(&c)
		} else if c.opC == RecvOp {
			SetChannelAsLastReceive(&c)
		}
	}

	if analysisCases["sendOnClosed"] {
		for i, c := range se.cases {
			if i == se.chosenIndex {
				continue
			}

			if _, ok := closeData[c.id]; ok {
				if c.opC == SendOp {
					foundSendOnClosedChannel(se.routine, c.id, se.pos, false)
				} else if c.opC == RecvOp {
					foundReceiveOnClosedChannel(&c, false)
				}
			}
		}
	}

	if analysisCases["leak"] {
		for _, c := range se.cases {
			CheckForLeakChannelRun(se.routine, c.id,
				elemWithVc{
					vc:   se.vc.Copy(),
					elem: se},
				int(c.opC), c.IsBuffered())
		}
	}
}

// MARK: Copy

/*
 * Copy the element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (se *TraceElementSelect) Copy() TraceElement {
	cases := make([]TraceElementChannel, 0)
	for _, c := range se.cases {
		cases = append(cases, *c.Copy().(*TraceElementChannel))
	}

	chosenCase := *se.chosenCase.Copy().(*TraceElementChannel)

	return &TraceElementSelect{
		routine:         se.routine,
		tPre:            se.tPre,
		tPost:           se.tPost,
		id:              se.id,
		cases:           cases,
		chosenCase:      chosenCase,
		chosenIndex:     se.chosenIndex,
		containsDefault: se.containsDefault,
		chosenDefault:   se.chosenDefault,
		pos:             se.pos,
		vc:              se.vc.Copy(),
	}
}
