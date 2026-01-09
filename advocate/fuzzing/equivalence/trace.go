// Copyright (c) 2025 Erik Kassubek
//
// File: traceMin.go
// Brief: Functions and structs to implement a minimal representation of a trace
//
// Author: Erik Kassubek
// Created: 2025-10-22
//
// License: BSD-3-Clause

package equivalence

import (
	"advocate/analysis/baseA"
	"advocate/fuzzing/baseF"
	"advocate/trace"
	"advocate/utils/types"
	"math"
)

type TraceEq struct {
	trace []trace.Element

	illFormedBug        bool
	IllFormedImpossible bool
	vcHaveBeenCalc      bool

	minT      int
	closed    map[int]struct{} // channel id
	qCount    map[int]int      // channel id -> send-recv
	qMessage  map[int]*types.Queue[*trace.ElementChannel]
	critSec   map[int]types.Pair[bool, int] // mutex id -> is rw lock, number of currently hold
	onceDo    map[int]struct{}              // once id
	wgCounter map[int]int                   // wg counter
	condVal   map[int]int                   // wg counter, number signal - release, 0.5*maxInt if broadcast

	fullSig string
}

// NewTraceEq creates a new, empty trace
//
// Returns:
//   - TraceMin: a new, empty trace
func NewTraceEq() TraceEq {
	return TraceEq{
		trace: make([]trace.Element, 0),

		illFormedBug:        false,
		IllFormedImpossible: false,

		minT:     0,
		closed:   make(map[int]struct{}),
		qCount:   make(map[int]int),
		qMessage: make(map[int]*types.Queue[*trace.ElementChannel]),

		critSec:   make(map[int]types.Pair[bool, int]),
		onceDo:    make(map[int]struct{}),
		wgCounter: make(map[int]int),
		condVal:   make(map[int]int),
	}
}

// Trace returns the elements in the TraceMin
//
// Returns:
//   - int: the elements in the trace
func (this *TraceEq) Trace() []trace.Element {
	return this.trace
}

// Len returns the number of elements in the TraceMin
//
// Returns:
//   - int: the number of elements in the trace
func (this *TraceEq) Len() int {
	return len(this.trace)
}

// Create a copy of the TraceMin
//
// Returns:
//   - TraceMin: the copy
func (this *TraceEq) Clone() TraceEq {
	copy := make([]trace.Element, 0)
	for i := 0; i < len(this.trace); i++ {
		copy = append(copy, this.trace[i])
	}
	return TraceEq{
		trace: copy,
	}
}

// Create a subtrace of the TraceMin
//
// Parameter:
//   - start int: start index
//   - end int: end index
//
// Returns:
//   - TraceMin: the trace min with subtrace
func (this *TraceEq) CloneSub(start, end int) TraceEq {
	copy := this.trace[start:end]
	return TraceEq{
		trace: copy,
	}
}

// Swap two elements in the TraceMin
//
// Parameter:
//   - i: first index
//   - j: second index
func (this *TraceEq) Flip(i, j int) {
	this.trace[i], this.trace[j] = this.trace[j], this.trace[i]
}

// AddElem adds an elem to the min trace
//
// Parameter:
//   - elem ElemMin: the element to add
func (this *TraceEq) AddElem(elem trace.Element) {
	// objID := elem.GetID()

	this.minT++
	elem.SetT(this.minT)

	objId := elem.GetObjId()

	switch e := elem.(type) {
	case *trace.ElementChannel:
		switch e.GetType(true) {
		case trace.ChannelClose:
			// close on closed
			if _, ok := this.closed[objId]; ok {
				this.illFormedBug = true
			}

			this.closed[objId] = struct{}{}
		case trace.ChannelSend:
			// send on closed
			if _, ok := this.closed[objId]; ok {
				this.illFormedBug = true
			}

			this.qCount[objId]++
			if _, ok := this.qMessage[objId]; !ok {
				this.qMessage[objId] = types.NewQueue[*trace.ElementChannel]()
			}
			this.qMessage[objId].Push(e)
		case trace.ChannelRecv:
			this.qCount[objId]--

			// recv before send
			if this.qCount[objId] < 0 {
				this.IllFormedImpossible = true
			}

			if _, ok := this.qMessage[objId]; !ok {
				this.qMessage[objId] = types.NewQueue[*trace.ElementChannel]()
			}
			m := this.qMessage[objId].Pop()
			if m != nil {
				e.SetPartner(m)
				m.SetPartner(e)
			}
		}
		_, ok := this.closed[objId]
		e.SetClosed(ok)
		e.SetQCount(this.qCount[objId])
	case *trace.ElementMutex:
		switch e.GetType(true) {
		case trace.MutexLock:
			// double lock
			if _, ok := this.critSec[objId]; ok {
				this.illFormedBug = true
			}

			this.critSec[objId] = types.NewPair(false, 1)
		case trace.MutexRLock:
			val, ok := this.critSec[objId]

			// lock followed by rlock
			if ok && !val.X {
				this.illFormedBug = true
				break
			}

			if ok {
				this.critSec[objId] = types.NewPair(true, val.Y+1)
			} else {
				this.critSec[objId] = types.NewPair(true, 1)
			}
		case trace.MutexTryLock:
			if _, ok := this.critSec[objId]; ok {
				e.SetSuc(false)
			} else {
				this.critSec[objId] = types.NewPair(false, 1)
			}
		case trace.MutexTryRLock:
			if v, ok := this.critSec[objId]; !ok || v.X {
				this.critSec[objId] = types.NewPair(true, v.Y+1)
			} else {
				e.SetSuc(false)
			}
		case trace.MutexUnlock:
			val, ok := this.critSec[objId]

			// unlock without lock
			if !ok {
				this.illFormedBug = true
				break
			}

			// unlock of rlocked
			if val.X {
				this.illFormedBug = true
				break
			}

			delete(this.critSec, objId)
		case trace.MutexRUnlock:
			val, ok := this.critSec[objId]

			// unlock without lock
			if !ok {
				this.illFormedBug = true
				break
			}

			// runlock of locked
			if !val.X {
				this.illFormedBug = true
				break
			}

			newVal := val.Y - 1
			if newVal <= 0 {
				delete(this.critSec, objId)
			} else {
				this.critSec[objId] = types.NewPair(true, newVal)
			}

		}
	case *trace.ElementOnce:
		if _, ok := this.onceDo[objId]; !ok {
			e.SetSuc(true)
			this.onceDo[objId] = struct{}{}
		}
	case *trace.ElementSelect:
		cc := e.GetChosenCase()
		this.AddElem(cc)
	case *trace.ElementWait:
		if e.GetType(true) == trace.WaitWait {
			// wait release with non 0 value
			if val, ok := this.wgCounter[objId]; !ok || val != 0 {
				this.IllFormedImpossible = true
			}
			break
		}

		this.wgCounter[objId] += e.GetDelta()

		// negative wg
		if this.wgCounter[objId] < 0 {
			this.illFormedBug = true
		}

		e.SetVal(this.wgCounter[objId])
	case *trace.ElementCond:
		switch e.GetType(true) {
		case trace.CondWait:
			this.condVal[objId]--
			// release without signal/broadcast
			if this.condVal[objId] < 0 {
				this.IllFormedImpossible = true
			}
		case trace.CondSignal:
			this.condVal[objId]++
		case trace.CondBroadcast:
			this.condVal[objId] = math.MaxInt / 2
		}
	case *trace.ElementAtomic:
		// do nothing
	}

	this.trace = append(this.trace, elem)
}

// IsEqual returns if two traceMin are equal
//
// Parameter:
//   - other *TraceMin: the trace to compare against
//
// Returns:
//   - bool: true if the traces are equal, false otherwise
func (this *TraceEq) IsEqual(other *TraceEq) bool {
	if this.Len() != other.Len() {
		return false
	}

	for i := 0; i < this.Len(); i++ {
		if !this.trace[i].IsEqual(other.trace[i]) {
			return false
		}
	}

	return true
}

func (this *TraceEq) Get(index int) trace.Element {
	return this.trace[index]
}

// func (this *TraceMin) Key() string {
// 	res := ""
// 	for _, elem := range this.trace {
// 		res += (elem.Key() + ";")
// 	}
// 	return res
// }

// traceEqFromChain creates a trace min from a chain
//
// Parameter:
//   - chain Chain: the chain
func TraceEqFromConstraint(chain baseF.Constraint) TraceEq {
	res := NewTraceEq()

	mapping := make(map[string]trace.Element)

	minTPost := chain.ElemWithSmallestTPost().GetTSort()

	// add elements before chain
	traceIter := baseA.MainTrace.AsIterator()
	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if elem.GetTSort() >= minTPost {
			break
		}

		if trace.IsOp(elem) {
			res.AddElem(elem.Copy(mapping, false))
		}
	}

	// add chain
	for _, e := range chain.Elems {
		res.AddElem(e.Copy(mapping, false))
	}

	return res
}

// TraceEqFromTrace creates a traceEq from a trace
//
// Parameter:
//   - t *Trace: the full trace
//
// Returns:
//   - TraceMin: the minimal trace
func TraceEqFromTrace(t *trace.Trace) TraceEq {
	res := NewTraceEq()

	traceIter := t.AsIterator()

	for elem := traceIter.Next(); elem != nil; elem = traceIter.Next() {
		if trace.IsOp(elem) {
			res.trace = append(res.trace, elem)
		}
	}

	return res
}
