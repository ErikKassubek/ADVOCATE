// Copyright (c) 2025 Erik Kassubek
//
// File: hbAtomic.go
// Brief: Store the vc for new and routine element
//
// Author: Erik Kassubek
// Created: 2025-07-20
//
// License: BSD-3-Clause

package vc

import "advocate/trace"

// UpdateHBNew store the vector clock of the element
// Parameter:
//   - n *trace.TraceElementNew: the new trace element
func UpdateHBNew(n *trace.ElementNew) {
	routine := n.GetRoutine()
	n.SetVc(CurrentVC[routine])
	n.SetWVc(CurrentWVC[routine])
}

// UpdateHBRoutineEnd store the vector clock of the element
// Parameter:
//   - re *trace.TraceElementRoutineEnd: the new trace element
func UpdateHBRoutineEnd(re *trace.ElementRoutineEnd) {
	routine := re.GetRoutine()
	re.SetVc(CurrentVC[routine])
	re.SetWVc(CurrentWVC[routine])
}
