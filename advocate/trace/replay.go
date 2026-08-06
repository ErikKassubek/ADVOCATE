// Copyright (c) 2024 Erik Kassubek
//
// File: /advocate/trace/replay.go
// Brief: Struct and functions for replay control elements in the trace
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package trace

import (
	"advocate/analysis/hb/a_clock"
	"fmt"
	"strconv"
)

// ========================================================
// MARK: Data
// ========================================================

// ElementReplay is a struct to save an end of replay marker in the trace
// Fields:
//   - t int: The timestamp of the event
//   - exitCode int: expected exit code
type ElementReplay struct {
	ElementBase

	t        int
	exitCode int
}

// ========================================================
// MARK: Constructor
// ========================================================

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
		ElementBase: this.newElementBase(0),
		t:           ts,
		exitCode:    exitCode,
	}

	this.AddElement(&elem)

	return nil
}

// ========================================================
// MARK: Timestamps
// ========================================================

// GetTPre returns the t of the element.
//
//   - int: The tPost of the element
func (this *ElementReplay) T(_ timeType) int {
	return this.t
}

// SetT sets the tPre and tPost of the element
//
// Parameter:
//   - time int: The tPre and tPost of the element
func (this *ElementReplay) SetT(_ timeType, time int) {
	this.t = time
}

// SetTWithoutNotExecuted set the timer, that is used for the sorting of the trace, only if the original
// value was not 0
//
// Parameter:
//   - tSort int: The timer of the element
func (this *ElementReplay) SetTWithoutNotExecuted(tSort int) {
	tSort = max(1, tSort)
	this.SetT(Request, tSort)
	this.t = tSort
}

// Committed returns if the operation was committed (tPost != 0)
//
// Returns:
//   - bool: true if committed, false if not
func (this *ElementReplay) Committed() bool {
	return true
}

// ========================================================
// MARK: Position
// ========================================================

// Pos returns the position of the operation in the form [file]:[line].
//
// Returns:
//   - position: the position
func (this *ElementReplay) Pos() Position {
	return newPosition("", 0)
}

// File returns the file of the element
//
// Returns:
//   - The file of the element
func (this *ElementReplay) File() string {
	return ""
}

// Line returns the line of the element
//
// Returns:
//   - The line of the element
func (this *ElementReplay) Line() int {
	return 0
}

// ========================================================
// MARK: Index
// ========================================================

// Routine returns the routine ID of the element.
//
// Returns:
//   - int: The routine of the element
func (this *ElementReplay) Routine() int {
	return 1
}

// TraceIndex returns the trace local index of the element in the trace
//
// Returns:
//   - int: the routine id of the element
//   - int: The trace local index of the element in the trace
func (this *ElementReplay) TraceIndex() (int, int) {
	return -1, -1
}

// ObjID returns the ID of the primitive on which the operation was executed
//
// Returns:
//   - int: The id of the element
func (this *ElementReplay) ObjID() int {
	return 0
}

// ========================================================
// MARK: Operation
// ========================================================

// Type returns the object type
//
// Parameter:
//   - operation bool: if true get the operation code, otherwise only the primitive code
//
// Returns:
//   - ObjectType: the object type
func (this *ElementReplay) Type(operation bool) OperationType {
	if !operation {
		return Replay
	}
	return ReplayOP
}

// ========================================================
// MARK: Equal
// ========================================================

// IsEqual checks if an trace element is equal to this element
//
// Parameter:
//   - elem TraceElement: The element to check against
//
// Returns:
//   - bool: true if it is the same operation, false otherwise
func (this *ElementReplay) IsEqual(elem Element) bool {
	return this.id == elem.ID()
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
// MARK: String
// ========================================================

// String returns the simple string representation of the element.
//
// Returns:
//   - string: The simple string representation of the element
func (this *ElementReplay) String() string {
	res := "X," + strconv.Itoa(this.t) + "," + strconv.Itoa(this.exitCode)
	return res
}

// String returns the simple string representation of the element with leading routine
//
// Returns:
//   - string: The simple string representation of the element with leading routine
func (this *ElementReplay) StringDebug() string {
	routine := fmt.Sprintf("%4d", this.Routine())
	if this.ElementBase.init {
		routine = "   *"
	}
	return fmt.Sprintf("%s@%s", routine, this.String())
}

// ========================================================
// MARK: VC
// ========================================================

// Vc is a dummy function to implement the TraceElement interface
func (this *ElementReplay) Vc(_ a_clock.VcType, _ *a_clock.VectorClock) {
}

// GetVC is a dummy function to implement the TraceElement interface
func (this *ElementReplay) GetVC(_ a_clock.VcType) *a_clock.VectorClock {
	return &a_clock.VectorClock{}
}

// ========================================================
// MARK: Function
// ========================================================

func (this *ElementReplay) Function() *ElementFunc {
	return nil
}

// ========================================================
// MARK: Concurrent
// ========================================================

// NumberConcurrent returns the number of elements concurrent to the element
// If not set, it returns -1
//
// Returns:
//   - int: -1
func (this *ElementReplay) NumberConcurrent(_, _ bool) int {
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
// MARK: Replay
// ========================================================

// ReplayID returns the replay id of the element
//
// Returns:
//   - The replay id
func (this *ElementReplay) ReplayID() string {
	return ""
}

// ========================================================
// MARK: Copy
// ========================================================

// Copy creates a copy of the element
//
// Parameter:
//   - _ map[string]Element: map containing all already copied elements.
//   - keep bool: if true, keep vc and order information
//
// Returns:
//   - TraceElement: The copy of the element
func (this *ElementReplay) Copy(_ map[int]Element, keep bool) Element {
	if !keep {
		return &ElementReplay{
			ElementBase: this.ElementBase.Copy(),
			t:           0,
			exitCode:    this.exitCode,
		}
	}
	return &ElementReplay{
		ElementBase: this.ElementBase.Copy(),
		t:           this.t,
		exitCode:    this.exitCode,
	}
}

// ========================================================
// MARK: Valid
// ========================================================

func (this *ElementReplay) IsValid() bool {
	return this != nil
}
