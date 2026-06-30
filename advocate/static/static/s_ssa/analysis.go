// Copyright (c) 2026 Erik Kassubek
//
// File: analysis.go
// Brief: Read the ssa and pass into a format that is simpler to work with
//
// Author: Erik Kassubek
// Created: 2026-05-07
//
// License: BSD-3-Clause

package s_ssa

import (
	"advocate/static/static/s_base"
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// ================================================================
// Function
// ================================================================

type function struct {
	name   string
	pkg    string
	loc    string
	blocks []block
	sv     []staticVar
}

func (self *function) string() string {
	res := fmt.Sprintf("Name: %s\nPackage: %s\nLocation: %s\n", self.name, self.pkg, self.loc)
	for _, block := range self.blocks {
		res += "\t" + block.string()
	}

	res += "\n --------------------------------------------------\n"

	for _, s := range self.sv {
		res += s.string() + "\n"
	}

	return res
}

func (self *Data) analysisFunction(fn *ssa.Function) function {
	f := function{
		name:   fn.Name(),
		pkg:    fn.Pkg.Pkg.Path(),
		loc:    self.ast.GetPosFromPos(fn.Pos()),
		blocks: make([]block, len(fn.Blocks)),
		sv:     make([]staticVar, 0),
	}

	for i, block := range fn.Blocks {
		sv := []staticVar{}
		f.blocks[i], sv = self.analysisBlock(block)
		f.sv = append(f.sv, sv...)
	}

	return f
}

// ================================================================
// Block
// ================================================================

type block struct {
	id    int
	insts []Instruction
}

func (self *block) string() string {
	res := fmt.Sprintf("%d:\n", self.id)
	for _, inst := range self.insts {
		res += "\t\t" + inst.string() + "\n"
	}
	return res
}

func (self *Data) analysisBlock(bl *ssa.BasicBlock) (block, []staticVar) {
	b := block{
		id:    bl.Index,
		insts: make([]Instruction, len(bl.Instrs)),
	}

	v := make([]staticVar, 0)

	for i, instr := range bl.Instrs {
		sv := staticVar{}
		b.insts[i], sv = self.analysisInstruction(instr)
		if sv.name != "" && len(sv.objects) != 0 {
			v = append(v, sv)
		}
	}

	return b, v
}

// ================================================================
// Instruction
// ================================================================

type instClass string

const (
	unknown instClass = "unknown"
	alloc   instClass = "alloc"
	// TODO: list all relevant
)

type staticVar struct {
	name    string
	objects []s_base.ObjName
	equal   []staticVar
}

func (self *staticVar) string() string {
	res := self.name + " -> "
	for i, obj := range self.objects {
		if i != 0 {
			res += ","
		}
		res += string(obj)
	}

	return res
}

type Instruction struct {
	name string
	inst ssa.Instruction

	class instClass

	ptr bool

	hasConc [4]bool
}

func (self *Instruction) string() (res string) {
	if self.name == "" {
		res = self.inst.String()
	} else {
		res = fmt.Sprintf("%s = %s", self.name, self.inst)
	}

	// name
	name := func(i int) s_base.ObjName {
		switch i {
		case 0:
			return s_base.Channel
		case 1:
			return s_base.Mutex
		case 2:
			return s_base.CondVar
		case 3:
			return s_base.Wg
		default:
			return s_base.UnknownObj
		}
	}

	found := false
	for i := 0; i < 4; i++ {
		if self.hasConc[i] {
			if !found {
				res += "\t\t-> "
			} else {
				res += ", "
			}
			res += string(name(i))
			found = true
		}
	}

	if !found {
		return
	}

	if self.class != unknown {
		res += "\t\t-> " + string(self.class)
	}

	return
}

func (self *Data) analysisInstruction(instr ssa.Instruction) (Instruction, staticVar) {
	inst := Instruction{
		inst: instr,
	}

	if v, ok := instr.(ssa.Value); ok {
		inst.name = v.Name()
	}

	inst.hasConc[0], inst.hasConc[1], inst.hasConc[2], inst.hasConc[3] = containsSyncPrimitive(instr)

	sv := staticVar{
		name: inst.name,
	}

	inst.class = self.analysisClass(instr)

	return inst, sv
}

func containsSyncPrimitive(instr ssa.Instruction) (hasChan, hasMutex, hasCond, hasWaitGroup bool) {
	checkType := func(t types.Type) {
		if t == nil {
			return
		}

		if _, ok := t.Underlying().(*types.Chan); ok {
			hasChan = true
		}

		for {
			p, ok := t.(*types.Pointer)
			if !ok {
				break
			}
			t = p.Elem()
		}

		named, ok := t.(*types.Named)
		if !ok {
			return
		}

		obj := named.Obj()
		if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "sync" {
			return
		}

		switch obj.Name() {
		case "Mutex", "RWMutex":
			hasMutex = true
		case "Cond":
			hasCond = true
		case "WaitGroup":
			hasWaitGroup = true
		}
	}

	// Explicit channel instructions.
	switch i := instr.(type) {
	case *ssa.MakeChan, *ssa.Send, *ssa.Select:
		hasChan = true
	case *ssa.UnOp:
		if i.Op == token.ARROW {
			hasChan = true
		}
	}

	// Result value.
	if v, ok := instr.(ssa.Value); ok {
		checkType(v.Type())
	}

	// Operand values.
	for _, op := range instr.Operands(nil) {
		if op == nil || *op == nil {
			continue
		}
		checkType((*op).Type())
	}

	return
}

func (self *Data) analysisClass(instr ssa.Instruction) instClass {
	switch instr.(type) {
	case *ssa.Alloc, *ssa.MakeChan:
		return alloc
		// TODO: implement all relevant
	}

	return unknown
}

// ================================================================
// Analysis
// ================================================================

func (self *Data) runSSAAnalysis() {
	seen := make(map[string]bool)

	self.funcs = make([]function, 0)

	for _, pkg := range self.ssaPkgs {
		if pkg == nil {
			continue
		}

		path := pkg.Pkg.Path()
		if seen[path] {
			continue
		}
		seen[path] = true

		for _, mem := range pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				fn := self.analysisFunction(fn)
				self.funcs = append(self.funcs, fn)
			}
		}
	}

	for _, fn := range self.funcs {
		fmt.Println(fn.string())
		fmt.Print("\n\n\n==================================================\n\n\n")
	}
}
