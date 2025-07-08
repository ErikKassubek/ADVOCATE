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

	"advocate/analysis/concurrent/clock"
	"advocate/utils/log"
)

// OpChannel is an enum for opC
type OpChannel int

// Values for the opChannel enum
const (
	SendOp OpChannel = iota
	RecvOp
	CloseOp
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
//   - opC OpChannel: The operation on the channel
//   - cl bool: Whether the channel has closed
//   - oID int: The id of the other communication
//   - qSize int: The size of the channel queue
//   - qCount int: The number of elements in the queue after the operation
//   - file string: The file of the channel operation in the code
//   - line int: The line of the channel operation
//   - sel *traceElementSelect: The select operation, if the channel operation
//     is part of a select, otherwise nil
//   - partner *ElementChannel: The partner of the channel operation
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - children []TraceElement: children in partial order graph
//   - parent []TraceElement: parents in partial order graph
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
type ElementChannel struct {
	traceID          int
	index            int
	routine          int
	tPre             int
	tPost            int
	id               int
	opC              OpChannel
	cl               bool
	oID              int
	qSize            int
	qCount           int
	file             string
	line             int
	sel              *ElementSelect
	partner          *ElementChannel
	vc               *clock.VectorClock
	wVc              *clock.VectorClock
	children         []Element
	parents          []Element
	numberConcurrent int
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

	var opCInt OpChannel
	switch opC {
	case "S":
		opCInt = SendOp
	case "R":
		opCInt = RecvOp
	case "C":
		opCInt = CloseOp
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
		index:            t.numberElemsInTrace[routine],
		routine:          routine,
		tPre:             tPreInt,
		tPost:            tPostInt,
		id:               idInt,
		opC:              opCInt,
		cl:               clBool,
		oID:              oIDInt,
		qSize:            qSizeInt,
		qCount:           qCountInt,
		file:             file,
		line:             line,
		vc:               nil,
		wVc:              nil,
		children:         make([]Element, 0),
		parents:          make([]Element, 0),
		numberConcurrent: -1,
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
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (ch *ElementChannel) GetTID() string {
	return ch.GetPos() + "@" + strconv.Itoa(ch.tPre)
}

// GetOID returns the operation ID of the element
//
// Returns:
//   - int: The oID of the element
func (ch *ElementChannel) GetOID() int {
	return ch.oID
}

// GetOpC returns the operation of the channel  (opC)
//
// Returns:
//   - OpChannel: the operation
func (ch *ElementChannel) GetOpC() OpChannel {
	return ch.opC
}

// IsBuffered returns if the channel is buffered
//
// Returns:
//   - bool: Whether the channel operation is buffered
func (ch *ElementChannel) IsBuffered() bool {
	return ch.qSize != 0
}

// Operation returns the type of the operation
//
// Returns:
//   - OpChannel: The type of the operation
func (ch *ElementChannel) Operation() OpChannel {
	return ch.opC
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
	ch.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (ch *ElementChannel) GetVC() *clock.VectorClock {
	return ch.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (ch *ElementChannel) GetWVc() *clock.VectorClock {
	return ch.wVc
}

// GetTPost returns the tPost of the element
//
// Returns:
//   - int: The tPost of the element
func (ch *ElementChannel) GetTPost() int {
	return ch.tPost
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (ch *ElementChannel) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeChannel
	}

	switch ch.opC {
	case SendOp:
		return ObjectTypeChannel + "S"
	case RecvOp:
		return ObjectTypeChannel + "R"
	case CloseOp:
		return ObjectTypeChannel + "C"
	}
	return ObjectTypeChannel
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
	op := ""
	switch ch.opC {
	case SendOp:
		op = "S"
	case RecvOp:
		op = "R"
	case CloseOp:
		op = "C"
	default:
		log.Error("Unknown channel operation: " + strconv.Itoa(int(ch.opC)))
		op = "-"
	}

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
// Returns:
//   - TraceElement: The copy of the element
func (ch *ElementChannel) Copy() Element {
	children := make([]Element, len(ch.children))
	copy(children, ch.children)
	parents := make([]Element, len(ch.parents))
	copy(parents, ch.parents)

	newCh := ElementChannel{
		traceID:          ch.traceID,
		index:            ch.index,
		routine:          ch.routine,
		tPre:             ch.tPre,
		tPost:            ch.tPost,
		id:               ch.id,
		opC:              ch.opC,
		cl:               ch.cl,
		oID:              ch.oID,
		qSize:            ch.qSize,
		file:             ch.file,
		line:             ch.line,
		sel:              ch.sel,
		partner:          ch.partner,
		vc:               ch.vc.Copy(),
		wVc:              ch.wVc.Copy(),
		children:         children,
		parents:          parents,
		numberConcurrent: ch.numberConcurrent,
	}
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

// AddChild adds an element as a child of this node in the partial order graph
//
// Parameter:
//   - elem *TraceElement: the element to add
func (ch *ElementChannel) AddChild(elem Element) {
	ch.children = append(ch.children, elem)
}

// AddParent adds an element as a parent of this node in the partial order graph
//
// Parameter:
//   - elem *TraceElement: the element to add
func (ch *ElementChannel) AddParent(elem Element) {
	ch.parents = append(ch.parents, elem)
}

// GetChildren returns all children of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the children
func (ch *ElementChannel) GetChildren() []Element {
	return ch.children
}

// GetParents returns all parents of this node in the partial order graph
//
// Returns:
//   - []*TraceElement: the parents
func (ch *ElementChannel) GetParents() []Element {
	return ch.parents
}

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - number of concurrent element, or -1
func (ch *ElementChannel) GetNumberConcurrent() int {
	return ch.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
func (ch *ElementChannel) SetNumberConcurrent(c int) {
	ch.numberConcurrent = c
}
