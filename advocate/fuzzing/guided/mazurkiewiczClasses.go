// Copyright (c) 2025 Erik Kassubek
//
// File: mazurkiewiczClasses.go
// Brief: Some elements in traces can be swapped without changing a program
//    execution regarding concurrency bugs. Functions in this file check
//    if two traces consist of only such switches, meaning there is no need
//    to run both execution during fuzzing
//
// Author: Erik Kassubek
// Created: 2025-10-10
//
// License: BSD-3-Clause

package guided

// areConflicting checks if two operations are conflicting. To operations
// are conflicting if changing the order of those operations in the trace could
// transform the execution of the trace from one not containing any concurrency
// bugs to one, that contains them
//
// Parameter:
//   - trace *trace.Trace: the trace
//   - op1 trace.Element: the first element in the trace
//   - op2 trace.Element: the second element in the trace
// func areConflicting(trace *trace.Trace, op1, op2 trace.Element) bool {
// 	if op1.GetObjType(false) != op2.GetObjType(false) {

// 	}
// }
