// Copyright (c) 2025 Erik Kassubek
//
// File: vcOther.go
// Brief: Function for happens before info for
//   elements that do not change, but only store the vc
//
// Author: Erik Kassubek
// Created: 2025-04-26
//
// License: BSD-3-Clause

package analysis

import (
	"advocate/analysis/data"
	"advocate/trace"
)

// UpdateHBNew store the vector clock of the element
// Parameter:
//   - n *trace.TraceElementNew: the new trace element
func UpdateHBNew(n *trace.ElementNew) {
	routine := n.GetRoutine()
	n.SetVc(data.CurrentVC[routine])
	n.SetWVc(data.CurrentWVC[routine])
}

// UpdateHBRoutineEnd store the vector clock of the element
// Parameter:
//   - re *trace.TraceElementRoutineEnd: the new trace element
func UpdateHBRoutineEnd(re *trace.ElementRoutineEnd) {
	routine := re.GetRoutine()
	re.SetVc(data.CurrentVC[routine])
	re.SetWVc(data.CurrentWVC[routine])
}
