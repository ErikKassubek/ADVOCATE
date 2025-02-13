// Copyrigth (c) 2024 Erik Kassubek
//
// File: rewriter.go
// Brief: Main functions to rewrite the trace
//
// Author: Erik Kassubek
// Created: 2024-10-25
//
// License: BSD-3-Clause

// Package rewriter provides functions for rewriting traces.

package rewriter

// ========= Not executed select with partner =========================
// func rewriteNotExecutedSelect(bug bugs.Bug, index int) error {
// 	sel := bug.TraceElement1[0]
// 	if sel.GetTSort() == 0 {
// 		return fmt.Errorf("Cannot rewrite not executed case, select was not executed")
// 	}

// 	ca := bug.TraceElement1Sel[0]
// 	if index >= len(bug.TraceElement2) {
// 		return fmt.Errorf("Index %d not in possible partner (len %d)", index, len(bug.TraceElement2))
// 	}

// 	selPartner := sel.(*analysis.TraceElementSelect).GetPartner()
// 	if selPartner != nil {
// 		selPartner.SetTPost(0)
// 	}

// 	if ca.ObjType == "CS" {
// 		bug.TraceElement2[index].SetTSort(sel.GetTSort() + 1)
// 		bug.TraceElement1[0].(*analysis.TraceElementSelect).SetCase(ca.ID, analysis.SendOp)
// 	} else {
// 		bug.TraceElement2[index].SetTSort(sel.GetTSort() - 1)
// 		bug.TraceElement1[0].(*analysis.TraceElementSelect).SetCase(ca.ID, analysis.RecvOp)
// 	}

// 	partner := bug.TraceElement1[0].(*analysis.TraceElementSelect).GetPartner()
// 	if partner != nil {
// 		analysis.RemoveElementFromTrace(bug.TraceElement1[0].(*analysis.TraceElementSelect).GetPartner().GetTID())
// 	}

// 	return nil
// }
