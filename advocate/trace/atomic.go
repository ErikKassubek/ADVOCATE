// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementAtomic.go
// Brief: Struct and functions for atomic operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/clock"
	"advocate/utils/log"
	"errors"
	"fmt"
	"strconv"
)

// ElementAtomic is a struct to save an atomic event in the trace
// Fields:
//
//   - traceID: id of the element, should never be changed
//   - index int: index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp of the event
//   - id int: The id of the atomic variable
//   - op ObjectType: The operation on the atomic variable
//   - vc *clock.VectorClock: The vector clock of the operation
//   - wVc *clock.VectorClock: The weak vector clock of the operation
//   - file string: the file of the operation
//   - line int: the line of the operation
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentWeak: number of weak concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementAtomic struct {
	traceID                  int
	index                    int
	routine                  int
	tPost                    int
	id                       int
	op                       ObjectType
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	file                     string
	line                     int
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementAtomic adds a new atomic trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp of the event
//   - id string: The id of the atomic variable
//   - operation string: The operation on the atomic variable
//   - pos string: The position of the atomic
func (t Trace) AddTraceElementAtomic(routine int, tPost string,
	id string, operation string, pos string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPost is not an integer")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	var opAInt ObjectType
	switch operation {
	case "L":
		opAInt = AtomicLoad
	case "S":
		opAInt = AtomicStore
	case "A":
		opAInt = AtomicAdd
	case "W":
		opAInt = AtomicSwap
	case "C":
		opAInt = AtomicCompAndSwap
	case "N":
		opAInt = AtomicAnd
	case "O":
		opAInt = AtomicOr
	default:
		return fmt.Errorf("Atomic operation '%s' is not a valid operation", operation)
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		log.Error("Cannot read pos string ", pos)
		return err
	}

	elem := ElementAtomic{
		index:                    t.numberElemsInTrace[routine],
		routine:                  routine,
		tPost:                    tPostInt,
		id:                       idInt,
		op:                       opAInt,
		file:                     file,
		line:                     line,
		vc:                       nil,
		wVc:                      nil,
		numberConcurrent:         -1,
		numberConcurrentWeak:     -1,
		numberConcurrentSame:     -1,
		numberConcurrentWeakSame: -1,
	}

	t.AddElement(&elem)
	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (at *ElementAtomic) GetID() int {
	return at.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (at *ElementAtomic) GetRoutine() int {
	return at.routine
}

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (at *ElementAtomic) GetTPre() int {
	return at.tPost
}

// GetTPost returns the tPost of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPost of the element
func (at *ElementAtomic) GetTPost() int {
	return at.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (at *ElementAtomic) GetTSort() int {
	return at.tPost
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (at *ElementAtomic) GetPos() string {
	return fmt.Sprintf("%s:%d", at.file, at.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (at *ElementAtomic) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", at.routine, at.file, at.line)
}

// GetFile returns the file where the operation represented by the element was executed
//
// Returns:
//   - The file of the element
func (at *ElementAtomic) GetFile() string {
	return at.file
}

// GetLine returns the line where the operation represented by the element was executed
//
// Returns:
//   - The line of the element
func (at *ElementAtomic) GetLine() int {
	return at.line
}

// GetTID returns the tID of the element.
// The tID is a string of form A@[file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (at *ElementAtomic) GetTID() string {
	return "A@" + at.GetPos() + "@" + strconv.Itoa(at.tPost)
}

// SetVc sets the vector clock
//
// Parameter:
//   - cl *clock.VectorClock: the vector clock
func (at *ElementAtomic) SetVc(cl *clock.VectorClock) {
	at.vc = cl.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - cl *clock.VectorClock: the vector clock
func (at *ElementAtomic) SetWVc(cl *clock.VectorClock) {
	at.wVc = cl.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (at *ElementAtomic) GetVC() *clock.VectorClock {
	return at.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The weak vector clock of the element
func (at *ElementAtomic) GetWVC() *clock.VectorClock {
	return at.wVc
}

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (at *ElementAtomic) GetType(operation bool) ObjectType {
	if !operation {
		return Atomic
	}

	return at.op
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (at *ElementAtomic) IsEqual(elem Element) bool {
	return at.routine == elem.GetRoutine() && at.ToString() == elem.ToString()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: true if at and elem are operations on the same atomic variable
func (at *ElementAtomic) IsSameElement(elem Element) bool {
	if elem.GetType(false) != Atomic {
		return false
	}

	return at.id == elem.GetID()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (at *ElementAtomic) GetTraceIndex() (int, int) {
	return at.routine, at.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (at *ElementAtomic) SetT(time int) {
	at.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPost of the element
func (at *ElementAtomic) SetTPre(tPre int) {
	at.tPost = tPre
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (at *ElementAtomic) SetTSort(tSort int) {
	at.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (at *ElementAtomic) SetTWithoutNotExecuted(tSort int) {
	if at.tPost != 0 {
		at.tPost = tSort
	}
}

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (at *ElementAtomic) ToString() string {
	opString := string(at.op)[1]

	return fmt.Sprintf("A,%d,%d,%s,%s", at.tPost, at.id, opString, at.GetPos())
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (at *ElementAtomic) GetTraceID() int {
	return at.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (at *ElementAtomic) setTraceID(ID int) {
	at.traceID = ID
}

// Copy the atomic element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//     since atomics do not contain reference to other elements and no other
//     elements contain referents to atomics, this is not used
//
// Returns:
//   - TraceElement: The copy of the element
func (at *ElementAtomic) Copy(_ map[string]Element) Element {

	return &ElementAtomic{
		traceID:                  at.traceID,
		index:                    at.index,
		routine:                  at.routine,
		tPost:                    at.tPost,
		id:                       at.id,
		op:                       at.op,
		vc:                       at.vc.Copy(),
		wVc:                      at.wVc.Copy(),
		numberConcurrent:         at.numberConcurrent,
		numberConcurrentWeak:     at.numberConcurrentWeak,
		numberConcurrentSame:     at.numberConcurrentSame,
		numberConcurrentWeakSame: at.numberConcurrentWeakSame,
		file:                     at.file,
		line:                     at.line,
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
func (at *ElementAtomic) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return at.numberConcurrentWeakSame
		}
		return at.numberConcurrentWeak
	}
	if sameElem {
		return at.numberConcurrentSame
	}
	return at.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (at *ElementAtomic) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			at.numberConcurrentWeakSame = c
		} else {
			at.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			at.numberConcurrentSame = c
		} else {
			at.numberConcurrent = c
		}
	}
}
