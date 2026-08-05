// Copyright (c) 2026 Erik Kassubek
//
// File: InstructionIf.go
// Brief: If Instruction
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/ssa"
)

type InstructionIf struct {
	InstructionBase

	case_if   int
	case_else int
}

func newIf(f *Function, inst *ssa.If, i int, data *Data) *InstructionIf {
	succ0 := inst.Block().Succs[0]
	succ1 := inst.Block().Succs[1]

	caseIf := succ0.Index
	caseElse := succ1.Index

	ifStmt := findASTIf(data, inst.Pos())

	if ifStmt != nil {
		ifLine := firstStmtLine(data, ifStmt.Body)

		elseLine := -1
		if elseBlock, ok := ifStmt.Else.(*ast.BlockStmt); ok {
			elseLine = firstStmtLine(data, elseBlock)
		}

		line0 := firstInstrLine(data, succ0)
		line1 := firstInstrLine(data, succ1)

		// Pick the successor closest to the source if-body.
		if ifLine != -1 {
			dist0 := abs(line0 - ifLine)
			dist1 := abs(line1 - ifLine)

			if dist1 < dist0 {
				caseIf = succ1.Index
				caseElse = succ0.Index
			}
		}

		// Optional sanity check: if the else branch is closer to the
		// else body and the if branch was not identified correctly.
		if elseLine != -1 {
			ifLineDist0 := abs(line0 - ifLine)
			ifLineDist1 := abs(line1 - ifLine)

			elseDist0 := abs(line0 - elseLine)
			elseDist1 := abs(line1 - elseLine)

			if elseDist0 < elseDist1 && ifLineDist1 < ifLineDist0 {
				caseIf = succ1.Index
				caseElse = succ0.Index
			}
		}
	}

	return &InstructionIf{
		InstructionBase: newInstructionBase(f, Ic_if, inst, i),
		case_if:         caseIf,
		case_else:       caseElse,
	}
}

func (this *InstructionIf) Instruction() *ssa.If {
	return this.inst.(*ssa.If)
}

func (this *InstructionIf) setRelevant(_ *Data) {
	this.relevant = true
	this.inTrace = true
}

func (this *InstructionIf) If() int {
	return this.case_if
}

func (this *InstructionIf) Else() int {
	return this.case_else
}

func firstInstrLine(data *Data, b *ssa.BasicBlock) int {
	for _, instr := range b.Instrs {
		if instr.Pos().IsValid() {
			return data.ast.Fset.Position(instr.Pos()).Line
		}
	}

	return -1
}

func firstStmtLine(data *Data, block *ast.BlockStmt) int {
	if block == nil || len(block.List) == 0 {
		return -1
	}

	return data.ast.Fset.Position(block.List[0].Pos()).Line
}

func findASTIf(data *Data, pos token.Pos) *ast.IfStmt {
	for _, files := range data.ast.AstMap {
		for _, file := range files {
			var result *ast.IfStmt

			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil || result != nil {
					return false
				}

				if ifs, ok := n.(*ast.IfStmt); ok && ifs.If == pos {
					result = ifs
					return false
				}

				return true
			})

			if result != nil {
				return result
			}
		}
	}

	return nil
}

func abs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}
