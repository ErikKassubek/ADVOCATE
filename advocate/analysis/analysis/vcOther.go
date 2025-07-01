// Copyright (c) 2025 Erik Kassubek
//
// File: vcOther.go
// Brief: Function involving the vector clocks for
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

// UpdateVCNew store the vector clock of the element
// Parameter:
//   - n *trace.TraceElementNew: the new trace element
func UpdateVCNew(n *trace.TraceElementNew) {
	routine := n.GetRoutine()
	n.SetVc(data.CurrentVC[routine])
	n.SetWVc(data.CurrentWVC[routine])
}

// UpdateVCRoutineEnd store the vector clock of the element
// Parameter:
//   - re *trace.TraceElementRoutineEnd: the new trace element
func UpdateVCRoutineEnd(re *trace.TraceElementRoutineEnd) {
	routine := re.GetRoutine()
	re.SetVc(data.CurrentVC[routine])
	re.SetWVc(data.CurrentWVC[routine])
}
