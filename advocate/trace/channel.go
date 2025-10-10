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
//   - traceID: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the channel
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
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	op                       ObjectType
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
func (t *Trace) AddTraceElementChannel(routine int, tPre string,
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

	var opCInt ObjectType
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
		index:                    t.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
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

	elem.findPartner(t)

	t.AddElement(&elem)
	return nil
}

// GetPartner returns the partner of the channel operation
//
// Returns:
//   - *TraceElementChannel: The partner of the channel operation
func (ch *ElementChannel) GetPartner() *ElementChannel {
	return ch.partner
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (ch *ElementChannel) GetID() int {
	return ch.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (ch *ElementChannel) GetRoutine() int {
	return ch.routine
}

// GetTPre returns the tPre of the element
//
// Returns:
//   - int: The tPre of the element
func (ch *ElementChannel) GetTPre() int {
	return ch.tPre
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - float32: The time of the element
func (ch *ElementChannel) GetTSort() int {
	if ch.tPost == 0 {
		return math.MaxInt
	}
	return ch.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (ch *ElementChannel) GetPos() string {
	return fmt.Sprintf("%s:%d", ch.file, ch.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (ch *ElementChannel) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", ch.routine, ch.file, ch.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (ch *ElementChannel) GetFile() string {
	return ch.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (ch *ElementChannel) GetLine() int {
	return ch.line
}

// GetTID returns the tID of the element.
// The tID is a string of form C@[file]:[line]@[tPre]. If it is part of a select,
// it has the form C@[file]:[line]@[tPre]@[index]
//
// Returns:
//   - string: The tID of the element
func (ch *ElementChannel) GetTID() string {
	tID := "C@" + ch.GetPos() + "@" + strconv.Itoa(ch.tPre)
	if ch.selIndex != -1 {
		tID += "@" + strconv.Itoa(ch.selIndex)
	}
	return tID
}

// GetOID returns the operation ID of the element
//
// Returns:
//   - int: The oID of the element
func (ch *ElementChannel) GetOID() int {
	return ch.oID
}

// IsBuffered returns if the channel is buffered
//
// Returns:
//   - bool: Whether the channel operation is buffered
func (ch *ElementChannel) IsBuffered() bool {
	return ch.qSize != 0
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (ch *ElementChannel) SetVc(vc *clock.VectorClock) {
	ch.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (ch *ElementChannel) SetWVc(vc *clock.VectorClock) {
	ch.wCl = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (ch *ElementChannel) GetVC() *clock.VectorClock {
	return ch.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (ch *ElementChannel) GetWVC() *clock.VectorClock {
	return ch.wCl
}

// GetTPost returns the tPost of the element
//
// Returns:
//   - int: The tPost of the element
func (ch *ElementChannel) GetTPost() int {
	return ch.tPost
}

// GetObjType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (ch *ElementChannel) GetType(operation bool) ObjectType {
	if !operation {
		return Channel
	}

	return ch.op
}

// GetQCount returns the number of elems in the queue after the operation
//
// Returns:
//   - VectorClock: The number of elems in the queue after the operation
func (ch *ElementChannel) GetQCount() int {
	return ch.qCount
}

// GetQSize returns the size of the buffer
//
// Returns:
//   - int: the size of the buffer
func (ch *ElementChannel) GetQSize() int {
	return ch.qSize
}

// GetSelect returns the select the element is a part of, if it is not part
// of a select, it returns nil
//
// Returns:
//   - VectorClock: The select the element is a part of, if not in select it is nil
func (ch *ElementChannel) GetSelect() *ElementSelect {
	return ch.sel
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (ch *ElementChannel) IsEqual(elem Element) bool {
	return ch.routine == elem.GetRoutine() && ch.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same channel
func (ch *ElementChannel) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Channel {
		return false
	}

	return ch.id == elem.GetID()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (ch *ElementChannel) GetTraceIndex() (int, int) {
	return ch.routine, ch.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
// - time int: The tPre and tPost of the element
func (ch *ElementChannel) SetT(time int) {
	ch.tPre = time
	ch.tPost = time
}

// SetPartner sets the partner of the channel operation
//
// Parameter:
//   - partner *TraceElementChannel: The partner of the channel operation
func (ch *ElementChannel) SetPartner(partner *ElementChannel) {
	ch.partner = partner
}

// SetClosed sets the cl value to closed
//
// Parameter:
//   - closed bool: the new cl value
func (ch *ElementChannel) SetClosed(closed bool) {
	ch.cl = closed
}

// GetClosed returns if the channel was closed
//
// Returns:
//   - bool: cl
func (ch *ElementChannel) GetClosed() bool {
	return ch.cl
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (ch *ElementChannel) SetTPre(tPre int) {
	ch.tPre = tPre
	if ch.tPost != 0 && ch.tPost < tPre {
		ch.tPost = tPre
	}

	if ch.sel != nil {
		ch.sel.SetTPre2(tPre)
	}
}

// SetTPre2 sets the tPre of the element. It does not set the tPre of the select operation
//
// Parameter:
//   - tPre int: The tPre of the element
func (ch *ElementChannel) SetTPre2(tPre int) {
	ch.tPre = tPre
	if ch.tPost != 0 && ch.tPost < tPre {
		ch.tPost = tPre
	}
}

// SetTPost sets the tPost of the element.
//
// Parameter:
//   - tPost int: The tPost of the element
func (ch *ElementChannel) SetTPost(tPost int) {
	ch.tPost = tPost
	if ch.sel != nil {
		ch.sel.SetTPost2(tPost)
	}
}

// SetTPost2 sets the tPost of the element. It does not set the tPost of the select operation
//
// Parameter:
//   - tPost int: The tPost of the element
func (ch *ElementChannel) SetTPost2(tPost int) {
	ch.tPost = tPost
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (ch *ElementChannel) SetTSort(tPost int) {
	ch.SetTPre(tPost)
	ch.tPost = tPost

	if ch.sel != nil {
		ch.sel.SetTSort2(tPost)
	}
}

// SetTSort2 sets the timer, that is used for the sorting of the trace.
// It does not set the tPost of the select operation
//
// Parameter:
//   - tSort int: The timer of the element
func (ch *ElementChannel) SetTSort2(tPost int) {
	ch.SetTPre(tPost)
	ch.tPost = tPost
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (ch *ElementChannel) SetTWithoutNotExecuted(tSort int) {
	ch.SetTPre(tSort)
	if ch.tPost != 0 {
		ch.tPost = tSort
	}

	if ch.sel != nil {
		ch.sel.SetTWithoutNotExecuted2(tSort)
	}
}

// SetTWithoutNotExecuted2 sets the timer, that is used for the sorting of the trace, only if the original
// value was not 0. Do not set the tPost of the select operation
//
// Parameter:
//   - tSort int: The timer of the element
func (ch *ElementChannel) SetTWithoutNotExecuted2(tSort int) {
	ch.SetTPre(tSort)
	if ch.tPost != 0 {
		ch.tPost = tSort
	}
}

// SetOID sets the operation ID of the element
//
// Parameter:
//   - oID int: The oID of the element
func (ch *ElementChannel) SetOID(oID int) {
	ch.oID = oID
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (ch *ElementChannel) ToString() string {
	return ch.toStringSep(",", false)
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
func (ch *ElementChannel) toStringSep(sep string, sel bool) string {
	op := string(ch.op)[1]

	cl := "f"
	if ch.cl {
		cl = "t"
	}

	timeString := ""
	posStr := ""
	if !sel {
		timeString = fmt.Sprintf("%s%d%s%d", sep, ch.GetTPre(), sep, ch.GetTPost())
		posStr = sep + ch.GetPos()
	}

	return fmt.Sprintf("C%s%s%d%s%s%s%s%s%d%s%d%s%d%s", timeString, sep, ch.id, sep, op, sep, cl, sep, ch.oID, sep, ch.qSize, sep, ch.qCount, posStr)
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (ch *ElementChannel) GetTraceID() int {
	return ch.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (ch *ElementChannel) setTraceID(ID int) {
	ch.traceID = ID
}

// Copy creates a copy of the channel element
//
//   - mapping map[string]Element: map containing all already copied elements.
//     Used to avoid double copy of references
//
// Returns:
//   - TraceElement: The copy of the element
func (ch *ElementChannel) Copy(mapping map[string]Element) Element {
	tID := ch.GetTID()
	if existing, ok := mapping[tID]; ok {
		return existing
	}

	newCh := ElementChannel{
		traceID:                  ch.traceID,
		index:                    ch.index,
		routine:                  ch.routine,
		tPre:                     ch.tPre,
		tPost:                    ch.tPost,
		id:                       ch.id,
		op:                       ch.op,
		cl:                       ch.cl,
		oID:                      ch.oID,
		qSize:                    ch.qSize,
		file:                     ch.file,
		line:                     ch.line,
		selIndex:                 ch.selIndex,
		vc:                       ch.vc.Copy(),
		wCl:                      ch.wCl.Copy(),
		numberConcurrent:         ch.numberConcurrent,
		numberConcurrentWeak:     ch.numberConcurrentWeak,
		numberConcurrentSame:     ch.numberConcurrentSame,
		numberConcurrentWeakSame: ch.numberConcurrentWeakSame,
	}

	mapping[tID] = &newCh

	var newPartner *ElementChannel
	if ch.partner != nil {
		newPartner = ch.partner.Copy(mapping).(*ElementChannel)
	}

	var newSelect *ElementSelect
	if ch.sel != nil {
		newSelect = ch.sel.Copy(mapping).(*ElementSelect)
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
func (ch *ElementChannel) findPartner(tr *Trace) *ElementChannel {
	id := ch.GetID()
	oID := ch.GetOID()

	// return -1 if closed by channel
	if ch.GetClosed() || ch.GetTPost() == 0 {
		return nil
	}

	// find partner has already been applied to the partner and the communication
	// was fund. An repeated search is not necessary
	if ch.GetPartner() != nil {
		return ch.GetPartner()
	}

	// check if partner has already been processed
	if partner, ok := tr.channelWithoutPartner[id][oID]; ok {
		if ch.IsEqual(partner) {
			return nil
		}

		// partner was already processed
		ch.SetPartner(partner)
		partner.SetPartner(ch)

		delete(tr.channelWithoutPartner[id], oID)

		return partner
	}

	if tr.channelWithoutPartner[id] == nil {
		tr.channelWithoutPartner[id] = make(map[int]*ElementChannel)
	}
	tr.channelWithoutPartner[id][oID] = ch

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
func (ch *ElementChannel) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return ch.numberConcurrentWeakSame
		}
		return ch.numberConcurrentWeak
	}
	if sameElem {
		return ch.numberConcurrentSame
	}
	return ch.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (ch *ElementChannel) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			ch.numberConcurrentWeakSame = c
		} else {
			ch.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			ch.numberConcurrentSame = c
		} else {
			ch.numberConcurrent = c
		}
	}
}
