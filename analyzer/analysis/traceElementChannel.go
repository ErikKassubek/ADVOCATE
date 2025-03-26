// Copyrigth (c) 2024 Erik Kassubek
//
// File: traceElementChannel.go
// Brief: Struct and functions for channel operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package analysis

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"analyzer/clock"
	"analyzer/timer"
	"analyzer/utils"
)

// enum for opC
type OpChannel int

const (
	SendOp OpChannel = iota
	RecvOp
	CloseOp
)

/*
* TraceElementChannel is a trace element for a channel
* MARK: Struct
* Fields:
*   routine (int): The routine id
*   tpre (int): The timestamp at the start of the event
*   tpost (int): The timestamp at the end of the event
*   id (int): The id of the channel
*   opC (int, enum): The operation on the channel
*   cl (bool): Whether the channel has closed
*   oId (int): The id of the other communication
*   qSize (int): The size of the channel queue
*   qCount (int): The number of elements in the queue after the operation
*   file (string), line(int): The position of the channel operation in the code
*   sel (*traceElementSelect): The select operation, if the channel operation
*       is part of a select, otherwise nil
*   partner (*TraceElementChannel): The partner of the channel operation
*   tID (string): The id of the trace element, contains the position and the tpre
 */
type TraceElementChannel struct {
	routine int
	tPre    int
	tPost   int
	id      int
	opC     OpChannel
	cl      bool
	oID     int
	qSize   int
	qCount  int
	file    string
	line    int
	sel     *TraceElementSelect
	partner *TraceElementChannel
	vc      clock.VectorClock
}

/*
* Create a new channel trace element
* MARK: New
* Args:
*   routine (int): The routine id
*   tPre (string): The timestamp at the start of the event
*   tPost (string): The timestamp at the end of the event
*   id (string): The id of the channel
*   opC (string): The operation on the channel
*   cl (string): Whether the channel was finished because it was closed
*   oId (string): The id of the other communication
*   qSize (string): The size of the channel queue
*   qCount (string): The number of elements in the queue
*   pos (string): The position of the channel operation in the code
 */
func AddTraceElementChannel(routine int, tPre string,
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
		return errors.New("qSize is not an integer")
	}

	qCountInt, err := strconv.Atoi(qCount)
	if err != nil {
		return errors.New("qSize is not an integer")
	}

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementChannel{
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		opC:     opCInt,
		cl:      clBool,
		oID:     oIDInt,
		qSize:   qSizeInt,
		qCount:  qCountInt,
		file:    file,
		line:    line,
	}

	// check if partner was already processed, otherwise add to channelWithoutPartner
	if tPostInt != 0 {
		if _, ok := channelWithoutPartner[idInt][oIDInt]; ok {
			elem.partner = channelWithoutPartner[idInt][oIDInt]
			channelWithoutPartner[idInt][oIDInt].partner = &elem
			delete(channelWithoutPartner[idInt], oIDInt)
		} else {
			if _, ok := channelWithoutPartner[idInt]; !ok {
				channelWithoutPartner[idInt] = make(map[int]*TraceElementChannel)
			}

			channelWithoutPartner[idInt][oIDInt] = &elem
		}
	}

	return AddElementToTrace(&elem)
}

// MARK: Getter

/*
* Get the partner of the channel operation
* Returns:
*   *TraceElementChannel: The partner of the channel operation
 */
func (ch *TraceElementChannel) GetPartner() *TraceElementChannel {
	return ch.partner
}

/*
 * Get the id of the element
 * Returns:
 *   int: The id of the element
 */
func (ch *TraceElementChannel) GetID() int {
	return ch.id
}

/*
	* Get the routine of the element
 * Returns:
 *   int: The routine of the element
*/
func (ch *TraceElementChannel) GetRoutine() int {
	return ch.routine
}

/*
 * Get the tpre of the element
 * Returns:
 *   int: The tpre of the element
 */
func (ch *TraceElementChannel) GetTPre() int {
	return ch.tPre
}

/*
 * Get the timer, that is used for the sorting of the trace
 * Returns:
 *   float32: The time of the element
 */
func (ch *TraceElementChannel) GetTSort() int {
	if ch.tPost == 0 {
		return math.MaxInt
	}
	return ch.tPost
}

/*
 * Get the position of the operation.
 * Returns:
 *   string: The position of the element
 */
func (ch *TraceElementChannel) GetPos() string {
	return fmt.Sprintf("%s:%d", ch.file, ch.line)
}

func (ch *TraceElementChannel) GetFile() string {
	return ch.file
}

func (ch *TraceElementChannel) GetLine() int {
	return ch.line
}

/*
 * Get the tID of the element.
 * Returns:
 *   string: The tID of the element
 */
func (ch *TraceElementChannel) GetTID() string {
	return ch.GetPos() + "@" + strconv.Itoa(ch.tPre)
}

/*
 * Get the oID of the element
 * Returns:
 *   int: The oID of the element
 */
func (ch *TraceElementChannel) GetOID() int {
	return ch.oID
}

/*
 * Check if the channel operation is buffered
 * Returns:
 *   bool: Whether the channel operation is buffered
 */
func (ch *TraceElementChannel) IsBuffered() bool {
	return ch.qSize != 0
}

/*
 * Get the type of the operation
 * Returns:
 *   OpChannel: The type of the operation
 */
func (ch *TraceElementChannel) Operation() OpChannel {
	return ch.opC
}

/*
 * Get the vector clock of the element
 * Returns:
 *   VectorClock: The vector clock of the element
 */
func (ch *TraceElementChannel) GetVC() clock.VectorClock {
	return ch.vc
}

/*
 * Get the tpost of the element
 * Returns:
 *   int: The tpost of the element
 */
func (ch *TraceElementChannel) GetTPost() int {
	return ch.tPost
}

/*
 * Get the string representation of the object type
 */
func (ch *TraceElementChannel) GetObjType() string {
	switch ch.opC {
	case SendOp:
		return "CS"
	case RecvOp:
		return "CR"
	case CloseOp:
		return "CC"
	}
	return "C"
}

func (ch *TraceElementChannel) GetQCount() int {
	return ch.qCount
}

func (ch *TraceElementChannel) GetSelect() *TraceElementSelect {
	return ch.sel
}

// MARK: Setter

/*
 * Set the tPre and tPost of the element
 * Args:
 *   time (int): The tPre and tPost of the element
 */
func (ch *TraceElementChannel) SetT(time int) {
	ch.tPre = time
	ch.tPost = time
}

/*
 * Set the partner of the channel operation
 * Args:
 *   partner (*TraceElementChannel): The partner of the channel operation
 */
func (ch *TraceElementChannel) SetPartner(partner *TraceElementChannel) {
	ch.partner = partner
}

/*
* Set the tpre of the element.
* Args:
 *   tPre (int): The tpre of the element
*/
func (ch *TraceElementChannel) SetTPre(tPre int) {
	ch.tPre = tPre
	if ch.tPost != 0 && ch.tPost < tPre {
		ch.tPost = tPre
	}

	if ch.sel != nil {
		ch.sel.SetTPre2(tPre)
	}
}

/*
* Set the tpre of the element. Do not set the tpre of the select operation
* Args:
 *   tPre (int): The tpre of the element
*/
func (ch *TraceElementChannel) SetTPre2(tPre int) {
	ch.tPre = tPre
	if ch.tPost != 0 && ch.tPost < tPre {
		ch.tPost = tPre
	}
}

/*
 * Set the tpost of the element.
 * Args:
 *   tPost (int): The tpost of the element
 */
func (ch *TraceElementChannel) SetTPost(tPost int) {
	ch.tPost = tPost
	if ch.sel != nil {
		ch.sel.SetTPost2(tPost)
	}
}

/*
 * Set the tpost of the element. Do not set the tpost of the select operation
 * Args:
 *   tPost (int): The tpost of the element
 */
func (ch *TraceElementChannel) SetTPost2(tPost int) {
	ch.tPost = tPost
}

/*
 * Set the timer, that is used for the sorting of the trace
 * Args:
 *   tSort (int): The timer of the element
 */
func (ch *TraceElementChannel) SetTSort(tpost int) {
	ch.SetTPre(tpost)
	ch.tPost = tpost

	if ch.sel != nil {
		ch.sel.SetTSort2(tpost)
	}
}

/*
 * Set the timer, that is used for the sorting of the trace. Do not set the tpost of the select operation
 * Args:
 *   tSort (int): The timer of the element
 */
func (ch *TraceElementChannel) SetTSort2(tpost int) {
	ch.SetTPre(tpost)
	ch.tPost = tpost
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0
 * Args:
 *   tSort (int): The timer of the element
 */
func (ch *TraceElementChannel) SetTWithoutNotExecuted(tSort int) {
	ch.SetTPre(tSort)
	if ch.tPost != 0 {
		ch.tPost = tSort
	}

	if ch.sel != nil {
		ch.sel.SetTWithoutNotExecuted2(tSort)
	}
}

/*
 * Set the timer, that is used for the sorting of the trace, only if the original
 * value was not 0. Do not set the tpost of the select operation
 * Args:
 *   tSort (int): The timer of the element
 */
func (ch *TraceElementChannel) SetTWithoutNotExecuted2(tSort int) {
	ch.SetTPre(tSort)
	if ch.tPost != 0 {
		ch.tPost = tSort
	}
}

/*
 * Set the oID of the element
 * Args:
 *   oID (int): The oID of the element
 */
func (ch *TraceElementChannel) SetOID(oID int) {
	ch.oID = oID
}

// MARK: ToString

/*
 * Get the simple string representation of the element
 * Returns:
 *   string: The simple string representation of the element
 */
func (ch *TraceElementChannel) ToString() string {
	return ch.toStringSep(",", true)
}

/*
 * Get the simple string representation of the element
 * Args:
 *   sep (string): The separator between the values
 *   pos (bool): Whether the position should be included
 * Returns:
 *   string: The simple string representation of the element
 */
func (ch *TraceElementChannel) toStringSep(sep string, pos bool) string {
	op := ""
	switch ch.opC {
	case SendOp:
		op = "S"
	case RecvOp:
		op = "R"
	case CloseOp:
		op = "C"
	default:
		utils.LogError("Unknown channel operation: " + strconv.Itoa(int(ch.opC)))
		op = "-"
	}

	cl := "f"
	if ch.cl {
		cl = "t"
	}

	posStr := ""
	if pos {
		posStr = sep + ch.GetPos()
	}

	return fmt.Sprintf("C%s%d%s%d%s%d%s%s%s%s%s%d%s%d%s%d%s", sep, ch.tPre, sep, ch.tPost, sep, ch.id, sep, op, sep, cl, sep, ch.oID, sep, ch.qSize, sep, ch.qCount, posStr)
}

var second = false

/*
 * Update and calculate the vector clock of the element
 * MARK: Vector Clock
 */
func (ch *TraceElementChannel) updateVectorClock() {
	timer.Start(timer.AnaHb)
	defer timer.Stop(timer.AnaHb)

	ch.vc = currentVCHb[ch.routine].Copy()

	if ch.tPost == 0 {
		return
	}

	if ch.partner == nil {
		ch.findPartner()
	}

	// hold back receive operations, until the send operation is processed
	for _, elem := range waitingReceive {
		if elem.oID <= maxOpID[ch.id] {
			if len(waitingReceive) != 0 {
				waitingReceive = waitingReceive[1:]
			}
			elem.updateVectorClock()
		}
	}

	if ch.IsBuffered() {
		if ch.opC == SendOp {
			maxOpID[ch.id] = ch.oID
		} else if ch.opC == RecvOp {
			if ch.oID > maxOpID[ch.id] && !ch.cl {
				waitingReceive = append(waitingReceive, ch)
				return
			}
		}

		switch ch.opC {
		case SendOp:
			Send(ch, currentVCHb, fifo)
		case RecvOp:
			if ch.cl { // recv on closed channel
				RecvC(ch, currentVCHb, true)
			} else {
				Recv(ch, currentVCHb, fifo)
			}
		case CloseOp:
			Close(ch, currentVCHb)
		default:
			err := "Unknown operation: " + ch.ToString()
			utils.LogError(err)
		}
	} else { // unbuffered channel
		switch ch.opC {
		case SendOp:
			if ch.partner != nil {
				ch.partner.vc = currentVCHb[ch.partner.routine].Copy()
				if ch.partner.sel != nil {
					ch.partner.sel.vc = currentVCHb[ch.partner.routine].Copy()
				}
				Unbuffered(ch, ch.partner, currentVCHb)
				// advance index of receive routine, send routine is already advanced
				increaseIndex(ch.partner.routine)
			} else {
				if ch.cl { // recv on closed channel
					SendC(ch)
				} else {
					StuckChan(ch.routine, currentVCHb)
				}
			}

		case RecvOp: // should not occur, but better save than sorry
			if ch.partner != nil {
				ch.partner.vc = currentVCHb[ch.partner.routine].Copy()
				Unbuffered(ch.partner, ch, currentVCHb)
				// advance index of receive routine, send routine is already advanced
				increaseIndex(ch.partner.routine)
			} else {
				if ch.cl { // recv on closed channel
					RecvC(ch, currentVCHb, false)
				} else {
					StuckChan(ch.routine, currentVCHb)
				}
			}
		case CloseOp:
			Close(ch, currentVCHb)
		default:
			err := "Unknown operation: " + ch.ToString()
			utils.LogError(err)
		}
	}
}

/*
 * Find the partner of the channel operation
 * MARK: Partner
 * Returns:
 *   int: The routine id of the partner, -1 if no partner was found
 */
func (ch *TraceElementChannel) findPartner() int {

	// return -1 if closed by channel
	if ch.cl {
		return -1
	}

	for routine, trace := range traces {
		if currentIndex[routine] == -1 {
			continue
		}
		// if routine == ch.routine {
		// 	continue
		// }
		elem := trace[currentIndex[routine]]

		if elem.ToString() == ch.ToString() {
			continue
		}

		switch e := elem.(type) {
		case *TraceElementChannel:
			if e.id == ch.id && e.oID == ch.oID {
				ch.partner = e
				e.partner = ch
				return routine
			}
		case *TraceElementSelect:
			if e.chosenCase.tPost != 0 &&
				e.chosenCase.id == ch.id &&
				e.chosenCase.oID == ch.oID {
				ch.partner = &e.chosenCase
				e.chosenCase.partner = ch
				return routine
			}
		default:
			continue
		}
	}
	return -1
}

// MARK: Copy

/*
 * Create a copy of the channel element
 * Returns:
 *   TraceElement: The copy of the element
 */
func (ch *TraceElementChannel) Copy() TraceElement {
	newCh := TraceElementChannel{
		routine: ch.routine,
		tPre:    ch.tPre,
		tPost:   ch.tPost,
		id:      ch.id,
		opC:     ch.opC,
		cl:      ch.cl,
		oID:     ch.oID,
		qSize:   ch.qSize,
		file:    ch.file,
		line:    ch.line,
		sel:     ch.sel,
		partner: ch.partner,
		vc:      ch.vc.Copy(),
	}
	return &newCh
}
