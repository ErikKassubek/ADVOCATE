// Copyright (c) 2024 Erik Kassubek
//
// File: traceElementOnce.go
// Brief: Struct and functions for once operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-09-25
//
// License: BSD-3-Clause

package analysis

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"analyzer/clock"
)

// TraceElementOnce is a trace element for a once
// Fields:
//
//   - routine int: The routine id
//   - tPre int: The timestamp at the start of the event
//   - tPost int: The timestamp at the end of the event
//   - id int: The id of the mutex
//   - suc bool: Whether the operation was successful
//   - file (string), line int: The position of the mutex operation in the code
type TraceElementOnce struct {
	index   int
	routine int
	tPre    int
	tPost   int
	id      int
	suc     bool
	file    string
	line    int
	vc      *clock.VectorClock
	wVc     *clock.VectorClock
	rel1    []TraceElement
	rel2    []TraceElement
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
func AddTraceElementOnce(routine int, tPre string,
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

	file, line, err := posFromPosString(pos)
	if err != nil {
		return err
	}

	elem := TraceElementOnce{
		index:   numberElemsInTrace(routine),
		routine: routine,
		tPre:    tPreInt,
		tPost:   tPostInt,
		id:      idInt,
		suc:     sucBool,
		file:    file,
		line:    line,
		vc:      clock.NewVectorClock(MainTrace.numberOfRoutines),
		wVc:     clock.NewVectorClock(MainTrace.numberOfRoutines),
		rel1:    make([]TraceElement, 2),
		rel2:    make([]TraceElement, 0),
	}

	AddElementToTrace(&elem)

	return nil
}

// GetID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (on *TraceElementOnce) GetID() int {
	return on.id
}

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (on *TraceElementOnce) GetRoutine() int {
	return on.routine
}

// GetTPre returns the tPre of the element.
//
// Returns:
//   - int: The tPre of the element
func (on *TraceElementOnce) GetTPre() int {
	return on.tPre
}

// GetTPost returns the tPost of the element.
//
// Returns:
//   - int: The tPost of the element
func (on *TraceElementOnce) GetTPost() int {
	return on.tPost
}

// GetTSort returns the timer value, that is used for the sorting of the trace
//
// Returns:
//   - int: The timer of the element
func (on *TraceElementOnce) GetTSort() int {
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
func (on *TraceElementOnce) GetPos() string {
	return fmt.Sprintf("%s:%d", on.file, on.line)
}

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (on *TraceElementOnce) GetReplayID() string {
	return fmt.Sprintf("%d:%s:%d", on.routine, on.file, on.line)
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (on *TraceElementOnce) GetFile() string {
	return on.file
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (on *TraceElementOnce) GetLine() int {
	return on.line
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (on *TraceElementOnce) GetTID() string {
	return on.GetPos() + "@" + strconv.Itoa(on.tPre)
}

// GetVC returns the vector clock of the element
//
// Returns:
//   - VectorClock: The vector clock of the element
func (on *TraceElementOnce) GetVC() *clock.VectorClock {
	return on.vc
}

// GetWVc returns the weak vector clock of the element
//
// Returns:
//   - VectorClock: The weak vector clock of the element
func (on *TraceElementOnce) GetWVc() *clock.VectorClock {
	return on.wVc
}

// GetObjType returns the string representation of the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (on *TraceElementOnce) GetObjType(operation bool) string {
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
func (on *TraceElementOnce) GetSuc() bool {
	return on.suc
}

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (on *TraceElementOnce) IsEqual(elem TraceElement) bool {
	return on.routine == elem.GetRoutine() && on.ToString() == elem.ToString()
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (on *TraceElementOnce) GetTraceIndex() (int, int) {
	return on.routine, on.index
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (on *TraceElementOnce) SetT(time int) {
	on.tPre = time
	on.tPost = time
}

// SetTPre sets the tPre of the element.
//
// Parameter:
//   - tPre int: The tPre of the element
func (on *TraceElementOnce) SetTPre(tPre int) {
	on.tPre = tPre
	if on.tPost != 0 && on.tPost < tPre {
		on.tPost = tPre
	}
}

// SetTSort sets the timer, that is used for the sorting of the trace
//
// Parameter:
//   - tSort int: The timer of the element
func (on *TraceElementOnce) SetTSort(tSort int) {
	on.SetTPre(tSort)
	on.tPost = tSort
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (on *TraceElementOnce) SetTWithoutNotExecuted(tSort int) {
	on.SetTPre(tSort)
	if on.tPost != 0 {
		on.tPost = tSort
	}
}

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (on *TraceElementOnce) ToString() string {
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

// Update the vector clock of the trace and element
func (on *TraceElementOnce) updateVectorClock() {
	on.vc = currentVC[on.routine].Copy()
	on.wVc = currentVC[on.routine].Copy()

	if on.suc {
		DoSuc(on)
	} else {
		DoFail(on)
	}

}

// Copy the element
//
// Returns:
//   - TraceElement: The copy of the element
func (on *TraceElementOnce) Copy() TraceElement {
	return &TraceElementOnce{
		index:   on.index,
		routine: on.routine,
		tPre:    on.tPre,
		tPost:   on.tPost,
		id:      on.id,
		suc:     on.suc,
		file:    on.file,
		line:    on.line,
		vc:      on.vc.Copy(),
		wVc:     on.wVc.Copy(),
		rel1:    on.rel1,
		rel2:    on.rel1,
	}
}

// ========= For GoPie fuzzing ===========

// AddRel1 adds an element to the rel1 set of the element
//
// Parameter:
//
//	elem TraceElement: elem to add
//	pos int: before (0) or after (1)
func (on *TraceElementOnce) AddRel1(elem TraceElement, pos int) {
	if pos < 0 || pos > 1 {
		return
	}
	on.rel1[pos] = elem
}

// AddRel2 adds an element to the rel2 set of the element
//
// Parameter:
//   - elem TraceElement: elem to add
func (on *TraceElementOnce) AddRel2(elem TraceElement) {
	on.rel2 = append(on.rel2, elem)
}

// GetRel1 returns the rel1 set
//
// Returns:
//   - []*TraceElement: the rel1 set
func (on *TraceElementOnce) GetRel1() []TraceElement {
	return on.rel1
}

// GetRel2 returns the rel2 set
//
// Returns:
//   - []*TraceElement: the rel2 set
func (on *TraceElementOnce) GetRel2() []TraceElement {
	return on.rel2
}
