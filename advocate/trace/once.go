// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementOnce.go
// Brief: Struct and functions for once operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-09-25
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

// ElementOnce is a trace element for a once
// Fields:
//   - traceID: id of the element, should never be changed
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the mutex
//   - suc bool: Whether the operation was successful
//   - file (string), line int: The position of the mutex operation in the code
//   - vc *clock.VectorClock: the vector clock of the element
//   - wVc *clock.VectorClock: the weak vector clock of the element
//   - numberConcurrent: number of concurrent elements in the trace, -1 if not calculated
//   - numberConcurrentSame int: number of concurrent elements in the trace on the same element, -1 if not calculated
//   - numberConcurrentWeakSame int: number of weak concurrent elements in the trace on the same element, -1 if not calculated
type ElementOnce struct {
	traceID                  int
	index                    int
	routine                  int
	tPre                     int
	tPost                    int
	id                       int
	suc                      bool
	file                     string
	line                     int
	vc                       *clock.VectorClock
	wVc                      *clock.VectorClock
	numberConcurrent         int
	numberConcurrentWeak     int
	numberConcurrentSame     int
	numberConcurrentWeakSame int
}

// AddTraceElementOnce adds a new mutex trace element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPre string: The timestamp at the start of the event
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the mutex
//   - suc string: Whether the operation was successful (only for trylock else always true)
//   - pos string: The position of the mutex operation in the code
func (t *Trace) AddTraceElementOnce(routine int, tPre string,
	tPost string, id string, suc string, pos string) error {
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

	sucBool, err := strconv.ParseBool(suc)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	file, line, err := PosFromPosString(pos)
	if err != nil {
		return err
	}

	elem := ElementOnce{
		index:                    t.numberElemsInTrace[routine],
		routine:                  routine,
		tPre:                     tPreInt,
		tPost:                    tPostInt,
		id:                       idInt,
		suc:                      sucBool,
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
func (on *ElementOnce) GetID() int {
	return on.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (on *ElementOnce) GetRoutine() int {
	return on.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (on *ElementOnce) GetTPre() int {
	return on.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (on *ElementOnce) GetTPost() int {
	return on.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (on *ElementOnce) GetTSort() int {
	if on.tPost == 0 {
		// add at the end of the trace
		return math.MaxInt
	}
	return on.tPre
}

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The position of the element
func (on *ElementOnce) GetPos() string {
	return fmt.Sprintf("%s:%d", on.file, on.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (on *ElementOnce) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", on.routine, on.file, on.line)
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (on *ElementOnce) GetFile() string {
	return on.file
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (on *ElementOnce) GetLine() int {
	return on.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (on *ElementOnce) GetTID() string {
	return on.GetPos() + "@" + strconv.Itoa(on.tPre)
}

// SetVc sets the vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (on *ElementOnce) SetVc(vc *clock.VectorClock) {
	on.vc = vc.Copy()
}

// SetWVc sets the weak vector clock
//
// Parameter:
//   - vc *clock.VectorClock: the vector clock
func (on *ElementOnce) SetWVc(vc *clock.VectorClock) {
	on.wVc = vc.Copy()
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (on *ElementOnce) GetVC() *clock.VectorClock {
	return on.vc
}

// GetWVC returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The weak vector clock of the element
func (on *ElementOnce) GetWVC() *clock.VectorClock {
	return on.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (on *ElementOnce) GetObjType(operation bool) string {
	if !operation {
		return ObjectTypeOnce
	}

	if on.suc {
		return ObjectTypeOnce + "E"
	}
	return ObjectTypeOnce + "N"
}

// GetSuc returns whether the once do was executed (successful)
//
// Returns:
//   - bool: true if function in Do was executed, false otherwise
func (on *ElementOnce) GetSuc() bool {
	return on.suc
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (on *ElementOnce) IsEqual(elem Element) bool {
	return on.routine == elem.GetRoutine() && on.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (on *ElementOnce) GetTraceIndex() (int, int) {
	return on.routine, on.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (on *ElementOnce) SetT(time int) {
	on.tPre = time
	on.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (on *ElementOnce) SetTPre(tPre int) {
	on.tPre = tPre
	if on.tPost != 0 && on.tPost < tPre {
		on.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (on *ElementOnce) SetTSort(tSort int) {
	on.SetTPre(tSort)
	on.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (on *ElementOnce) SetTWithoutNotExecuted(tSort int) {
	on.SetTPre(tSort)
	if on.tPost != 0 {
		on.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (on *ElementOnce) ToString() string {
	res := "O,"
	res += strconv.Itoa(on.tPre) + ","
	res += strconv.Itoa(on.tPost) + ","
	res += strconv.Itoa(on.id) + ","
	if on.suc {
		res += "t"
	} else {
		res += "f"
	}
	res += "," + on.GetPos()
	return res
}

// GetTraceID returns the trace id
//
// Returns:
//   - int: the trace id
func (on *ElementOnce) GetTraceID() int {
	return on.traceID
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (on *ElementOnce) setTraceID(ID int) {
	on.traceID = ID
}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (on *ElementOnce) Copy() Element {
	return &ElementOnce{
		traceID:                  on.traceID,
		index:                    on.index,
		routine:                  on.routine,
		tPre:                     on.tPre,
		tPost:                    on.tPost,
		id:                       on.id,
		suc:                      on.suc,
		file:                     on.file,
		line:                     on.line,
		vc:                       on.vc.Copy(),
		wVc:                      on.wVc.Copy(),
		numberConcurrent:         on.numberConcurrent,
		numberConcurrentWeak:     on.numberConcurrentWeak,
		numberConcurrentSame:     on.numberConcurrentSame,
		numberConcurrentWeakSame: on.numberConcurrentWeakSame,
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
func (on *ElementOnce) GetNumberConcurrent(weak, sameElem bool) int {
	if weak {
		if sameElem {
			return on.numberConcurrentWeakSame
		}
		return on.numberConcurrentWeak
	}
	if sameElem {
		return on.numberConcurrentSame
	}
	return on.numberConcurrent
}

// SetNumberConcurrent sets the number of concurrent elements
//
// Parameter:
//   - c int: the number of concurrent elements
//   - weak bool: return number of weak concurrent
//   - sameElem bool: only operation on the same variable
func (on *ElementOnce) SetNumberConcurrent(c int, weak, sameElem bool) {
	if weak {
		if sameElem {
			on.numberConcurrentWeakSame = c
		} else {
			on.numberConcurrentWeak = c
		}
	} else {
		if sameElem {
			on.numberConcurrentSame = c
		} else {
			on.numberConcurrent = c
		}
	}
}
