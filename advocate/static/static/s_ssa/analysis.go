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
	"advocate/static/static/s_base"
	"advocate/utils/log"
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// ================================================================
// MARK: Function
// ================================================================

type Function struct {
	name   string
	pkg    string
	loc    string
	blocks []block
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
		blocks: make([]block, len(fn.Blocks)),
	}

	for i, block := range fn.Blocks {
		f.blocks[i] = this.analysisBlock(block)
	}

	f.fv = fn.FreeVars

	return f
}

func isMain(fn *ssa.Function) bool {
	return fn.Name() == "main" &&
		fn.Signature.Recv() == nil &&
		fn.Pkg != nil &&
		fn.Pkg.Pkg.Name() == "main"
}

// ================================================================
// MARK: Block
// ================================================================

type block struct {
	id    int
	insts []Instruction
}

func (self *block) string() string {
	res := fmt.Sprintf("%d:\n", self.id)
	for _, inst := range self.insts {
		res += "\t\t" + inst.StringInfo() + "\n"
	}
	return res
}

func (this *Data) analysisBlock(bl *ssa.BasicBlock) block {
	b := block{
		id:    bl.Index,
		insts: make([]Instruction, len(bl.Instrs)),
	}

	for i, instr := range bl.Instrs {
		b.insts[i] = this.analysisInstruction(instr)
	}

	return b
}

// ================================================================
// MARK: Instruction
// ================================================================

type InstClass string

const (
	Ic_unknown             InstClass = "unknown"
	Ic_alloc               InstClass = "alloc"
	Ic_binOp               InstClass = "binOp"
	Ic_call                InstClass = "call"
	Ic_changeInterface     InstClass = "changeInterface"
	Ic_changeType          InstClass = "changeType"
	Ic_const               InstClass = "const"
	Ic_convert             InstClass = "convert"
	Ic_debugRef            InstClass = "debugRef"
	Ic_defer               InstClass = "defer"
	Ic_extract             InstClass = "extract"
	Ic_field               InstClass = "field"
	Ic_fieldAddr           InstClass = "fieldAddr"
	Ic_freeVar             InstClass = "freeVar"
	Ic_function            InstClass = "function"
	Ic_go                  InstClass = "go"
	Ic_if                  InstClass = "if"
	Ic_index               InstClass = "index"
	Ic_indexAddr           InstClass = "indexAddr"
	Ic_jump                InstClass = "jump"
	Ic_lookup              InstClass = "lookup"
	Ic_makeChan            InstClass = "makeChan"
	Ic_makeClosure         InstClass = "makeClosure"
	Ic_makeInterface       InstClass = "makeInterface"
	Ic_makeMap             InstClass = "makeMap"
	Ic_makeSlice           InstClass = "makeSlice"
	Ic_mapUpdate           InstClass = "mapUpdate"
	Ic_multiConvert        InstClass = "multiConvert"
	Ic_next                InstClass = "next"
	Ic_panic               InstClass = "panic"
	Ic_parameter           InstClass = "parameter"
	Ic_phi                 InstClass = "phi"
	Ic_range               InstClass = "range"
	Ic_return              InstClass = "return"
	Ic_runDefers           InstClass = "runDefers"
	Ic_select              InstClass = "select"
	Ic_send                InstClass = "send"
	Ic_slice               InstClass = "slice"
	Ic_sliceToArrayPointer InstClass = "sliceToArrayPointer"
	Ic_store               InstClass = "store"
	Ic_typeAssert          InstClass = "typeAssert"
	Ic_unOp                InstClass = "unOp"
)

type Instruction struct {
	name string
	inst ssa.Instruction

	class   InstClass
	inTrace bool
	relvant bool

	ptr bool

	conc hasConcInfo
}

func (this *Instruction) String() (res string) {
	if this.name == "" {
		res = this.inst.String()
	} else {
		res = fmt.Sprintf("%s = %s", this.name, this.inst.String())
	}

	return
}

func (this *Instruction) StringInfo() (res string) {
	if this.relvant {
		res += "+"
	} else {
		res += "-"
	}

	if this.inTrace {
		res += "+ "
	} else {
		res += "- "
	}

	res += fmt.Sprintf("%-40s", this.String())

	// name
	obj := func(i int) s_base.ObjName {
		switch concRes(i) {
		case chanInd:
			return s_base.Channel
		case mutexInd:
			return s_base.Mutex
		case condVarInd:
			return s_base.CondVar
		case wgInd:
			return s_base.Wg
		default:
			return s_base.UnknownObj
		}
	}

	if this.class != Ic_unknown {
		res += "\t-> " + fmt.Sprintf("%-20s", string(this.class))
	}

	found := false
	for i := 0; i < 4; i++ {
		if this.conc[i] {
			if !found {
				res += "  -> "
			} else {
				res += ", "
			}
			res += string(obj(i))
			found = true
		}
	}

	return
}

func (this *Instruction) HasChannel() bool {
	return this.conc[chanInd]
}

func (this *Instruction) HasMutex() bool {
	return this.conc[mutexInd]
}

func (this *Instruction) HasCond() bool {
	return this.conc[condVarInd]
}

func (this *Instruction) HasWg() bool {
	return this.conc[wgInd]
}

func (this *Instruction) Class() InstClass {
	return this.class
}

func (this *Data) analysisInstruction(instr ssa.Instruction) Instruction {
	name := ""
	switch v := instr.(type) {
	case ssa.Value:
		name = v.Name()
	case nil:
		// Be robust against bad transforms.
		name = "<deleted>"
	}

	inst := Instruction{
		name: name,
		inst: instr,
	}

	inst.conc[0], inst.conc[1], inst.conc[2], inst.conc[3] = containsSyncPrimitive(instr)

	inst.class = this.analysisClass(instr)

	inst.relvant, inst.inTrace = this.isRelvant(inst)

	return inst
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

func (this *Data) analysisClass(instr ssa.Instruction) InstClass {
	switch instr.(type) {
	case *ssa.Alloc:
		return Ic_alloc
	case *ssa.BinOp:
		return Ic_binOp
	case *ssa.Call:
		return Ic_call
	case *ssa.ChangeInterface:
		return Ic_changeInterface
	case *ssa.ChangeType:
		return Ic_changeType
	case *ssa.Convert:
		return Ic_convert
	case *ssa.DebugRef:
		return Ic_debugRef
	case *ssa.Defer:
		return Ic_defer
	case *ssa.Extract:
		return Ic_extract
	case *ssa.Field:
		return Ic_field
	case *ssa.FieldAddr:
		return Ic_fieldAddr
	case *ssa.Go:
		return Ic_go
	case *ssa.If:
		return Ic_if
	case *ssa.Index:
		return Ic_index
	case *ssa.IndexAddr:
		return Ic_indexAddr
	case *ssa.Jump:
		return Ic_jump
	case *ssa.Lookup:
		return Ic_lookup
	case *ssa.MakeChan:
		return Ic_makeChan
	case *ssa.MakeClosure:
		return Ic_makeClosure
	case *ssa.MakeInterface:
		return Ic_makeInterface
	case *ssa.MakeMap:
		return Ic_makeMap
	case *ssa.MakeSlice:
		return Ic_makeSlice
	case *ssa.MapUpdate:
		return Ic_mapUpdate
	case *ssa.MultiConvert:
		return Ic_multiConvert
	case *ssa.Next:
		return Ic_next
	case *ssa.Panic:
		return Ic_panic
	case *ssa.Phi:
		return Ic_phi
	case *ssa.Range:
		return Ic_range
	case *ssa.Return:
		return Ic_return
	case *ssa.RunDefers:
		return Ic_runDefers
	case *ssa.Select:
		return Ic_select
	case *ssa.Send:
		return Ic_send
	case *ssa.Slice:
		return Ic_slice
	case *ssa.SliceToArrayPointer:
		return Ic_sliceToArrayPointer
	case *ssa.Store:
		return Ic_store
	case *ssa.TypeAssert:
		return Ic_typeAssert
	case *ssa.UnOp:
		return Ic_unOp
	}

	log.Error("Unknown SSA class: ", instr.String())
	return Ic_unknown
}

// Determine if instruction is relevant
//
// Parameter:
//   - instr ssa.Instruction: the instruciton
//
// Returns:
//   - bool: is relevant
//   - bool: is in trace
func (this *Data) isRelvant(instr Instruction) (bool, bool) {
	resource := instr.conc.Resource()

	switch instr.class {
	case Ic_go, Ic_makeChan, Ic_return, Ic_select, Ic_send:
		return true, true
	case Ic_alloc, Ic_unOp:
		return resource, resource
	case Ic_mapUpdate, Ic_lookup, Ic_extract, Ic_store:
		return resource, false
	case Ic_field, Ic_fieldAddr, Ic_freeVar, Ic_index, Ic_indexAddr, Ic_jump, Ic_makeClosure, Ic_makeInterface, Ic_makeMap, Ic_makeSlice, Ic_range, Ic_next, Ic_runDefers, Ic_slice, Ic_sliceToArrayPointer:
		return true, false
	case Ic_call:
		i := instr.inst.(*ssa.Call)
		fn := i.Common().StaticCallee()

		if fn == nil {
			return resource, resource
		}

		if _, ok := this.funcs[fn.String()]; !ok {
			return resource, resource
		}

		return true, true
	case Ic_if: // TODO: should be recorded, change when recorded
		return true, false

	}

	return false, false
}

func (this *Data) getInstructionPos(instr ssa.Instruction) (string, int) {
	pos := instr.Pos()
	if pos == token.NoPos {
		return "", 0
	}

	position := this.ssa.Fset.Position(pos)
	return position.Filename, position.Line
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

func (this *hasConcInfo) Resource() bool {
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
