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
	ic_unknown             instClass = "unknown"
	ic_alloc               instClass = "alloc"
	ic_binOp               instClass = "binOp"
	ic_builtin             instClass = "builtin"
	ic_call                instClass = "call"
	ic_changeInterface     instClass = "changeInterface"
	ic_changeType          instClass = "changeType"
	ic_const               instClass = "const"
	ic_convert             instClass = "convert"
	ic_debugRef            instClass = "debugRef"
	ic_defer               instClass = "defer"
	ic_extract             instClass = "extract"
	ic_field               instClass = "field"
	ic_fieldAddr           instClass = "fieldAddr"
	ic_freeVar             instClass = "freeVar"
	ic_function            instClass = "function"
	ic_global              instClass = "global"
	ic_go                  instClass = "go"
	ic_if                  instClass = "if"
	ic_index               instClass = "index"
	ic_indexAddr           instClass = "indexAddr"
	ic_jump                instClass = "jump"
	ic_lookup              instClass = "lookup"
	ic_makeChan            instClass = "makeChan"
	ic_makeClosure         instClass = "makeClosure"
	ic_makeInterface       instClass = "makeInterface"
	ic_makeMap             instClass = "makeMap"
	ic_makeSlice           instClass = "makeSlice"
	ic_mapUpdate           instClass = "mapUpdate"
	ic_multiConvert        instClass = "multiConvert"
	ic_namedConst          instClass = "namedConst"
	ic_next                instClass = "next"
	ic_panic               instClass = "panic"
	ic_parameter           instClass = "parameter"
	ic_phi                 instClass = "phi"
	ic_range               instClass = "range"
	ic_return              instClass = "return"
	ic_runDefers           instClass = "runDefers"
	ic_select              instClass = "select"
	ic_send                instClass = "send"
	ic_slice               instClass = "slice"
	ic_sliceToArrayPointer instClass = "sliceToArrayPointer"
	ic_store               instClass = "store"
	ic_type                instClass = "type"
	ic_typeAssert          instClass = "typeAssert"
	ic_unOp                instClass = "unOp"
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

	if self.class != ic_unknown {
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
	case *ssa.Alloc:
		return ic_alloc
	case *ssa.BinOp:
		return ic_binOp
	case *ssa.Call:
		return ic_call
	case *ssa.ChangeInterface:
		return ic_changeInterface
	case *ssa.ChangeType:
		return ic_changeType
	case *ssa.Convert:
		return ic_convert
	case *ssa.DebugRef:
		return ic_debugRef
	case *ssa.Defer:
		return ic_defer
	case *ssa.Extract:
		return ic_extract
	case *ssa.Field:
		return ic_field
	case *ssa.FieldAddr:
		return ic_fieldAddr
	case *ssa.Go:
		return ic_go
	case *ssa.If:
		return ic_if
	case *ssa.Index:
		return ic_index
	case *ssa.IndexAddr:
		return ic_indexAddr
	case *ssa.Jump:
		return ic_jump
	case *ssa.Lookup:
		return ic_lookup
	case *ssa.MakeChan:
		return ic_makeChan
	case *ssa.MakeClosure:
		return ic_makeClosure
	case *ssa.MakeInterface:
		return ic_makeInterface
	case *ssa.MakeMap:
		return ic_makeMap
	case *ssa.MakeSlice:
		return ic_makeSlice
	case *ssa.MapUpdate:
		return ic_mapUpdate
	case *ssa.MultiConvert:
		return ic_multiConvert
	case *ssa.Next:
		return ic_next
	case *ssa.Panic:
		return ic_panic
	case *ssa.Phi:
		return ic_phi
	case *ssa.Range:
		return ic_range
	case *ssa.Return:
		return ic_return
	case *ssa.RunDefers:
		return ic_runDefers
	case *ssa.Select:
		return ic_select
	case *ssa.Send:
		return ic_send
	case *ssa.Slice:
		return ic_slice
	case *ssa.SliceToArrayPointer:
		return ic_sliceToArrayPointer
	case *ssa.Store:
		return ic_store
	case *ssa.TypeAssert:
		return ic_typeAssert
	case *ssa.UnOp:
		return ic_unOp
	}

	return ic_unknown
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

	// for _, fn := range self.funcs {
	// 	fmt.Println(fn.string())
	// 	fmt.Print("\n\n\n==================================================\n\n\n")
	// }
}
