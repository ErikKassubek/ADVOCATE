// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/routineEnd.go
// Brief: Struct and functions for fork operations in the trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"errors"
	"strconv"
)

// ElementRoutineEnd is a trace element for the termination of a routine end
// Fields:
//   - id: id of the element, should never be changed
//   - index int: Index in the routine
//   - routine int: The routine id
//   - tPost int: The timestamp at the end of the event
//   - vc clock.VectorClock: The vector clock
type ElementRoutineEnd struct {
	id      int
	index   int
	routine int
	tPost   int
	vc      *a_clock.VectorClock
	wVc     *a_clock.VectorClock
}

// AddTraceElementRoutineEnd add a routine and element to the main trace
//
// Parameter:
//   - routine int: The routine id
//   - tPost string: The timestamp at the end of the event
//   - id string: The id of the new routine
//   - pos string: The position of the trace element in the file
func (this *Trace) AddTraceElementRoutineEnd(routine int, tPost string) error {
	tPostInt, err := strconv.Atoi(tPost)
	if err != nil {
		return errors.New("tPre is not an integer")
	}

	elem := ElementRoutineEnd{
		index:   this.NumberElemInRoutine(routine),
		routine: routine,
		tPost:   tPostInt,
		vc:      nil,
		wVc:     nil,
	}

	this.AddElement(&elem)

	return nil
}

// ========================================================
// ID
// ========================================================

// GetID returns the trace id
//
// Returns:
//   - int: the trace id
func (this *ElementRoutineEnd) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementRoutineEnd) setID(ID int) {
	this.id = ID
}

// GetObjId is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (this *ElementRoutineEnd) GetObjId() int {
	return 0
}

// ========================================================
// Timestamps
// ========================================================

// GetTPre returns the tPre of the element. For atomic elements, tPre and tPost are the same
//
// Returns:
//   - int: The tPre of the element
func (this *ElementRoutineEnd) GetT(_ timeType) int {
	return this.tPost
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementRoutineEnd) SetT(_ timeType, time int) {
	this.tPost = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementRoutineEnd) SetTWithoutNotExecuted(tSort int) {
	this.SetT(Request, tSort)
	if this.tPost != 0 {
		this.tPost = tSort
	}
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementRoutineEnd) Committed() bool {
	return true
}

// ========================================================
// Position
// ========================================================

// GetPos is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) GetPos() string {
	return ""
}

// GetFile is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) GetFile() string {
	return ""
}

// GetLine is a dummy function to implement the traceElement interface
//
// Returns:
//   - int: 0
func (this *ElementRoutineEnd) GetLine() int {
	return 0
}

// ========================================================
// Index
// ========================================================

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementRoutineEnd) GetRoutine() int {
	return this.routine
}

// GetTraceIndex returns trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementRoutineEnd) GetTraceIndex() (int, int) {
	return this.routine, this.index
}

// ========================================================
// Operation
// ========================================================

// GetType returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - string: the object type
func (this *ElementRoutineEnd) GetType(operation bool) OperationType {
	if !operation {
		return End
	}
	return EndRoutine
}

// ========================================================
// Equal
// ========================================================

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementRoutineEnd) IsEqual(elem Element) bool {
	return this.id == elem.GetID()
}

// IsSameElement returns checks if the element on which the at and elem
// where performed are the same
//
// Parameter:
//   - elem Element: the element to compare against
//
// Returns:
//   - bool: always false
func (this *ElementRoutineEnd) IsSameElement(elem Element) bool {
	return false
}

// ========================================================
// String
// ========================================================

// ToString returns the simple string representation of the element
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementRoutineEnd) ToString() string {
	return "E" + "," + strconv.Itoa(this.tPost)
}

// GetTID returns the tID of the element.
// The tID is a string of form [file]:[line]@[tPre]
//
// Returns:
//   - string: The tID of the element
func (this *ElementRoutineEnd) GetTID() string {
	return ""
}

// ========================================================
// VC
// ========================================================

// SetVc sets the vector clock
//
// Parameter:
//   - weak bool
//   - vc *clock.VectorClock: the vector clock
func (this *ElementRoutineEnd) SetVc(weak a_clock.VcType, vc *a_clock.VectorClock) {
	if weak == a_clock.Weak {
		this.wVc = vc.Copy()
	} else {
		this.vc = vc.Copy()
	}
}

// GetVC returns the vector clock of the element
//
// Parameter:
//   - weak bool
//
// Returns:
//   - VectorClock: The vector clock of the element
func (this *ElementRoutineEnd) GetVC(weak a_clock.VcType) *a_clock.VectorClock {
	if weak == a_clock.Weak {
		return this.wVc
	}
	return this.vc
}

// ========================================================
// Concurrent
// ========================================================

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
func (this *ElementRoutineEnd) GetNumberConcurrent(_, _ bool) int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
func (this *ElementRoutineEnd) SetNumberConcurrent(_ int, _, _ bool) {}

// GetConcurrent returns the elements that are concurrent to the element
func (this *ElementRoutineEnd) GetConcurrent(_, _ bool) []Element {
	return []Element{}
}

// SetConcurrent sets the concurrent elements
func (this *ElementRoutineEnd) SetConcurrent(_ []Element, _, _ bool) {
}

// ========================================================
// Replay
// ========================================================

// GetReplayID is a dummy function to implement the traceElement interface
//
// Returns:
//   - string: empty string
func (this *ElementRoutineEnd) GetReplayID() string {
	return ""
}

// ========================================================
// Copy
// ========================================================

// Copy the element
//
// Parameter:
//   - mapping map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementRoutineEnd) Copy(mapping map[string]Element, keep bool) Element {
	if !keep {
		return &ElementRoutineEnd{
			id:      this.id,
			index:   0,
			routine: this.routine,
			tPost:   0,
			vc:      nil,
			wVc:     nil,
		}
	}

	return &ElementRoutineEnd{
		id:      this.id,
		index:   this.index,
		routine: this.routine,
		tPost:   this.tPost,
		vc:      this.vc.Copy(),
		wVc:     this.wVc.Copy(),
	}
}

// ========================================================
// Valid
// ========================================================

func (this *ElementRoutineEnd) IsValid() bool {
	return this != nil
}
