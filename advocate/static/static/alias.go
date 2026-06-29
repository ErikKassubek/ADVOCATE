// Copyright (c) 2026 Erik Kassubek
//
// File: alias.go
// Brief: Determine if two variables in the ast correspond to the same object
//
// Author: Erik Kassubek
// Created: 2026-05-07
//
// License: BSD-3-Clause

package static

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// ================================================================
// Alias Result
// ================================================================

type aliasResult int

const (
	never aliasResult = iota
	sometimes
	always
)

func (self *aliasResult) string() string {
	switch *self {
	case never:
		return "never"
	case sometimes:
		return "sometimes"
	case always:
		return "always"
	default:
		return "unknown"
	}
}

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

func (self *staticData) aliasFunction(fn *ssa.Function) function {
	f := function{
		name:   fn.Name(),
		pkg:    fn.Pkg.Pkg.Path(),
		loc:    self.getPosFromPos(fn.Pos()),
		blocks: make([]block, len(fn.Blocks)),
		sv:     make([]staticVar, 0),
	}

	for i, block := range fn.Blocks {
		sv := []staticVar{}
		f.blocks[i], sv = self.aliasBlock(block)
		f.sv = append(f.sv, sv...)
	}

	return f
}

// ================================================================
// Block
// ================================================================

type block struct {
	id    int
	insts []instruction
}

func (self *block) string() string {
	res := fmt.Sprintf("%d:\n", self.id)
	for _, inst := range self.insts {
		res += "\t\t" + inst.string() + "\n"
	}
	return res
}

func (self *staticData) aliasBlock(bl *ssa.BasicBlock) (block, []staticVar) {
	b := block{
		id:    bl.Index,
		insts: make([]instruction, len(bl.Instrs)),
	}

	v := make([]staticVar, 0)

	for i, instr := range bl.Instrs {
		sv := staticVar{}
		b.insts[i], sv = self.aliasInstruction(instr)
		if sv.name != "" && len(sv.objects) != 0 {
			v = append(v, sv)
		}
	}

	return b, v
}

// ================================================================
// Instruction
// ================================================================

type staticVar struct {
	name    string
	objects []objName
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

type instruction struct {
	name string
	inst string

	// class instructionClass

	ptr     bool
	objType []objName
}

func (self *instruction) string() (res string) {
	if self.name == "" {
		res = self.inst
	} else {
		res = fmt.Sprintf("%s = %s", self.name, self.inst)
	}

	if len(self.objType) != 0 {
		res += "\t\t-> "
		for i, obj := range self.objType {
			if i != 0 {
				res += ","
			}
			res += string(obj)
		}
	}
	return
}

func (self *staticData) aliasInstruction(instr ssa.Instruction) (instruction, staticVar) {
	inst := instruction{
		inst:    instr.String(),
		objType: make([]objName, 0),
	}

	if v, ok := instr.(ssa.Value); ok {
		inst.name = v.Name()
	}

	hasChan, hasMutex, hasCond, hasWaitGroup := ContainsSyncPrimitive(instr)

	if hasChan {
		inst.objType = append(inst.objType, channel)
	}
	if hasMutex {
		inst.objType = append(inst.objType, mutex)
	}
	if hasCond {
		inst.objType = append(inst.objType, condVar)
	}
	if hasWaitGroup {
		inst.objType = append(inst.objType, wg)
	}

	sv := staticVar{
		name:    inst.name,
		objects: inst.objType,
	}

	return inst, sv
}

func ContainsSyncPrimitive(instr ssa.Instruction) (hasChan, hasMutex, hasCond, hasWaitGroup bool) {
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

// ================================================================
// Main
// ================================================================

func (self *staticData) runAliasAnalysis() {
	seen := make(map[string]bool)

	funcs := make([]function, 0)

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
				fn := self.aliasFunction(fn)
				funcs = append(funcs, fn)
			}
		}
	}

	for _, fn := range funcs {
		fmt.Println(fn.string())
		fmt.Print("\n\n\n==================================================\n\n\n")
	}
}
