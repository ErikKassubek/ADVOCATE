// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/replay.go
// Brief: Struct and functions for replay control elements in the trace
//
// Author: Erik Kassubek
// Created: 2024-04-03
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"strconv"
)

// ElementReplay is a struct to save an end of replay marker in the trace
// Fields:
//   - id: id of the element, should never be changed
//   - tPost int: The timestamp of the event
//   - exitCode int: expected exit code
type ElementReplay struct {
	id       int
	tPost    int
	exitCode int
}

// AddTraceElementReplay adds an replay end element to a trace
//
// Parameter:
//   - ts string: The timestamp of the event
//   - exitCode int: The exit code of the event
//
// Returns:
//   - error
func (this *Trace) AddTraceElementReplay(ts int, exitCode int) error {
	elem := ElementReplay{
		tPost:    ts,
		exitCode: exitCode,
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
func (this *ElementReplay) GetID() int {
	return this.id
}

// GetTraceID sets the trace id
//
// Parameter:
//   - ID int: the trace id
func (this *ElementReplay) setID(ID int) {
	this.id = ID
}

// ========================================================
// Timestamps
// ========================================================

// GetTPre returns the t of the element.
//
//   - int: The tPost of the element
func (this *ElementReplay) GetT(_ timeType) int {
	return this.tPost
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementReplay) SetT(_ timeType, time int) {
	this.tPost = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	this.SetT(Request, tSort)
	this.tPost = tSort
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementReplay) Committed() bool {
	return true
}

// ========================================================
// Position
// ========================================================

// GetPos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - string: The file of the element
func (this *ElementReplay) GetPos() string {
	return ""
}

// GetFile returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementReplay) GetFile() string {
	return ""
}

// GetLine returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementReplay) GetLine() int {
	return 0
}

// ========================================================
// Index
// ========================================================

// GetRoutine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementReplay) GetRoutine() int {
	return 1
}

// GetTraceIndex returns the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementReplay) GetTraceIndex() (int, int) {
	return -1, -1
}

// GetObjId returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementReplay) GetObjId() int {
	return 0
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
//   - ObjectType: the object type
func (this *ElementReplay) GetType(operation bool) OperationType {
	if !operation {
		return Replay
	}
	return ReplayOP
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
func (this *ElementReplay) IsEqual(elem Element) bool {
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
func (this *ElementReplay) IsSameElement(elem Element) bool {
	return false
}

// ========================================================
// String
// ========================================================

// ToString returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementReplay) ToString() string {
	res := "X," + strconv.Itoa(this.tPost) + "," + strconv.Itoa(this.exitCode)
	return res
}

// GetTID returns the tID of the element.
// The tID is normally a string of form [file]:[line]@[tPre]
// Since the replay element is not used for any analysis, it returns an empty string
//
// Returns:
//   - string: The tID of the element
func (this *ElementReplay) GetTID() string {
	return ""
}

// ========================================================
// VC
// ========================================================

// SetVc is a dummy function to implement the TraceElement interface
func (this *ElementReplay) SetVc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

// GetVC is a dummy function to implement the TraceElement interface
func (this *ElementReplay) GetVC(_ a_clock.VcType) *a_clock.VectorClock {
	return &a_clock.VectorClock{}
}

// ========================================================
// Concurrent
// ========================================================

// GetNumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - int: -1
func (this *ElementReplay) GetNumberConcurrent(_, _ bool) int {
	return -1
}

// SetNumberConcurrent sets the number of concurrent elements
func (this *ElementReplay) SetNumberConcurrent(_ int, _, _ bool) {}

// GetConcurrent returns the elements that are concurrent to the element
//
// Parameter:
//   - weak bool: get number of weak concurrent
//
// Returns:
//   - []Element: empty
func (this *ElementReplay) GetConcurrent(_, _ bool) []Element {
	return []Element{}
}

// SetConcurrent sets the concurrent elements
func (this *ElementReplay) SetConcurrent(_ []Element, _, _ bool) {
}

// ========================================================
// Replay
// ========================================================

// GetReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementReplay) GetReplayID() string {
	return ""
}

// ========================================================
// Copy
// ========================================================

// Copy creates a copy of the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementReplay) Copy(_ map[string]Element, keep bool) Element {
	if !keep {
		return &ElementReplay{
			id:       this.id,
			tPost:    0,
			exitCode: this.exitCode,
		}
	}
	return &ElementReplay{
		id:       this.id,
		tPost:    this.tPost,
		exitCode: this.exitCode,
	}
}

// ========================================================
// Valid
// ========================================================

func (this *ElementReplay) IsValid() bool {
	return this != nil
}
