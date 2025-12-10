// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementChannel.go
// Brief: Struct and functions for channel operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"advocate/analysis/hb/clock"
)

// ElementChannel is a trace element for a channel
//
// Fields:
//   - id: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - objId int: The id of the channel
//   - op ObjectType: The operation on the channel
//   - cl bool: Whether the channel has closed
//   - oID int: The id of the other communication
//   - qSize int: The size of the channel queue
//   - qCount int: The number of elements in the queue after the operation
//   - file string: The file of the channel operation in the code
//   - line int: The line of the channel operation
//   - sel *traceElementSelect: The select operation, if the channel operation
//     is part of a select, otherwise nil
//   - selIndex int: index of the channel in sel.chases if sel != nil, otherwise -1
//   - partner *ElementChannel: The partner of the channel operation
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - numberConcurrent int: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak int: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementChannel struct {
	id                       int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	objId                    int
	op                       OperationType
	cl                       bool
	oID                      int
	qSize                    int
	qCount                   int
	file                     string
	line                     int
	sel                      *ElementSelect
	selIndex                 int
	partner                  *ElementChannel
	vc                       *clock.VectorClock
	wCl                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementChannel adds a new channel element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the channel
//   - opC string: The operation on the channel
//   - cl string: Whether the channel was finished because it was closed
//   - oId string: The id of the other communication
//   - qSize string: The size of the channel queue
//   - qCount string: The number of elements in the queue
//   - pos string: The position of the channel operation in the code
//
// Returns:
//   - error
func (this *Trace) AddTraceElementChannel(routine int, tPre string,
	tPost string, id string, opC string, cl string, oID string, qSize string,
	qCount string, pos string) error {

	tPreInt, err := strconv.Atoi(tPre)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt := -1
	if id != "*" {
		idInt, err = strconv.Atoi(id)
		if err != nil {
			return errors.New("id is not an integer")
		}
	}

	var opCInt OperationType
	switch opC {
	case "S":
		opCInt = ChannelSend
	case "R":
		opCInt = ChannelRecv
	case "C":
		opCInt = ChannelClose
	default:
		return errors.New("opC is not a valid operation")
	}

	clBool, err := strconv.ParseBool(cl)
	if err != nil {
		return errors.New("cl is not a boolean")
	}

	oIDInt, err := strconv.Atoi(oID)
	if err != nil {
		return errors.New("oId is not an integer")
	}

	qSizeInt, err := strconv.Atoi(qSize)
	if err != nil {
		return fmt.Errorf("qSize '%s' is not an integer", qSize)
	}

	qCountInt, err := strconv.Atoi(qCount)
	if err != nil {
		return fmt.Errorf("qCount '%s' is not an integer", qCount)
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementChannel{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		objId:                    idInt,
		op:                       opCInt,
		cl:                       clBool,
		oID:                      oIDInt,
		qSize:                    qSizeInt,
		qCount:                   qCountInt,
		file:                     file,
		line:                     line,
		selIndex:                 -1,
		vc:                       nil,
		wCl:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
	}

	elem.findPartner(this)

	this.AddElement(&elem)
	return nil
}

// GetPartner returns the partner of the channel operation
//
// Returns:
//   - *TraceElementChannel: The partner of the channel operation
func (this *ElementChannel) GetPartner() *ElementChannel {
	return this.partner
}

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementChannel) GetObjId() int {
	return this.objId
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementChannel) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element
//
// Returns:
//   - int: The tPre of the element
func (this *ElementChannel) GetTPre() int {
	return this.tPre
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - float32: The time of the element
func (this *ElementChannel) GetTSort() int {
	if this.tPost == 0 {
		return math.MaxInt
	}
	return this.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementChannel) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementChannel) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (this *ElementChannel) GetFile() string {
	return this.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (this *ElementChannel) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form C@[file]:[line]@[tPre]. If it is part of a select,
// it has the form C@[file]:[line]@[tPre]@[index]
//
// Returns:
//   - string: The tID of the element
func (this *ElementChannel) GetTID() string {
	tID := "C@" + this.GetPos() + "@" + strconv.Itoa(this.tPre)
	if this.selIndex != -1 {
		tID += "@" + strconv.Itoa(this.selIndex)
	}
	return tID
}

// GetOID returns the operation ID of the element
//
// Returns:
//   - int: The oID of the element
func (this *ElementChannel) GetOID() int {
	return this.oID
}

// IsBuffered returns if the channel is buffered
//
// Returns:
//   - bool: Whether the channel operation is buffered
func (this *ElementChannel) IsBuffered() bool {
	return this.qSize != 0
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementChannel) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementChannel) SetWVc(vc *clock.VectorClock) {
	this.wCl = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementChannel) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementChannel) GetWVC() *clock.VectorClock {
	return this.wCl
}

// GetTPost returns the tPost of the element
//
// Returns:
//   - int: The tPost of the element
func (this *ElementChannel) GetTPost() int {
	return this.tPost
}

// GetObjType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementChannel) GetType(operation bool) OperationType {
	if !operation {
		return Channel
	}

	return this.op
}

// GetQCount returns the number of elems in the queue after the operation
//
// Returns:
//   - int: The number of elems in the queue after the operation
func (this *ElementChannel) GetQCount() int {
	return this.qCount
}

// GetQCount sets the number of elems in the queue after the operation
//
// Parameter:
//   - qCount int: The number of elems in the queue after the operation
func (this *ElementChannel) SetQCount(qc int) {
	this.qCount = qc
}

// GetQSize returns the size of the buffer
//
// Returns:
//   - int: the size of the buffer
func (this *ElementChannel) GetQSize() int {
	return this.qSize
}

// GetSelect returns the select the element is a part of, if it is not part
// of a select, it returns nil
//
// Returns:
//   - VectorClock: The select the element is a part of, if not in select it is nil
func (this *ElementChannel) GetSelect() *ElementSelect {
	return this.sel
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementChannel) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same channel
func (this *ElementChannel) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Channel {
		return false
	}

	return this.objId == elem.GetObjId()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementChannel) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
// - time int: The tPre and tPost of the element
func (this *ElementChannel) SetT(time int) {
	this.tPre = time
	this.tPost = time
}

// SetPartner sets the partner of the channel operation
//
// Parameter:
//   - partner *TraceElementChannel: The partner of the channel operation
func (this *ElementChannel) SetPartner(partner *ElementChannel) {
	this.partner = partner
}

// SetClosed sets the cl value to closed
//
// Parameter:
//   - closed bool: the new cl value
func (this *ElementChannel) SetClosed(closed bool) {
	this.cl = closed
}

// GetClosed returns if the channel was closed
//
// Returns:
//   - bool: cl
func (this *ElementChannel) GetClosed() bool {
	return this.cl
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementChannel) SetTPre(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}

	if this.sel != nil {
		this.sel.SetTPre2(tPre)
	}
}

// SetTPre2 sets the tPre of the element. It does not set the tPre of the select operation
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementChannel) SetTPre2(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}
}

// SetTPost sets the tPost of the element.
//
// Parameter:
//   - tPost int: The tPost of the element
func (this *ElementChannel) SetTPost(tPost int) {
	this.tPost = tPost
	if this.sel != nil {
		this.sel.SetTPost2(tPost)
	}
}

// SetTPost2 sets the tPost of the element. It does not set the tPost of the select operation
//
// Parameter:
//   - tPost int: The tPost of the element
func (this *ElementChannel) SetTPost2(tPost int) {
	this.tPost = tPost
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTSort(tPost int) {
	this.SetTPre(tPost)
	this.tPost = tPost

	if this.sel != nil {
		this.sel.SetTSort2(tPost)
	}
}

// SetTSort2 sets the timer, that is used for the sorting of the trace.
// It does not set the tPost of the select operation
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTSort2(tPost int) {
	this.SetTPre(tPost)
	this.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}

	if this.sel != nil {
		this.sel.SetTWithoutNotExecuted2(tSort)
	}
}

// SetTWithoutNotExecuted2 sets the timer, that is used for the sorting of the trace, only if the original
// value was not 0. Do not set the tPost of the select operation
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementChannel) SetTWithoutNotExecuted2(tSort int) {
	this.SetTPre(tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// SetOID sets the operation ID of the element
//
// Parameter:
//   - oID int: The oID of the element
func (this *ElementChannel) SetOID(oID int) {
	this.oID = oID
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementChannel) ToString() string {
	return this.toStringSep(",", false)
}

// ToStringSep returns the simple string representation of the element with a
// custom separator
//
// Parameter:
//   - sep string: The separator between the values
//   - sel bool: Whether the channel is part of a select do not add time and pos)
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementChannel) toStringSep(sep string, sel bool) string {
	op := string(string(this.op)[1])

	cl := "f"
	if this.cl {
		cl = "t"
	}

	timeString := ""
	posStr := ""
	if !sel {
		timeString = fmt.Sprintf("%s%d%s%d", sep, this.GetTPre(), sep, this.GetTPost())
		posStr = sep + this.GetPos()
	}

	return fmt.Sprintf("C%s%s%d%s%s%s%s%s%d%s%d%s%d%s", timeString, sep, this.objId, sep, op, sep, cl, sep, this.oID, sep, this.qSize, sep, this.qCount, posStr)
}

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementChannel) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementChannel) setID(ID int) {
	this.id = ID
}

// Copy creates a copy of the channel element
//
//   - mapping map[string]Element: map containing all already copied elements,
//     Used to avoid double copy of references
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementChannel) Copy(mapping map[string]Element, keep bool) Element {
	tID := this.GetTID()
	if existing, ok := mapping[tID]; ok {
		return existing
	}

	if !keep {
		newCh := ElementChannel{
			id:                       this.id,
			index:                    0,
			routine:                  this.routine,
			tPre:                     0,
			tPost:                    0,
			objId:                    this.objId,
			op:                       this.op,
			cl:                       false,
			oID:                      0,
			qSize:                    this.qSize,
			qCount:                   0,
			file:                     this.file,
			line:                     this.line,
			selIndex:                 this.selIndex,
			vc:                       nil,
			wCl:                      nil,
			numberConcurrent:         0,
			numberConcurrentWeak:     0,
			numberConcurrentSame:     0,
			numberConcurrentWeakSame: 0,
		}

		mapping[tID] = &newCh

		var newPartner *ElementChannel
		if this.partner != nil {
			newPartner = this.partner.Copy(mapping, keep).(*ElementChannel)
		}

		var newSelect *ElementSelect
		if this.sel != nil {
			newSelect = this.sel.Copy(mapping, keep).(*ElementSelect)
		}

		newCh.partner = newPartner
		newCh.sel = newSelect

		return &newCh
	}

	newCh := ElementChannel{
		id:                       this.id,
		index:                    this.index,
		routine:                  this.routine,
		tPre:                     this.tPre,
		tPost:                    this.tPost,
		objId:                    this.objId,
		op:                       this.op,
		cl:                       this.cl,
		oID:                      this.oID,
		qSize:                    this.qSize,
		qCount:                   this.qCount,
		file:                     this.file,
		line:                     this.line,
		selIndex:                 this.selIndex,
		vc:                       this.vc.Copy(),
		wCl:                      this.wCl.Copy(),
		numberConcurrent:         this.numberConcurrent,
		numberConcurrentWeak:     this.numberConcurrentWeak,
		numberConcurrentSame:     this.numberConcurrentSame,
		numberConcurrentWeakSame: this.numberConcurrentWeakSame,
	}

	mapping[tID] = &newCh

	var newPartner *ElementChannel
	if this.partner != nil {
		newPartner = this.partner.Copy(mapping, keep).(*ElementChannel)
	}

	var newSelect *ElementSelect
	if this.sel != nil {
		newSelect = this.sel.Copy(mapping, keep).(*ElementSelect)
	}

	newCh.partner = newPartner
	newCh.sel = newSelect

	return &newCh
}

// Find the partner of the channel operation
//
// Parameter:
//   - tr *Trace: the trace, the element is in
//
// Returns:
//   - *TraceElementChannel: The partner, -1 if not found
func (this *ElementChannel) findPartner(tr *Trace) *ElementChannel {
	id := this.GetObjId()
	oID := this.GetOID()

	// return -1 if closed by channel
	if this.GetClosed() || this.GetTPost() == 0 {
		return nil
	}

	// find partner has already been applied to the partner and the communication
	// was fund. An repeated search is not necessary
	if this.GetPartner() != nil {
		return this.GetPartner()
	}

	// check if partner has already been processed
	if partner, ok := tr.channelWithoutPartner[id][oID]; ok {
		if this.IsEqual(partner) {
			return nil
		}

		// partner was already processed
		this.SetPartner(partner)
		partner.SetPartner(this)

		delete(tr.channelWithoutPartner[id], oID)

		return partner
	}

	if tr.channelWithoutPartner[id] == nil {
		tr.channelWithoutPartner[id] = make(map[int]*ElementChannel)
	}
	tr.channelWithoutPartner[id][oID] = this

	return nil
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
func (this *ElementChannel) GetNumberConcurrent(weak, sameElem bool) int {
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
func (this *ElementChannel) SetNumberConcurrent(c int, weak, sameElem bool) {
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
}
