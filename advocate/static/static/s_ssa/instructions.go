// Copyright (c) 2026 Erik Kassubek
//
// File: instructions.go
// Brief: Instructions
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
	"strings"

	"golang.org/x/tools/go/ssa"
)

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

type Instruction interface {
	Variable() string
	Term() string
	VariableGlobal() bool
	TermGlobal() bool
	String() string
	StringInfo() string
	Inst() ssa.Instruction
	InTrace() bool
	Relevant() bool

	Conc() hasConcInfo
	HasChannel() bool
	HasMutex() bool
	HasCond() bool
	HasWG() bool

	Class() InstClass

	setVariable(name string, global bool)
	setTerm(term string, global bool)
	setRelevant(rel, trace bool)
	setInst(inst ssa.Instruction)
	setConc(conc hasConcInfo)

	Function() *Function
	Block() *Block
	Next() Instruction
	FirstInBlock(b_id int) Instruction
}

// ================================================================
// MARK: Base
// ================================================================

type InstructionBase struct {
	variable string
	term     string
	inst     ssa.Instruction

	variableGlobal bool
	termGlobal     bool

	varPtr bool

	inTrace bool
	relvant bool

	class InstClass

	conc hasConcInfo

	f    *Function
	b_id int
	i_id int
}

func newInstructionBase(f *Function, c InstClass, inst ssa.Instruction, index int) InstructionBase {
	name := ""
	switch v := inst.(type) {
	case ssa.Value:
		name = v.Name()
	case nil:
		// Be robust against bad transforms.
		name = "<deleted>"
	}

	globalName := false
	globalTerm := false

	if i, ok := inst.(*ssa.Store); ok {
		switch i.Addr.(type) {
		case *ssa.Global:
			globalName = true
		}

		switch i.Val.(type) {
		case *ssa.Global:
			globalTerm = true
		}
	}

	if i, ok := inst.(*ssa.UnOp); ok {
		switch i.X.(type) {
		case *ssa.Global:
			globalTerm = true
		}
	}

	// global var assign
	term := inst.String()
	if name == "" && strings.Contains(term, " = ") {
		fields := strings.Split(term, " = ")
		name = fields[0]
		term = fields[1]
	}

	b := InstructionBase{class: c, f: f, b_id: inst.Block().Index, i_id: index}

	b.setVariable(name, globalName)
	b.setTerm(term, globalTerm)
	b.setInst(inst)

	return b
}

func (this *InstructionBase) String() (res string) {
	if this.variable == "" {
		res = this.term
	} else {
		if this.varPtr {
			res = fmt.Sprintf("*%s = %s", this.variable, this.term)
		} else {
			res = fmt.Sprintf("%s = %s", this.variable, this.term)
		}
	}

	return
}

func (this *InstructionBase) StringInfo() (res string) {
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

func (this *InstructionBase) Variable() string {
	return this.variable
}

func (this *InstructionBase) Term() string {
	return this.term
}

func (this *InstructionBase) VariableGlobal() bool {
	return this.variableGlobal
}

func (this *InstructionBase) TermGlobal() bool {
	return this.termGlobal
}

func (this *InstructionBase) Class() InstClass {
	return this.class
}

func (this *InstructionBase) Inst() ssa.Instruction {
	return this.inst
}

func (this *InstructionBase) Conc() hasConcInfo {
	return this.conc
}

func (this *InstructionBase) HasChannel() bool {
	return this.conc[chanInd]
}

func (this *InstructionBase) HasMutex() bool {
	return this.conc[mutexInd]
}

func (this *InstructionBase) HasCond() bool {
	return this.conc[condVarInd]
}

func (this *InstructionBase) HasWG() bool {
	return this.conc[wgInd]
}

func (this *InstructionBase) Relevant() bool {
	return this.relvant
}

func (this *InstructionBase) InTrace() bool {
	return this.inTrace
}

func (this *InstructionBase) setRelevant(rel, trace bool) {
	this.relvant = rel
	this.inTrace = trace
}

func (this *InstructionBase) setVariable(name string, global bool) {
	variable := strings.TrimPrefix(name, "*")
	this.variable = variable
	this.varPtr = (name != variable)
	this.variableGlobal = global
}

func (this *InstructionBase) setTerm(term string, global bool) {
	this.term = term
	this.termGlobal = global
}

func (this *InstructionBase) setInst(inst ssa.Instruction) {
	this.inst = inst
}

func (this *InstructionBase) setConc(conc hasConcInfo) {
	this.conc = conc
}

func (this *InstructionBase) Function() *Function {
	return this.f
}

func (this *InstructionBase) Block() *Block {
	return this.f.blocks[this.b_id]
}

func (this *Data) analysisInstruction(f *Function, inst ssa.Instruction, i int) Instruction {

	var instru Instruction

	switch inst := inst.(type) {
	case *ssa.Alloc:
		instru = &InstructionAlloc{InstructionBase: newInstructionBase(f, Ic_alloc, inst, i)}
	case *ssa.BinOp:
		instru = &InstructionBinOp{InstructionBase: newInstructionBase(f, Ic_binOp, inst, i)}
	case *ssa.Call:
		instru = NewInstructionCall(f, inst, i)
	case *ssa.ChangeInterface:
		instru = &InstructionChangeInterface{InstructionBase: newInstructionBase(f, Ic_changeInterface, inst, i)}
	case *ssa.ChangeType:
		instru = &InstructionChangeType{InstructionBase: newInstructionBase(f, Ic_changeType, inst, i)}
	case *ssa.Convert:
		instru = &InstructionConvert{InstructionBase: newInstructionBase(f, Ic_convert, inst, i)}
	case *ssa.DebugRef:
		instru = &InstructionDebugRef{InstructionBase: newInstructionBase(f, Ic_debugRef, inst, i)}
	case *ssa.Defer:
		instru = &InstructionDefer{InstructionBase: newInstructionBase(f, Ic_defer, inst, i)}
	case *ssa.Extract:
		instru = &InstructionExtract{InstructionBase: newInstructionBase(f, Ic_extract, inst, i)}
	case *ssa.Field:
		instru = &InstructionField{InstructionBase: newInstructionBase(f, Ic_field, inst, i)}
	case *ssa.FieldAddr:
		instru = &InstructionFieldAddr{InstructionBase: newInstructionBase(f, Ic_fieldAddr, inst, i)}
	case *ssa.Go:
		instru = &InstructionGo{InstructionBase: newInstructionBase(f, Ic_go, inst, i)}
	case *ssa.If:
		instru = newInstructionIf(f, inst, i)
	case *ssa.Index:
		instru = &InstructionIndex{InstructionBase: newInstructionBase(f, Ic_index, inst, i)}
	case *ssa.IndexAddr:
		instru = &InstructionIndexAddr{InstructionBase: newInstructionBase(f, Ic_indexAddr, inst, i)}
	case *ssa.Jump:
		instru = newInstructionJump(f, inst, i)
	case *ssa.Lookup:
		instru = &InstructionLookup{InstructionBase: newInstructionBase(f, Ic_lookup, inst, i)}
	case *ssa.MakeChan:
		instru = &InstructionMakeChan{InstructionBase: newInstructionBase(f, Ic_makeChan, inst, i)}
	case *ssa.MakeClosure:
		instru = &InstructionMakeClosure{InstructionBase: newInstructionBase(f, Ic_makeClosure, inst, i)}
	case *ssa.MakeInterface:
		instru = &InstructionMakeInterface{InstructionBase: newInstructionBase(f, Ic_makeInterface, inst, i)}
	case *ssa.MakeMap:
		instru = &InstructionMakeMap{InstructionBase: newInstructionBase(f, Ic_makeMap, inst, i)}
	case *ssa.MakeSlice:
		instru = &InstructionMakeSlice{InstructionBase: newInstructionBase(f, Ic_makeSlice, inst, i)}
	case *ssa.MapUpdate:
		instru = &InstructionMapUpdate{InstructionBase: newInstructionBase(f, Ic_mapUpdate, inst, i)}
	case *ssa.MultiConvert:
		instru = &InstructionMultiConvert{InstructionBase: newInstructionBase(f, Ic_multiConvert, inst, i)}
	case *ssa.Next:
		instru = &InstructionNext{InstructionBase: newInstructionBase(f, Ic_next, inst, i)}
	case *ssa.Panic:
		instru = &InstructionPanic{InstructionBase: newInstructionBase(f, Ic_panic, inst, i)}
	case *ssa.Phi:
		instru = &InstructionPhi{InstructionBase: newInstructionBase(f, Ic_phi, inst, i)}
	case *ssa.Range:
		instru = &InstructionRange{InstructionBase: newInstructionBase(f, Ic_range, inst, i)}
	case *ssa.Return:
		instru = &InstructionReturn{InstructionBase: newInstructionBase(f, Ic_return, inst, i)}
	case *ssa.RunDefers:
		instru = &InstructionRunDefers{InstructionBase: newInstructionBase(f, Ic_runDefers, inst, i)}
	case *ssa.Select:
		instru = &InstructionSelect{InstructionBase: newInstructionBase(f, Ic_select, inst, i)}
	case *ssa.Send:
		instru = &InstructionSend{InstructionBase: newInstructionBase(f, Ic_send, inst, i)}
	case *ssa.Slice:
		instru = &InstructionSlice{InstructionBase: newInstructionBase(f, Ic_slice, inst, i)}
	case *ssa.SliceToArrayPointer:
		instru = &InstructionSliceToArrayPointer{InstructionBase: newInstructionBase(f, Ic_sliceToArrayPointer, inst, i)}
	case *ssa.Store:
		instru = &InstructionStore{InstructionBase: newInstructionBase(f, Ic_store, inst, i)}
	case *ssa.TypeAssert:
		instru = &InstructionTypeAssert{InstructionBase: newInstructionBase(f, Ic_typeAssert, inst, i)}
	case *ssa.UnOp:
		instru = &InstructionUnOp{InstructionBase: newInstructionBase(f, Ic_unOp, inst, i)}
	default:
		log.Error("Found unknown instruction: ", inst.String())
		instru = &InstructionUnknown{InstructionBase: newInstructionBase(f, Ic_unknown, inst, i)}
	}

	setContainsPrimitive(instru)

	this.isRelvant(instru)

	return instru
}

func setContainsPrimitive(inst Instruction) {
	hasChan := false
	hasMutex := false
	hasCond := false
	hasWaitGroup := false

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

	instr := inst.Inst()

	// Explicit channel instructions.
	switch i := instr.(type) {
	case *ssa.Alloc:
		hasChan = strings.HasPrefix(inst.Term(), "new chan")
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

	inst.setConc(hasConcInfo{hasChan, hasMutex, hasCond, hasWaitGroup})
}

// Determine if instruction is relevant
//
// Parameter:
//   - instr ssa.Instruction: the instruciton
//
// Returns:rele
//   - bool: is relevant
//   - bool: is in trace
func (this *Data) isRelvant(instr Instruction) {
	resource := instr.Conc().Resource()

	switch instr := instr.(type) {
	case *InstructionGo, *InstructionMakeChan, *InstructionReturn, *InstructionSelect, *InstructionSend, *InstructionIf:
		instr.setRelevant(true, true)
	case *InstructionAlloc:
		instr.setRelevant(resource, !instr.Conc().Channel())
	case *InstructionUnOp:
		instr.setRelevant(resource, instr.Inst().(*ssa.UnOp).Op == token.ARROW) // receive
	case *InstructionMapUpdate, *InstructionLookup, *InstructionExtract, *InstructionStore:
		instr.setRelevant(resource, false)
	case *InstructionField,
		*InstructionFieldAddr,
		*InstructionFreeVar,
		*InstructionIndex,
		*InstructionIndexAddr,
		*InstructionJump,
		*InstructionMakeClosure,
		*InstructionMakeInterface,
		*InstructionMakeMap,
		*InstructionMakeSlice,
		*InstructionRange,
		*InstructionNext,
		*InstructionRunDefers,
		*InstructionSlice,
		*InstructionSliceToArrayPointer:
		instr.setRelevant(true, false)
	case *InstructionCall:
		i := instr.Inst().(*ssa.Call)
		fn := i.Common().StaticCallee()

		if fn == nil {
			instr.setRelevant(resource, resource)
			return
		}

		if _, ok := this.funcs[fn.String()]; !ok {
			instr.setRelevant(resource, resource)
			return
		}

		instr.setRelevant(true, true)
	default:
		instr.setRelevant(false, false)
	}
}

func (this *Data) getInstructionPos(instr ssa.Instruction) (string, int) {
	pos := instr.Pos()
	if pos == token.NoPos {
		return "", 0
	}

	position := this.ssa.Fset.Position(pos)
	return position.Filename, position.Line
}

func (this *InstructionBase) FirstInBlock(b_id int) Instruction {
	return this.Function().Blocks()[b_id].Instrs()[0]
}

func (this *InstructionBase) Next() Instruction {
	instID := this.i_id + 1

	res := this.Block().Instrs()[instID]

	log.Debug("NEXT: ", this.String(), " -> ", res.String())

	return res
}

func NewSsaPosFunc(f *Function) Instruction {
	b := f.Blocks()[0]
	return b.Instrs()[0]
}

func NewSsaPosFuncBlock(f *Function, blockID int) Instruction {
	b := f.Blocks()[blockID]
	return b.Instrs()[0]
}

// ================================================================
// MARK: InstructionUnknown
// ================================================================

type InstructionUnknown struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionAlloc
// ================================================================

type InstructionAlloc struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionBinOp
// ================================================================

type InstructionBinOp struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionCall
// ================================================================

type InstructionCall struct {
	InstructionBase

	funcName string
	f        *Function
}

func NewInstructionCall(f *Function, inst *ssa.Call, i int) *InstructionCall {
	name := ""
	if callee := inst.Common().StaticCallee(); callee != nil {
		name = callee.String()
	}

	return &InstructionCall{InstructionBase: newInstructionBase(f, Ic_call, inst, i),
		funcName: name}
}

func (this *InstructionCall) GetFunc(data *Data) *Function {
	if this.f != nil {
		return this.f
	}

	for _, fu := range data.funcs {
		if this.funcName == fu.name {
			this.f = fu
			return fu
		}
	}

	return nil
}

// ================================================================
// MARK: InstructionChangeInterface
// ================================================================

type InstructionChangeInterface struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionChangeType
// ================================================================

type InstructionChangeType struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionConst
// ================================================================

type InstructionConst struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionConvert
// ================================================================

type InstructionConvert struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionDebugRef
// ================================================================

type InstructionDebugRef struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionDefer
// ================================================================

type InstructionDefer struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionExtract
// ================================================================

type InstructionExtract struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionField
// ================================================================

type InstructionField struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionFieldAddr
// ================================================================

type InstructionFieldAddr struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionFreeVar
// ================================================================

type InstructionFreeVar struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionFunction
// ================================================================

type InstructionFunction struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionGo
// ================================================================

type InstructionGo struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionIf
// ================================================================

type InstructionIf struct {
	InstructionBase

	case_if   int
	case_else int
}

func newInstructionIf(f *Function, inst *ssa.If, i int) *InstructionIf {
	ind := inst.Block().Succs[0].Index
	e := inst.Block().Succs[1].Index
	return &InstructionIf{InstructionBase: newInstructionBase(f, Ic_jump, inst, i), case_if: ind, case_else: e}
}

func (this *InstructionIf) If() int {
	return this.case_if
}

func (this *InstructionIf) Else() int {
	return this.case_else
}

// ================================================================
// MARK: InstructionIndex
// ================================================================

type InstructionIndex struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionIndexAddr
// ================================================================

type InstructionIndexAddr struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionJump
// ================================================================

type InstructionJump struct {
	InstructionBase

	to int
}

func newInstructionJump(f *Function, inst *ssa.Jump, i int) *InstructionJump {
	return &InstructionJump{InstructionBase: newInstructionBase(f, Ic_jump, inst, i), to: inst.Block().Succs[0].Index}
}

func (this *InstructionJump) To() int {
	return this.to
}

// ================================================================
// MARK: InstructionLookup
// ================================================================

type InstructionLookup struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeChan
// ================================================================

type InstructionMakeChan struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeClosure
// ================================================================

type InstructionMakeClosure struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeInterface
// ================================================================

type InstructionMakeInterface struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeMap
// ================================================================

type InstructionMakeMap struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMakeSlice
// ================================================================

type InstructionMakeSlice struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMapUpdate
// ================================================================

type InstructionMapUpdate struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionMultiConvert
// ================================================================

type InstructionMultiConvert struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionNext
// ================================================================

type InstructionNext struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionPanic
// ================================================================

type InstructionPanic struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionParameter
// ================================================================

type InstructionParameter struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionPhi
// ================================================================

type InstructionPhi struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionRange
// ================================================================

type InstructionRange struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionReturn
// ================================================================

type InstructionReturn struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionRunDefers
// ================================================================

type InstructionRunDefers struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSelect
// ================================================================

type InstructionSelect struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSend
// ================================================================

type InstructionSend struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSlice
// ================================================================

type InstructionSlice struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionSliceToArrayPointer
// ================================================================

type InstructionSliceToArrayPointer struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionStore
// ================================================================

type InstructionStore struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionTypeAssert
// ================================================================

type InstructionTypeAssert struct {
	InstructionBase
}

// ================================================================
// MARK: InstructionUnOp
// ================================================================

type InstructionUnOp struct {
	InstructionBase
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

func (this hasConcInfo) Channel() bool {
	return this[chanInd]
}

func (this hasConcInfo) Mutex() bool {
	return this[mutexInd]
}

func (this hasConcInfo) CondVar() bool {
	return this[condVarInd]
}

func (this hasConcInfo) WaitGroup() bool {
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
