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

	setVariable(name string)
	setTerm(term string)
	setRelevant(rel, trace bool)
	setInst(inst ssa.Instruction)
	setConc(conc hasConcInfo)
}

// ================================================================
// MARK: Base
// ================================================================

type InstructionBase struct {
	variable string
	term     string
	inst     ssa.Instruction

	varPtr bool

	inTrace bool
	relvant bool

	class InstClass

	conc hasConcInfo
}

func newInstructionBase(c InstClass, inst ssa.Instruction) InstructionBase {
	name := ""
	switch v := inst.(type) {
	case ssa.Value:
		name = v.Name()
	case nil:
		// Be robust against bad transforms.
		name = "<deleted>"
	}

	// global var assign
	term := inst.String()
	if name == "" && strings.Contains(term, " = ") {
		fields := strings.Split(term, " = ")
		name = fields[0]
		term = fields[1]
	}

	b := InstructionBase{class: c}

	b.setVariable(name)
	b.setTerm(term)
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

func (this *InstructionBase) setVariable(name string) {
	variable := strings.TrimPrefix(name, "*")
	this.variable = variable
	this.varPtr = (name != variable)
}

func (this *InstructionBase) setTerm(term string) {
	this.term = term
}

func (this *InstructionBase) setInst(inst ssa.Instruction) {
	this.inst = inst
}

func (this *InstructionBase) setConc(conc hasConcInfo) {
	this.conc = conc
}

func (this *Data) analysisInstruction(instr ssa.Instruction) Instruction {

	var inst Instruction

	switch instr := instr.(type) {
	case *ssa.Alloc:
		inst = &InstructionAlloc{InstructionBase: newInstructionBase(Ic_alloc, instr)}
	case *ssa.BinOp:
		inst = &InstructionBinOp{InstructionBase: newInstructionBase(Ic_binOp, instr)}
	case *ssa.Call:
		inst = NewInstructionCall(instr)
	case *ssa.ChangeInterface:
		inst = &InstructionChangeInterface{InstructionBase: newInstructionBase(Ic_changeInterface, instr)}
	case *ssa.ChangeType:
		inst = &InstructionChangeType{InstructionBase: newInstructionBase(Ic_changeType, instr)}
	case *ssa.Convert:
		inst = &InstructionConvert{InstructionBase: newInstructionBase(Ic_convert, instr)}
	case *ssa.DebugRef:
		inst = &InstructionDebugRef{InstructionBase: newInstructionBase(Ic_debugRef, instr)}
	case *ssa.Defer:
		inst = &InstructionDefer{InstructionBase: newInstructionBase(Ic_defer, instr)}
	case *ssa.Extract:
		inst = &InstructionExtract{InstructionBase: newInstructionBase(Ic_extract, instr)}
	case *ssa.Field:
		inst = &InstructionField{InstructionBase: newInstructionBase(Ic_field, instr)}
	case *ssa.FieldAddr:
		inst = &InstructionFieldAddr{InstructionBase: newInstructionBase(Ic_fieldAddr, instr)}
	case *ssa.Go:
		inst = &InstructionGo{InstructionBase: newInstructionBase(Ic_go, instr)}
	case *ssa.If:
		inst = newInstructionIf(instr)
	case *ssa.Index:
		inst = &InstructionIndex{InstructionBase: newInstructionBase(Ic_index, instr)}
	case *ssa.IndexAddr:
		inst = &InstructionIndexAddr{InstructionBase: newInstructionBase(Ic_indexAddr, instr)}
	case *ssa.Jump:
		inst = newInstructionJump(instr)
	case *ssa.Lookup:
		inst = &InstructionLookup{InstructionBase: newInstructionBase(Ic_lookup, instr)}
	case *ssa.MakeChan:
		inst = &InstructionMakeChan{InstructionBase: newInstructionBase(Ic_makeChan, instr)}
	case *ssa.MakeClosure:
		inst = &InstructionMakeClosure{InstructionBase: newInstructionBase(Ic_makeClosure, instr)}
	case *ssa.MakeInterface:
		inst = &InstructionMakeInterface{InstructionBase: newInstructionBase(Ic_makeInterface, instr)}
	case *ssa.MakeMap:
		inst = &InstructionMakeMap{InstructionBase: newInstructionBase(Ic_makeMap, instr)}
	case *ssa.MakeSlice:
		inst = &InstructionMakeSlice{InstructionBase: newInstructionBase(Ic_makeSlice, instr)}
	case *ssa.MapUpdate:
		inst = &InstructionMapUpdate{InstructionBase: newInstructionBase(Ic_mapUpdate, instr)}
	case *ssa.MultiConvert:
		inst = &InstructionMultiConvert{InstructionBase: newInstructionBase(Ic_multiConvert, instr)}
	case *ssa.Next:
		inst = &InstructionNext{InstructionBase: newInstructionBase(Ic_next, instr)}
	case *ssa.Panic:
		inst = &InstructionPanic{InstructionBase: newInstructionBase(Ic_panic, instr)}
	case *ssa.Phi:
		inst = &InstructionPhi{InstructionBase: newInstructionBase(Ic_phi, instr)}
	case *ssa.Range:
		inst = &InstructionRange{InstructionBase: newInstructionBase(Ic_range, instr)}
	case *ssa.Return:
		inst = &InstructionReturn{InstructionBase: newInstructionBase(Ic_return, instr)}
	case *ssa.RunDefers:
		inst = &InstructionRunDefers{InstructionBase: newInstructionBase(Ic_runDefers, instr)}
	case *ssa.Select:
		inst = &InstructionSelect{InstructionBase: newInstructionBase(Ic_select, instr)}
	case *ssa.Send:
		inst = &InstructionSend{InstructionBase: newInstructionBase(Ic_send, instr)}
	case *ssa.Slice:
		inst = &InstructionSlice{InstructionBase: newInstructionBase(Ic_slice, instr)}
	case *ssa.SliceToArrayPointer:
		inst = &InstructionSliceToArrayPointer{InstructionBase: newInstructionBase(Ic_sliceToArrayPointer, instr)}
	case *ssa.Store:
		inst = &InstructionStore{InstructionBase: newInstructionBase(Ic_store, instr)}
	case *ssa.TypeAssert:
		inst = &InstructionTypeAssert{InstructionBase: newInstructionBase(Ic_typeAssert, instr)}
	case *ssa.UnOp:
		inst = &InstructionUnOp{InstructionBase: newInstructionBase(Ic_unOp, instr)}
	default:
		log.Error("Found unknown instruction: ", instr.String())
		inst = &InstructionUnknown{InstructionBase: newInstructionBase(Ic_unknown, instr)}
	}

	setContainsPrimitive(inst)

	this.isRelvant(inst)

	return inst
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

func NewInstructionCall(inst *ssa.Call) *InstructionCall {
	name := ""
	if callee := inst.Common().StaticCallee(); callee != nil {
		name = callee.String()
	}

	return &InstructionCall{InstructionBase: newInstructionBase(Ic_call, inst),
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

func newInstructionIf(inst *ssa.If) *InstructionIf {
	i := inst.Block().Succs[0].Index
	e := inst.Block().Succs[1].Index
	return &InstructionIf{InstructionBase: newInstructionBase(Ic_jump, inst), case_if: i, case_else: e}
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

func newInstructionJump(inst *ssa.Jump) *InstructionJump {
	return &InstructionJump{InstructionBase: newInstructionBase(Ic_jump, inst), to: inst.Block().Succs[0].Index}
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
