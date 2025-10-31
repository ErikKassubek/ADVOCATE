// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/cond.go
// Brief: Struct and functions for operations of conditional variables in the trace
//
// Author: Erik Kassubek
// Created: 2023-12-25
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ElementCond is a trace element for a condition variable
// Fields:
//   - traceID: id of the element, should never be changed
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the condition variable
//   - op objectType: The operation on the condition variable
//   - file string, The file of the condition variable operation in the code
//   - line int, The line of the condition variable operation in the code
//   - children []TraceElement: children in partial order graph
//   - parent []TraceElement: parents in partial order graph
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementCond struct {
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	op                       ObjectType
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementCond adds a new condition variable element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the condition variable
//   - opC string: The operation on the condition variable
//   - pos string: The position of the condition variable operation in the code
func (this *Trace) AddTraceElementCond(routine int, tPre string, tPost string, id string, opN string, pos string) error {
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
	var op ObjectType
	switch opN {
	case "W":
		op = CondWait
	case "S":
		op = CondSignal
	case "B":
		op = CondBroadcast
	default:
		return errors.New("op is not a valid operation")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementCond{
		index:                    this.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
		op:                       op,
		file:                     file,
		line:                     line,
		vc:                       nil,
		wVc:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
	}

	this.AddElement(&elem)
	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementCond) GetID() int {
	return this.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine id
func (this *ElementCond) GetRoutine() int {
	return this.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (this *ElementCond) GetTPre() int {
	return this.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (this *ElementCond) GetTPost() int {
	return this.tPost
}

// GetTSort returns the timer, that is used for sorting the trace
//
// Returns:
//   - int: The timer of the element
func (this *ElementCond) GetTSort() int {
	t := this.tPre
	if this.op == CondWait {
		t = this.tPost
	}
	if t == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return t
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (this *ElementCond) GetPos() string {
	return fmt.Sprintf("%s:%d", this.file, this.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementCond) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", this.routine, this.file, this.line)
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementCond) GetFile() string {
	return this.file
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementCond) GetLine() int {
	return this.line
}

// GetTID returns the tID of the element.
// The tID is a string of form D@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementCond) GetTID() string {
	return "D@" + this.GetPos() + "@" + strconv.Itoa(this.tPre)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementCond) SetVc(vc *clock.VectorClock) {
	this.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (this *ElementCond) SetWVc(vc *clock.VectorClock) {
	this.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementCond) GetVC() *clock.VectorClock {
	return this.vc
}

// GetWVC returns the vector clock of the element for the weak must happens before relation
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementCond) GetWVC() *clock.VectorClock {
	return this.wVc
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementCond) GetType(operation bool) ObjectType {
	if !operation {
		return Cond
	}

	return this.op
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementCond) IsEqual(elem Element) bool {
	return this.routine == elem.GetRoutine() && this.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same conditional variable
func (this *ElementCond) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Cond {
		return false
	}

	return this.id == elem.GetID()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementCond) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementCond) SetT(time int) {
	this.tPre = time
	this.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (this *ElementCond) SetTPre(tPre int) {
	this.tPre = tPre
	if this.tPost != 0 && this.tPost < tPre {
		this.tPost = tPre
	}
}

// SetTSort sets the timer that is used for sorting the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementCond) SetTSort(tSort int) {
	this.SetTPre(tSort)
	if this.op == CondWait {
		this.tPost = tSort
	}
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementCond) SetTWithoutNotExecuted(tSort int) {
	this.SetTPre(tSort)
	if this.op == CondWait {
		if this.tPost != 0 {
			this.tPost = tSort
		}
		return
	}
	if this.tPre != 0 {
		this.tPre = tSort
	}
}

// ToString returns the string representation of the element
//
// Returns:
//   - string: The string representation of the element
func (this *ElementCond) ToString() string {
	res := "D,"
	res += strconv.Itoa(this.tPre) + "," + strconv.Itoa(this.tPost) + ","
	res += strconv.Itoa(this.id) + ","
	switch this.op {
	case CondWait:
		res += "W"
	case CondSignal:
		res += "S"
	case CondBroadcast:
		res += "B"
	}
	res += "," + this.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementCond) GetTraceID() int {
	return this.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementCond) setTraceID(ID int) {
	this.traceID = ID
}

// Copy the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since conds do not contain reference to other elements and no other
//     elements contain referents to conds, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementCond) Copy(_ map[string]Element) Element {
	return &ElementCond{
		traceID:                  this.traceID,
		index:                    this.index,
		routine:                  this.routine,
		tPre:                     this.tPre,
		tPost:                    this.tPost,
		id:                       this.id,
		op:                       this.op,
		file:                     this.file,
		line:                     this.line,
		vc:                       this.vc.Copy(),
		wVc:                      this.wVc.Copy(),
		numberConcurrent:         this.numberConcurrent,
		numberConcurrentWeak:     this.numberConcurrentWeak,
		numberConcurrentSame:     this.numberConcurrentSame,
		numberConcurrentWeakSame: this.numberConcurrentWeakSame,
	}
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
func (this *ElementCond) GetNumberConcurrent(weak, sameElem bool) int {
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
func (this *ElementCond) SetNumberConcurrent(c int, weak, sameElem bool) {
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
