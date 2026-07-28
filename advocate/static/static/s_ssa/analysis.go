// Copyright (c) 2026 Erik Kassubek
//
// File: analysis.go
// Brief: Read the ssa and pass into a format that is simpler to work with
//
// Author: Erik Kassubek
//
// License: BSD-3-Clause

package s_ssa

import (
	"fmt"

	"golang.org/x/tools/go/ssa"
)

// ================================================================
// MARK: Function
// ================================================================

type Function struct {
	name   string
	pkg    string
	loc    string
	blocks []*Block
	fv     []*ssa.FreeVar
}

func (this *Function) Name() string {
	return this.name
}

func (this *Function) string() string {
	res := fmt.Sprintf("Name: %s\n", this.name)

	if this.fv != nil {
		res += "Free variables:\n"
		for i, fv := range this.fv {
			res += fmt.Sprintf("\t%d:\t%s\n", i, fv.Name())
		}
	}

	res += "Func:\n"

	for _, block := range this.blocks {
		res += "\t" + block.string()
	}

	res += "\n --------------------------------------------------\n"

	return res
}

func (this *Data) analysisFunction(fn *ssa.Function) Function {
	f := Function{
		name:   fn.String(),
		pkg:    fn.Pkg.Pkg.Path(),
		loc:    this.ast.GetPosFromPos(fn.Pos()),
		blocks: make([]*Block, len(fn.Blocks)),
	}

	for i, block := range fn.Blocks {
		f.blocks[i] = this.analysisBlock(block)
	}

	f.fv = fn.FreeVars

	return f
}

func (this *Data) funcFromName(name string) *Function {
	return this.Funcs()[name]
}

func isMain(fn *ssa.Function) bool {
	return fn.Name() == "main" &&
		fn.Signature.Recv() == nil &&
		fn.Pkg != nil &&
		fn.Pkg.Pkg.Name() == "main"
}

func isInit(fn *ssa.Function) bool {
	return fn.Name() == "init" &&
		fn.Signature.Recv() == nil &&
		fn.Pkg != nil &&
		fn.Pkg.Pkg.Name() == "main"
}

func (this *Function) Blocks() []*Block {
	return this.blocks
}

func (this *Function) FreeVar() []*ssa.FreeVar {
	return this.fv
}

// ================================================================
// MARK: Block
// ================================================================

type Block struct {
	id    int
	insts []Instruction
}

func (self *Block) string() string {
	res := fmt.Sprintf("%d:\n", self.id)
	for _, inst := range self.insts {
		res += "\t\t" + inst.StringInfo() + "\n"
	}
	return res
}

func (this *Block) Instrs() []Instruction {
	if this == nil {
		return make([]Instruction, 0)
	}
	return this.insts
}

func (this *Data) analysisBlock(bl *ssa.BasicBlock) *Block {
	b := Block{
		id:    bl.Index,
		insts: make([]Instruction, len(bl.Instrs)),
	}

	for i, instr := range bl.Instrs {
		b.insts[i] = this.analysisInstruction(instr)
	}

	return &b
}

// ================================================================
// MARK: ConcInfo
// ================================================================

type hasConcInfo [4]bool

type concRes int

const (
	chanInd concRes = iota
	mutexInd
	condVarInd
	wgInd
)

func (this hasConcInfo) Resource() bool {
	for i := 0; i < 4; i++ {
		if this[i] {
			return true
		}
	}

	return false
}

func (this *hasConcInfo) Channel() bool {
	return this[chanInd]
}

func (this *hasConcInfo) Mutex() bool {
	return this[mutexInd]
}

func (this *hasConcInfo) CondVar() bool {
	return this[condVarInd]
}

func (this *hasConcInfo) WaitGroup() bool {
	return this[wgInd]
}

func (this *concRes) string() string {
	switch *this {
	case chanInd:
		return "chan"
	case mutexInd:
		return "mutex"
	case condVarInd:
		return "condVar"
	case wgInd:
		return "wg"
	}
	return ""
}

// ================================================================
// MARK: Free Var
// ================================================================

type freeVar struct {
	id   int
	name string
	t    concRes
}

func (this *freeVar) string() string {
	return fmt.Sprintf("%d:  %s %s", this.id, this.name, this.t.string())
}

// ================================================================
// MARK: Analysis
// ================================================================

func (this *Data) runSSAAnalysis() {
	seen := make(map[string]bool)

	this.funcs = make(map[string]*Function)

	for _, pkg := range this.ssaPkgs {
		if pkg == nil {
			continue
		}

		path := pkg.Pkg.Path()
		if seen[path] {
			continue
		}
		seen[path] = true

		var addFunc func(*ssa.Function)
		addFunc = func(fn *ssa.Function) {
			if fn == nil {
				return
			}
			f := this.analysisFunction(fn)
			this.funcs[f.Name()] = &f

			if isMain(fn) {
				this.mainFunc = &f
			} else if isInit(fn) {
				this.initFunc = &f
			}

			for _, anon := range fn.AnonFuncs {
				addFunc(anon)
			}
		}

		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				addFunc(fn)
			}
		}

	}

	// for _, fn := range self.funcs {
	// 	fmt.emln(fn.string())
	// 	fmt.Print("\n\n\n==================================================\n\n\n")
	// }
}
