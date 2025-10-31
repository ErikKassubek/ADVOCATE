// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/select.go
// Brief: Struct and functions for select operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ElementSelect is a trace element for a select statement
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
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementSelect struct {
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	cases                    []ElementChannel
	chosenCase               ElementChannel
	chosenIndex              int
	containsDefault          bool
	chosenDefault            bool
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	casesWithPosPartner      []int
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
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
func (this *Trace) AddTraceElementSelect(routine int, tPre string,
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

	elem := ElementSelect{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
		chosenIndex:              chosenIndexInt,
		file:                     file,
		line:                     line,
		casesWithPosPartner:      make([]int, 0),
		vc:                       nil,
		wVc:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
	}

	cs := strings.Split(cases, "~")
	casesList := make([]ElementChannel, 0)
	containsDefault := false
	chosenDefault := false
	for i, c := range cs {
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

		cID := -1
		if caseList[1] != "*" {
			cID, err = strconv.Atoi(caseList[1])
			if err != nil {
				return errors.New("c_id is not an integer")
			}
		}
		var cOpC = ChannelSend
		switch caseList[2] {
		case "R":
			cOpC = ChannelRecv
		case "C":
			return errors.New("Close in select case list")
		}

		cCl, err := strconv.ParseBool(caseList[3])
		if err != nil {
			return errors.New("c_cr is not a boolean")
		}

		cOID, err := strconv.Atoi(caseList[4])
		if err != nil {
			return errors.New("c_oId is not an integer")
		}
		cOSize, err := strconv.Atoi(caseList[5])
		if err != nil {
			return errors.New("c_oSize is not an integer")
		}

		cTPost := 0
		if i == chosenIndexInt {
			cTPost = tPostInt
		}

		elemCase := ElementChannel{
			routine:  routine,
			tPre:     tPreInt,
			tPost:    cTPost,
			id:       cID,
			op:       cOpC,
			cl:       cCl,
			oID:      cOID,
			qSize:    cOSize,
			sel:      &elem,
			selIndex: len(caseList),
			file:     file,
			line:     line,
		}

		casesList = append(casesList, elemCase)
		if elemCase.tPost != 0 {
			elem.chosenCase = elemCase
			elemCase.findPartner(this)
		}
	}

	elem.containsDefault = containsDefault
	elem.chosenDefault = chosenDefault
	elem.cases = casesList

	this.AddElement(&elem)

	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementSelect) GetID() int {
	return this.id
}

// GetCases returns the cases of the select statement
//
// Returns:
//   - []traceElementChannel: The cases of the select statement
func (this *ElementSelect) GetCases() []ElementChannel {
	return this.cases
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementSelect) GetRoutine() int {
	return this.routine
}

// GetTPre returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the start of the event
func (this *ElementSelect) GetTPre() int {
	return this.tPre
}

// GetTPost returns the timestamp at the start of the event
//
// Returns:
//   - int: The timestamp at the end of the event
func (this *ElementSelect) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementSelect) GetTSort() int {
	if this.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return this.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementSelect) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the operations
//
// Returns:
//   - string: The replay id of the element
func (this *ElementSelect) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - string: The file of the element
func (this *ElementSelect) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - string: The line of the element
func (this *ElementSelect) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementSelect) GetTID() string {
	return "S@" + this.GetPos() + "@" + strconv.Itoa(this.tPre)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementSelect) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementSelect) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementSelect) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementSelect) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetChosenCase returns the chosen case
//
// Returns:
//   - the chosen case
func (this *ElementSelect) GetChosenCase() *ElementChannel {
	if this.chosenDefault || this.tPost == 0 {
		return nil
	}
	return &this.chosenCase
}

// GetChosenIndex returns the index of the chosen case in se.cases
//
// Returns:
//   - The internal index of the chosen case
func (this *ElementSelect) GetChosenIndex() int {
	return this.chosenIndex
}

// GetContainsDefault returns whether the select contains a default case
//
// Returns:
//   - bool: true if select contains default, false otherwise
func (this *ElementSelect) GetContainsDefault() bool {
	return this.chosenDefault
}

// GetPartner returns the communication partner of the select. If there is none,
// it returns nil
//
// Returns:
//   - *TraceElementChannel: The communication partner of the select or nil
func (this *ElementSelect) GetPartner() *ElementChannel {
	if this.chosenCase.tPost != 0 && !this.chosenDefault {
		return this.chosenCase.partner
	}
	return nil
}

// GetType returns he object type
//
// Parameter:
//   - operations bool: if true, the operation id contains the operations, otherwise just that it is select
//
// Returns:
//   - the object type
func (this *ElementSelect) GetType(operation bool) ObjectType {
	if !operation {
		return Select
	}

	return SelectOp
}

// GetCasiWithPosPartner returns a list of all internal indices, where the
// corresponding case as a potential partner
//
// Returns:
//   - []int: list of indices
func (this *ElementSelect) GetCasiWithPosPartner() []int {
	return this.casesWithPosPartner
}

// IsEqual checks if the given element is equal to the select
//
// Parameter:
//   - elem TraceElement: The element
//
// Returns:
//   - bool: true if they are equal, false otherwise
func (this *ElementSelect) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementSelect) IsSameElement(elem Element) bool {
	return false
}

// GetTraceIndex returns the index of the element in the routine
// Returns
//
//   - int: routine index
//   - int: routine local index of the element
func (this *ElementSelect) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementSelect) SetT(time int) {
	this.tPre = time
	this.tPost = time

	this.chosenCase.tPost = time

	for i := range this.cases {
		this.cases[i].tPre = time
	}
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementSelect) SetTPre(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}

	for _, c := range this.cases {
		c.SetTPre2(tPre)
	}
}

// SetTPre2 sets the tPre of the element. It does not update the chosen case
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementSelect) SetTPre2(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}

	for _, c := range this.cases {
		c.SetTPre2(tPre)
	}
}

// AddCasesWithPosPartner adds an casi to casesWithPosPartner
//
// Parameter:
//   - casi int: the case id to add
func (this *ElementSelect) AddCasesWithPosPartner(casi int) {
	this.casesWithPosPartner = append(this.casesWithPosPartner, casi)
}

// GetCasesWithPosPartner returns casesWithPosPartner
//
// Returns:
//   - []int: list of cases with potential partner
func (this *ElementSelect) GetCasesWithPosPartner() []int {
	return this.casesWithPosPartner
}

// SetChosenCase sets the chosen case of a select
//
// Parameter:
//   - index of the case that should be set as the chosen case
//
// Returns:
//   - error
func (this *ElementSelect) SetChosenCase(index int) error {
	if index >= len(this.cases) {
		return fmt.Errorf("Invalid index %d for size %d", index, len(this.cases))
	}
	this.cases[this.chosenIndex].tPost = 0
	this.chosenIndex = index
	this.cases[index].tPost = this.tPost

	return nil
}

// SetTPost sets the tPost
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementSelect) SetTPost(tPost int) {
	this.tPost = tPost
	this.chosenCase.SetTPost2(tPost)
}

// SetTPost2 sets the tPost. It does not update the chosen case
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementSelect) SetTPost2(tPost int) {
	this.tPost = tPost
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementSelect) SetTSort(tSort int) {
	this.SetTPre(tSort)
	this.tPost = tSort
}

// SetTSort2 set the timer, that is used for the sorting of the trace.
// It does not update the chosen case
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementSelect) SetTSort2(tSort int) {
	this.SetTPre2(tSort)
	this.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter: tSort int: The timer of the element
func (this *ElementSelect) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
	this.chosenCase.SetTWithoutNotExecuted2(tSort)
}

// SetTWithoutNotExecuted2 sets the timer, that is used for the sorting of the trace, only if the original
// value was not 0. Do not update the chosen case
//
// Parameter: tSort int: The timer of the element
func (this *ElementSelect) SetTWithoutNotExecuted2(tSort int) {
	this.SetTPre2(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// GetChosenDefault if the default case is the executed case
//
// Returns: bool: true if default case
func (this *ElementSelect) GetChosenDefault() bool {
	return this.chosenDefault
}

// SetCaseByIndex set the case to the case at the given index or default if index = -1
//
// Parameter:
//   - index of the case, -1 for default
//
// Returns:
//   - error
func (this *ElementSelect) SetCaseByIndex(index int) error {
	if index > len(this.cases) {
		return fmt.Errorf("Invalid index for select: %d [%d]", index, len(this.cases))
	}

	for i := range this.cases {
		this.cases[i].SetTPost(0)
	}

	if index < 0 {
		this.chosenDefault = true
		this.chosenIndex = -1
		return nil
	}

	this.cases[index].SetTPost(this.GetTPost())
	this.chosenIndex = index
	this.chosenDefault = false
	return nil
}

// SetCase set the case where the channel id and direction is correct as the active one
//
// Parameter:
//   - chanID int: id of the channel in the case, -1 for default
//   - send opChannel: channel operation of case
//
// Returns:
//   - error
func (this *ElementSelect) SetCase(chanID int, op ObjectType) error {
	if chanID == -1 {
		if this.containsDefault {
			this.chosenDefault = true
			this.chosenIndex = -1
			for i := range this.cases {
				this.cases[i].SetTPost(0)
			}
			return nil
		}

		return fmt.Errorf("Tried to set select without default to default")
	}

	found := false
	for i, c := range this.cases {
		if c.id == chanID && c.op == op {
			tPost := this.GetTPost()
			if !this.chosenDefault {
				this.cases[this.chosenIndex].SetTPost(0)
			} else {
				this.chosenDefault = false
			}
			this.cases[i].SetTPost(tPost)
			this.chosenIndex = i
			this.chosenDefault = false
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
func (this *ElementSelect) ToString() string {
	res := "S" + "," + strconv.Itoa(this.tPre) + "," +
		strconv.Itoa(this.tPost) + "," + strconv.Itoa(this.id) + ","

	notNil := 0
	for _, ca := range this.cases { // cases
		if ca.tPre != 0 { // ignore nil cases
			if notNil != 0 {
				res += "~"
			}
			res += ca.toStringSep(".", true)
			notNil++
		}
	}

	if this.containsDefault {
		if notNil != 0 {
			res += "~"
		}
		if this.chosenDefault {
			res += "D"
		} else {
			res += "d"
		}
	}
	res += "," + strconv.Itoa(this.chosenIndex)
	res += "," + this.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementSelect) GetTraceID() int {
	return this.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementSelect) setTraceID(ID int) {
	this.traceID = ID
}

// Copy the element
//
// Parameter:
//   - mapping map[string]Element: map containing all already copied elements.
//     This avoids double copy of referenced elements
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementSelect) Copy(mapping map[string]Element) Element {
	tID := this.GetTID()

	if existing, ok := mapping[tID]; ok {
		return existing
	}

	elem := &ElementSelect{
		traceID:                  this.traceID,
		index:                    this.index,
		routine:                  this.routine,
		tPre:                     this.tPre,
		tPost:                    this.tPost,
		id:                       this.id,
		chosenIndex:              this.chosenIndex,
		containsDefault:          this.containsDefault,
		chosenDefault:            this.chosenDefault,
		file:                     this.file,
		line:                     this.line,
		vc:                       this.vc.Copy(),
		wVc:                      this.wVc.Copy(),
		numberConcurrent:         this.numberConcurrent,
		numberConcurrentWeak:     this.numberConcurrentWeak,
		numberConcurrentSame:     this.numberConcurrentSame,
		numberConcurrentWeakSame: this.numberConcurrentWeakSame,
	}

	mapping[tID] = elem

	elem.cases = make([]ElementChannel, 0)
	for _, c := range this.cases {
		elem.cases = append(elem.cases, *c.Copy(mapping).(*ElementChannel))
	}

	elem.chosenCase = *this.chosenCase.Copy(mapping).(*ElementChannel)

	for _, c := range elem.cases {
		c.sel = elem
	}

	return elem
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Parameter:
//   - weak bool: get number of weak concurrent
//   - sameElem bool: only operation on the same variable
//
// Returns:
//   - number of concurrent element, or -1
func (this *ElementSelect) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return this.numberConcurrentWeakSame
		}
		return this.numberConcurrentWeak
	}
	if sameElem {
		return this.numberConcurrentSame
	}
	return this.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (this *ElementSelect) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			this.numberConcurrentWeakSame = c
		} else {
			this.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			this.numberConcurrentSame = c
		} else {
			this.numberConcurrent = c
		}
	}

	if this.GetChosenCase() != nil {
		this.GetChosenCase().SetNumberConcurrent(c, weak, sameElem)
	}
}

// HasCommonChannels returns if the set of cases that are in both the receiver
// select and the argument select is not empty. We do not consider the default case
//
// Parameter:
//   - s *trace.ElementSelect: the other select
//
// Returns:
//   - bool: true if this and s have at least one common channel
func (this *ElementSelect) HasCommonChannel(s *ElementSelect) bool {
	for _, c := range s.GetCases() {
		if this.IsInCases(&c) {
			return true
		}
	}

	return false
}

// GetCommonChannels returns the set of cases that are in both the receiver
// select and the argument select. The result does not contain the default case
//
// Parameter:
//   - s *trace.ElementSelect: the other select
//
// Returns:
//   - []trace.ElementChannel: the set of channels in both the receiver and the argument
func (this *ElementSelect) GetCommonChannel(s *ElementSelect) []ElementChannel {
	res := make([]ElementChannel, 0)
	for _, c := range s.GetCases() {
		if this.IsInCases(&c) {
			res = append(res, c)
		}
	}

	return res
}

// IsInCases returns true, if the given channel is in the cases. We only care
// about the channel, not the same operations
//
// Parameter:
//   - ch *trace.ElementChannel: the channel element
//
// Returns:
//   - bool: true if the channel of the operations in ch is in a case in the select
func (this *ElementSelect) IsInCases(ch *ElementChannel) bool {
	for _, c := range this.GetCases() {
		if c.IsSameElement(ch) {
			return true
		}
	}

	return false
}
